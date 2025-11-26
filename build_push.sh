#!/usr/bin/env bash
set -e

REG="crpi-cf5tmc2njsrq2eko.cn-hangzhou.personal.cr.aliyuncs.com"
NS="jiangqiao1/govideo"
TAG="frontend-$(date +%Y%m%d-%H%M%S)"      # 可以在命令前指定 TAG=xxx
GATEWAY_ENV="${GATEWAY_ENV:-prod}"        # dev|prod，自动选 .env.<ENV>

echo "Working dir: $(pwd)"
echo "Using tag: ${TAG}"

echo "Preparing gateway config..."
pushd gateway-service >/dev/null
ENV_FILE=".env.${GATEWAY_ENV}" ./gen-kong.sh
popd >/dev/null

echo "Building images..."
docker build --pull -f user-service/Dockerfile      -t ${REG}/${NS}:user-service${TAG:+-$TAG} .
docker build --pull -f upload-service/Dockerfile    -t ${REG}/${NS}:upload-service${TAG:+-$TAG} .
docker build --pull -f transcode-service/Dockerfile -t ${REG}/${NS}:transcode-service${TAG:+-$TAG} .
docker build --pull -f frontend/Dockerfile          -t ${REG}/${NS}:frontend${TAG:+-$TAG} .
docker build --pull -f gateway-service/Dockerfile   -t ${REG}/${NS}:gateway${TAG:+-$TAG} .

echo "Pushing images to ACR..."
docker push ${REG}/${NS}:user-service${TAG:+-$TAG}
docker push ${REG}/${NS}:upload-service${TAG:+-$TAG}
docker push ${REG}/${NS}:transcode-service${TAG:+-$TAG}
docker push ${REG}/${NS}:frontend${TAG:+-$TAG}
docker push ${REG}/${NS}:gateway${TAG:+-$TAG}

UPLOAD_IMG="${REG}/${NS}:upload-service${TAG:+-$TAG}"
TRANSCODE_IMG="${REG}/${NS}:transcode-service${TAG:+-$TAG}"
USER_IMG="${REG}/${NS}:user-service${TAG:+-$TAG}"
GATEWAY_IMG="${REG}/${NS}:gateway${TAG:+-$TAG}"
FRONTEND_IMG="${REG}/${NS}:frontend${TAG:+-$TAG}"

echo "Done."
echo ""
echo "Images:"
echo "upload-service:    ${UPLOAD_IMG}"
echo "transcode-service: ${TRANSCODE_IMG}"
echo "user-service:      ${USER_IMG}"
echo "gateway:           ${GATEWAY_IMG}"
echo "frontend:          ${FRONTEND_IMG}"
echo ""
echo "Apply to k3s (replace if your namespace/name differs):"
echo "k3s kubectl -n go-video set image deployment/upload-service upload-service=${UPLOAD_IMG}"
echo "k3s kubectl -n go-video set image deployment/transcode-service transcode-service=${TRANSCODE_IMG}"
echo "k3s kubectl -n go-video set image deployment/user-service user-service=${USER_IMG}"
echo "k3s kubectl -n go-video set image deployment/gateway kong=${GATEWAY_IMG}"
echo "k3s kubectl -n go-video set image deployment/frontend frontend=${FRONTEND_IMG}"
echo "k3s kubectl -n go-video rollout restart deployment/upload-service deployment/transcode-service deployment/user-service deployment/gateway deployment/frontend"
echo "k3s kubectl -n go-video rollout status deployment/upload-service deployment/transcode-service deployment/user-service deployment/gateway deployment/frontend"
