#!/usr/bin/env bash
set -e

REG="crpi-cf5tmc2njsrq2eko.cn-hangzhou.personal.cr.aliyuncs.com"
NS="jiangqiao1/govideo"
TAG="frontend-$(date +%Y%m%d-%H%M%S)"      # 可以在命令前指定 TAG=xxx

echo "Working dir: $(pwd)"
echo "Using tag: ${TAG}"

cleanup_local_images() {
  # 只清理本项目在本地构建的镜像，避免占满磁盘
  echo ""
  echo "Cleaning local Docker images for ${REG}/${NS}..."
  local imgs
  imgs=$(docker images "${REG}/${NS}" -q || true)
  if [ -z "${imgs}" ]; then
    echo "No local images to clean up for ${REG}/${NS}."
    return 0
  fi

  # 将所有镜像 ID 展开成多个参数传给 docker rmi
  # shellcheck disable=SC2086
  docker rmi ${imgs} || true
}

echo "Building images..."
docker build --pull -f user-service/Dockerfile \
  --build-arg CONFIG_PATH=/app/configs/config_prod.yaml \
  -t ${REG}/${NS}:user-service${TAG:+-$TAG} .
docker build --pull -f upload-service/Dockerfile \
  --build-arg CONFIG_PATH=/app/configs/config_prod.yaml \
  -t ${REG}/${NS}:upload-service${TAG:+-$TAG} .
docker build --pull -f transcode-service/Dockerfile \
  --build-arg CONFIG_PATH=/app/configs/config_prod.yaml \
  -t ${REG}/${NS}:transcode-service${TAG:+-$TAG} .
docker build --pull -f video-service/Dockerfile \
  --build-arg CONFIG_PATH=/app/configs/config_prod.yaml \
  -t ${REG}/${NS}:video-service${TAG:+-$TAG} .
docker build --pull -f notification-service/Dockerfile \
  --build-arg CONFIG_PATH=/app/configs/config_prod.yaml \
  -t ${REG}/${NS}:notification-service${TAG:+-$TAG} .
docker build --pull -f gateway-service/Dockerfile \
  -t ${REG}/${NS}:kong-custom${TAG:+-$TAG} .
docker build --pull -f frontend/Dockerfile          -t ${REG}/${NS}:frontend${TAG:+-$TAG} .

echo "Pushing images to ACR..."
docker push ${REG}/${NS}:user-service${TAG:+-$TAG}
docker push ${REG}/${NS}:upload-service${TAG:+-$TAG}
docker push ${REG}/${NS}:transcode-service${TAG:+-$TAG}
docker push ${REG}/${NS}:video-service${TAG:+-$TAG}
docker push ${REG}/${NS}:notification-service${TAG:+-$TAG}
docker push ${REG}/${NS}:kong-custom${TAG:+-$TAG}
docker push ${REG}/${NS}:frontend${TAG:+-$TAG}

UPLOAD_IMG="${REG}/${NS}:upload-service${TAG:+-$TAG}"
TRANSCODE_IMG="${REG}/${NS}:transcode-service${TAG:+-$TAG}"
USER_IMG="${REG}/${NS}:user-service${TAG:+-$TAG}"
VIDEO_IMG="${REG}/${NS}:video-service${TAG:+-$TAG}"
NOTIFICATION_IMG="${REG}/${NS}:notification-service${TAG:+-$TAG}"
KONG_CUSTOM_IMG="${REG}/${NS}:kong-custom${TAG:+-$TAG}"
FRONTEND_IMG="${REG}/${NS}:frontend${TAG:+-$TAG}"

echo "Done."
echo ""
echo "Images:"
echo "upload-service:    ${UPLOAD_IMG}"
echo "transcode-service: ${TRANSCODE_IMG}"
echo "user-service:      ${USER_IMG}"
echo "video-service:     ${VIDEO_IMG}"
echo "notification:      ${NOTIFICATION_IMG}"
echo "kong-custom:       ${KONG_CUSTOM_IMG}"
echo "frontend:          ${FRONTEND_IMG}"
echo ""
echo "Apply to k3s (replace if your namespace/name differs):"
echo "k3s kubectl -n go-video set image deployment/upload-service upload-service=${UPLOAD_IMG}"
echo "k3s kubectl -n go-video set image deployment/transcode-service transcode-service=${TRANSCODE_IMG}"
echo "k3s kubectl -n go-video set image deployment/user-service user-service=${USER_IMG}"
echo "k3s kubectl -n go-video set image deployment/video-service video-service=${VIDEO_IMG}"
echo "k3s kubectl -n go-video set image deployment/notification-service notification-service=${NOTIFICATION_IMG}"
echo "k3s kubectl -n go-video set image deployment/frontend frontend=${FRONTEND_IMG}"
echo "k3s kubectl -n go-video rollout restart deployment/upload-service deployment/transcode-service deployment/user-service deployment/video-service deployment/notification-service deployment/frontend"
echo "k3s kubectl -n go-video rollout status deployment/upload-service deployment/transcode-service deployment/user-service deployment/video-service deployment/notification-service deployment/frontend"

cleanup_local_images
