#!/bin/bash

# 视频上传压力测试脚本
# 使用curl进行真实的multipart/form-data文件上传测试

echo "🚀 开始视频上传压力测试"
echo "================================"

# 配置参数
BASE_URL="http://localhost:8080/api/v2/videos/upload"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX3V1aWQiOiJmMGM0NWUwMC0xMzk3LTQ1OTctYmJmYy1lMTZlNDRlOTkxMTIiLCJ1c2VyX2lkIjozLCJleHAiOjE3NTYwMTg0MjgsIm5iZiI6MTc1NTkzMjAyOCwiaWF0IjoxNzU1OTMyMDI4fQ._W8EXAy2LyRaM-jpMFPtxRiQWrkfxlvf5P8Q7POI3cI"
TEST_FILE="videoMp3/video1.mp4"

# 检查文件是否存在
if [ ! -f "$TEST_FILE" ]; then
    echo "❌ 测试文件不存在: $TEST_FILE"
    exit 1
fi

# 检查服务是否运行
echo "📡 检查服务状态..."
curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/health" > /dev/null
if [ $? -ne 0 ]; then
    echo "❌ 服务未运行，请先启动服务"
    exit 1
fi

echo "✅ 服务运行正常"
echo ""

# 创建临时目录
mkdir -p temp_test

# 单个请求测试函数
test_single_upload() {
    local id=$1
    local start_time=$(date +%s.%N)
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $TOKEN" \
        -F "title=压力测试视频_$id" \
        -F "description=并发测试视频_$id" \
        -F "format=mp4" \
        -F "file=@$TEST_FILE" \
        "$BASE_URL" 2>/dev/null)
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc -l)
    local http_code=$(echo "$response" | tail -n1)
    local response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        echo "✅ 请求 $id: 成功 (${duration}s)"
        return 0
    else
        echo "❌ 请求 $id: 失败 HTTP $http_code (${duration}s)"
        echo "   响应: $response_body"
        return 1
    fi
}

# 并发测试函数
test_concurrent_upload() {
    local concurrent=$1
    local description=$2
    
    echo "📊 测试场景: $description"
    echo "   并发数: $concurrent"
    echo "   开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "   ----------------------------------------"
    
    local start_time=$(date +%s.%N)
    local success=0
    local failed=0
    
    # 启动并发进程
    for ((i=1; i<=concurrent; i++)); do
        {
            if test_single_upload $i; then
                echo "SUCCESS" > temp_test/result_$i
            else
                echo "FAILED" > temp_test/result_$i
            fi
        } &
    done
    
    # 等待所有进程完成
    wait
    
    local end_time=$(date +%s.%N)
    local total_time=$(echo "$end_time - $start_time" | bc -l)
    
    # 统计结果
    for ((i=1; i<=concurrent; i++)); do
        if [ -f "temp_test/result_$i" ]; then
            if grep -q "SUCCESS" "temp_test/result_$i"; then
                ((success++))
            else
                ((failed++))
            fi
            rm -f "temp_test/result_$i"
        fi
    done
    
    local success_rate=$(echo "scale=2; $success * 100 / $concurrent" | bc -l)
    local qps=$(echo "scale=2; $concurrent / $total_time" | bc -l)
    
    echo "   总耗时: ${total_time}s"
    echo "   成功: $success, 失败: $failed"
    echo "   成功率: ${success_rate}%"
    echo "   QPS: $qps"
    echo "   完成时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "   ----------------------------------------"
    
    # 等待系统恢复
    echo "   ⏳ 等待系统恢复 (5秒)..."
    sleep 5
}

# 测试场景配置
declare -a SCENARIOS=(
    "5:轻量级测试"      # 5个并发
    "10:中等负载"       # 10个并发
    "20:高负载测试"     # 20个并发
    "50:极限测试"       # 50个并发
)

echo "🎯 开始执行测试场景"
echo "================================"

# 执行测试
for scenario in "${SCENARIOS[@]}"; do
    IFS=':' read -r concurrent description <<< "$scenario"
    echo ""
    test_concurrent_upload $concurrent "$description"
done

echo ""
echo "🎉 压力测试完成！"
echo "================================"

# 清理临时文件
rm -rf temp_test

echo "📋 测试总结:"
echo "1. 检查服务日志查看详细错误信息"
echo "2. 监控系统资源使用情况 (CPU, 内存, 数据库连接)"
echo "3. 查看MinIO存储状态"
echo "4. 分析数据库中的记录数量"
echo ""
echo "💡 建议的监控命令:"
echo "   - 查看进程: ps aux | grep go-video"
echo "   - 查看内存: free -h"
echo "   - 查看数据库连接: docker exec mysql mysql -u root -p123456 -e 'SHOW PROCESSLIST;'"
echo "   - 查看MinIO: curl http://localhost:9001"
echo "   - 查看上传记录: docker exec mysql mysql -u root -p123456 -e 'SELECT COUNT(*) FROM go_video.video; SELECT COUNT(*) FROM go_video.video_upload_task;'"