#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=${NAMESPACE:-go-video}
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-jiangqiao}
RUSTFS_ACCESS=${RUSTFS_ACCESS:-jiangqiao}
RUSTFS_SECRET=${RUSTFS_SECRET:-jiangqiao}
KAFKA_CLIENT_PORT=${KAFKA_CLIENT_PORT:-19092}

echo "namespace: $NAMESPACE"

if ! command -v helm >/dev/null 2>&1; then
  curl -fsSL https://ghproxy.com/https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi

kubectl get ns "$NAMESPACE" >/dev/null 2>&1 || kubectl create ns "$NAMESPACE"

helm repo add bitnami https://charts.bitnami.com/bitnami >/dev/null 2>&1 || true
helm repo update

helm upgrade --install mysql bitnami/mysql -n "$NAMESPACE" \
  --set auth.rootPassword="$MYSQL_ROOT_PASSWORD"

helm upgrade --install redis bitnami/redis -n "$NAMESPACE" \
  --set architecture=standalone \
  --set auth.enabled=false

helm upgrade --install kafka bitnami/kafka -n "$NAMESPACE" \
  --set kraft.enabled=true \
  --set controller.replicaCount=1 \
  --set broker.replicaCount=1

helm upgrade --install rustfs bitnami/minio -n "$NAMESPACE" \
  --set auth.rootUser="$RUSTFS_ACCESS" \
  --set auth.rootPassword="$RUSTFS_SECRET" \
  --set defaultBuckets="upload,transcode"

kubectl -n "$NAMESPACE" wait --for=condition=Ready pod -l app.kubernetes.io/name=mysql --timeout=300s || true
kubectl -n "$NAMESPACE" wait --for=condition=Ready pod -l app.kubernetes.io/name=redis --timeout=300s || true
kubectl -n "$NAMESPACE" wait --for=condition=Ready pod -l app.kubernetes.io/name=kafka --timeout=600s || true
kubectl -n "$NAMESPACE" wait --for=condition=Ready pod -l app.kubernetes.io/name=minio --timeout=300s || true

kubectl -n "$NAMESPACE" create secret generic mysql-root-pass \
  --from-literal=MYSQL_ROOT_PASSWORD="$MYSQL_ROOT_PASSWORD" \
  --dry-run=client -o yaml | kubectl apply -f -

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
kubectl -n "$NAMESPACE" create configmap mysql-init-sql \
  --from-file="$ROOT_DIR/scripts/mysql/init_all.sql" \
  --dry-run=client -o yaml | kubectl apply -f -

cat <<'EOF' | kubectl -n "$NAMESPACE" apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: mysql-init
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: mysql-init
          image: mysql:8.0
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mysql-root-pass
                  key: MYSQL_ROOT_PASSWORD
          command: ["sh","-c","mysql -h mysql -P 3306 -u root -p$MYSQL_ROOT_PASSWORD < /scripts/init_all.sql"]
          volumeMounts:
            - name: sql
              mountPath: /scripts
      volumes:
        - name: sql
          configMap:
            name: mysql-init-sql
EOF

# 确保服务名称与应用配置中一致（创建别名Service）
create_or_apply_service() {
  local name="$1" selector="$2" port="$3" target="$4"
  if ! kubectl -n "$NAMESPACE" get svc "$name" >/dev/null 2>&1; then
    cat <<EOF | kubectl -n "$NAMESPACE" apply -f -
apiVersion: v1
kind: Service
metadata:
  name: ${name}
spec:
  selector:
    app.kubernetes.io/name: ${selector}
  ports:
  - name: tcp-${port}
    port: ${port}
    targetPort: ${target}
EOF
  fi
}

create_or_apply_service mysql mysql 3306 3306
create_or_apply_service redis redis 6379 6379
create_or_apply_service rustfs minio 9000 9000

kubectl -n "$NAMESPACE" patch svc kafka -p "{\"spec\":{\"ports\":[{\"name\":\"client-19092\",\"port\":$KAFKA_CLIENT_PORT,\"targetPort\":9092}]}}" --type merge || true

echo "done"
