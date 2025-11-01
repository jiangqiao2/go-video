#!/bin/sh

# 确保目录存在并设置正确权限
mkdir -p /tmp/transcode/transcoded
chown -R appuser:appgroup /tmp/transcode

# 切换到appuser并启动应用
exec su-exec appuser ./transcode-service