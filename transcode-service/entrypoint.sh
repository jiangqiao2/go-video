#!/bin/sh

# 确保目录存在
mkdir -p /tmp/transcode/transcoded

# 以当前用户启动应用（镜像中已切到 appuser）
exec ./transcode-service
