import React, { useState, useRef } from 'react';
import { 
  Upload, 
  Button, 
  Progress, 
  Card, 
  message, 
  Typography, 
  Space, 
  Alert,
  List,
  Tag
} from 'antd';
import { 
  UploadOutlined, 
  PlayCircleOutlined, 
  PauseCircleOutlined,
  DeleteOutlined,
  CheckCircleOutlined
} from '@ant-design/icons';
import { useAuthStore } from '@/store/auth';
import apiService from '@/services/api';
import { calculateFileHash, calculateChunkHash, generateUUID } from '@/utils/crypto';
import { UploadVideoInfo, UploadChunkInfo } from '@/types/api';

const { Title, Text } = Typography;
const { Dragger } = Upload;

interface UploadTask {
  id: string;
  file: File;
  uploadInfo?: UploadVideoInfo;
  progress: number;
  status: 'waiting' | 'uploading' | 'paused' | 'completed' | 'error';
  error?: string;
  currentChunk: number;
}

const CHUNK_SIZE = 1024 * 1024 * 2; // 2MB per chunk

const VideoUpload: React.FC = () => {
  const [uploadTasks, setUploadTasks] = useState<UploadTask[]>([]);
  const { user } = useAuthStore();
  const abortControllersRef = useRef<Map<string, AbortController>>(new Map());

  // 添加文件到上传队列
  const handleFileSelect = async (file: File) => {
    if (!user) {
      message.error('请先登录');
      return false;
    }

    // 检查文件类型
    const allowedTypes = ['video/mp4', 'video/avi', 'video/mov', 'video/wmv', 'video/flv'];
    if (!allowedTypes.includes(file.type)) {
      message.error('只支持视频文件格式（MP4, AVI, MOV, WMV, FLV）');
      return false;
    }

    // 检查文件大小（最大5GB）
    const maxSize = 5 * 1024 * 1024 * 1024;
    if (file.size > maxSize) {
      message.error('文件大小不能超过5GB');
      return false;
    }

    const taskId = generateUUID();
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    const newTask: UploadTask = {
      id: taskId,
      file,
      progress: 0,
      status: 'waiting',
      currentChunk: 0,
    };

    setUploadTasks(prev => [...prev, newTask]);

    // 开始上传
    startUpload(taskId, file, totalChunks);

    return false; // 阻止默认上传行为
  };

  // 开始上传
  const startUpload = async (taskId: string, file: File, totalChunks: number) => {
    try {
      setUploadTasks(prev => prev.map(task => 
        task.id === taskId ? { ...task, status: 'uploading' } : task
      ));

      // 计算文件哈希
      const fileHash = await calculateFileHash(file);

      // 初始化上传
      const uploadInfo = await apiService.initVideoUpload({
        file_name: file.name,
        file_size: file.size,
        total_chunks: totalChunks,
        user_uuid: user!.user_uuid,
        file_hash: fileHash,
      });

      setUploadTasks(prev => prev.map(task => 
        task.id === taskId ? { ...task, uploadInfo } : task
      ));

      // 检查已上传的分片
      const uploadedChunks = uploadInfo.chunks.filter(chunk => chunk.status === 'completed');
      const startChunk = uploadedChunks.length;

      if (startChunk === totalChunks) {
        // 文件已完全上传，直接合并
        await mergeChunks(taskId, uploadInfo.upload_video_uuid);
        return;
      }

      // 创建AbortController用于取消上传
      const abortController = new AbortController();
      abortControllersRef.current.set(taskId, abortController);

      // 上传剩余分片
      await uploadChunks(taskId, file, uploadInfo, startChunk, abortController.signal);

    } catch (error: any) {
      console.error('Upload failed:', error);
      setUploadTasks(prev => prev.map(task => 
        task.id === taskId ? { 
          ...task, 
          status: 'error', 
          error: error.message || '上传失败'
        } : task
      ));
      message.error('上传失败：' + (error.message || '未知错误'));
    }
  };

  // 上传分片
  const uploadChunks = async (
    taskId: string, 
    file: File, 
    uploadInfo: UploadVideoInfo, 
    startChunk: number,
    signal: AbortSignal
  ) => {
    const totalChunks = uploadInfo.total_chunks;

    for (let i = startChunk; i < totalChunks; i++) {
      if (signal.aborted) {
        throw new Error('上传已取消');
      }

      const start = i * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, end);
      const chunkArrayBuffer = await chunk.arrayBuffer();
      const chunkHash = await calculateChunkHash(chunkArrayBuffer);

      const chunkInfo = uploadInfo.chunks[i];

      try {
        await apiService.uploadChunk({
          chunk_uuid: chunkInfo.chunk_uuid,
          user_uuid: user!.user_uuid,
          upload_video_uuid: uploadInfo.upload_video_uuid,
          chunk_size: chunk.size,
          chunk_index: i,
          chunk_data: chunkArrayBuffer,
          chunk_hash: chunkHash,
        });

        // 更新进度
        const progress = Math.round(((i + 1) / totalChunks) * 100);
        setUploadTasks(prev => prev.map(task => 
          task.id === taskId ? { 
            ...task, 
            progress, 
            currentChunk: i + 1 
          } : task
        ));

      } catch (error: any) {
        if (signal.aborted) {
          throw new Error('上传已取消');
        }
        throw new Error(`分片 ${i + 1} 上传失败: ${error.message}`);
      }
    }

    // 所有分片上传完成，合并文件
    await mergeChunks(taskId, uploadInfo.upload_video_uuid);
  };

  // 合并分片
  const mergeChunks = async (taskId: string, uploadVideoUuid: string) => {
    try {
      await apiService.mergeChunks({
        upload_video_uuid: uploadVideoUuid,
        user_uuid: user!.user_uuid,
      });

      setUploadTasks(prev => prev.map(task => 
        task.id === taskId ? { 
          ...task, 
          status: 'completed', 
          progress: 100 
        } : task
      ));

      message.success('视频上传完成！');
    } catch (error: any) {
      throw new Error(`合并文件失败: ${error.message}`);
    }
  };

  // 暂停上传
  const pauseUpload = (taskId: string) => {
    const abortController = abortControllersRef.current.get(taskId);
    if (abortController) {
      abortController.abort();
      abortControllersRef.current.delete(taskId);
    }

    setUploadTasks(prev => prev.map(task => 
      task.id === taskId ? { ...task, status: 'paused' } : task
    ));
  };

  // 恢复上传
  const resumeUpload = (taskId: string) => {
    const task = uploadTasks.find(t => t.id === taskId);
    if (task && task.uploadInfo) {
      const totalChunks = Math.ceil(task.file.size / CHUNK_SIZE);
      startUpload(taskId, task.file, totalChunks);
    }
  };

  // 删除任务
  const removeTask = (taskId: string) => {
    const abortController = abortControllersRef.current.get(taskId);
    if (abortController) {
      abortController.abort();
      abortControllersRef.current.delete(taskId);
    }

    setUploadTasks(prev => prev.filter(task => task.id !== taskId));
  };

  // 获取状态颜色
  const getStatusColor = (status: UploadTask['status']) => {
    switch (status) {
      case 'waiting': return 'default';
      case 'uploading': return 'processing';
      case 'paused': return 'warning';
      case 'completed': return 'success';
      case 'error': return 'error';
      default: return 'default';
    }
  };

  // 获取状态文本
  const getStatusText = (status: UploadTask['status']) => {
    switch (status) {
      case 'waiting': return '等待中';
      case 'uploading': return '上传中';
      case 'paused': return '已暂停';
      case 'completed': return '已完成';
      case 'error': return '上传失败';
      default: return '未知状态';
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Title level={3}>视频上传</Title>
        
        <Dragger
          name="file"
          multiple={false}
          beforeUpload={handleFileSelect}
          showUploadList={false}
          style={{ marginBottom: 24 }}
        >
          <p className="ant-upload-drag-icon">
            <UploadOutlined style={{ fontSize: 48, color: '#1890ff' }} />
          </p>
          <p className="ant-upload-text">点击或拖拽视频文件到此区域上传</p>
          <p className="ant-upload-hint">
            支持 MP4, AVI, MOV, WMV, FLV 格式，单个文件最大 5GB
          </p>
        </Dragger>

        {uploadTasks.length > 0 && (
          <div>
            <Title level={4}>上传任务</Title>
            <List
              dataSource={uploadTasks}
              renderItem={(task) => (
                <List.Item
                  actions={[
                    task.status === 'uploading' && (
                      <Button 
                        icon={<PauseCircleOutlined />} 
                        onClick={() => pauseUpload(task.id)}
                        size="small"
                      >
                        暂停
                      </Button>
                    ),
                    task.status === 'paused' && (
                      <Button 
                        icon={<PlayCircleOutlined />} 
                        onClick={() => resumeUpload(task.id)}
                        size="small"
                        type="primary"
                      >
                        继续
                      </Button>
                    ),
                    task.status !== 'uploading' && (
                      <Button 
                        icon={<DeleteOutlined />} 
                        onClick={() => removeTask(task.id)}
                        size="small"
                        danger
                      >
                        删除
                      </Button>
                    ),
                  ].filter(Boolean)}
                >
                  <List.Item.Meta
                    avatar={
                      task.status === 'completed' ? (
                        <CheckCircleOutlined style={{ fontSize: 24, color: '#52c41a' }} />
                      ) : (
                        <UploadOutlined style={{ fontSize: 24, color: '#1890ff' }} />
                      )
                    }
                    title={
                      <Space>
                        <Text strong>{task.file.name}</Text>
                        <Tag color={getStatusColor(task.status)}>
                          {getStatusText(task.status)}
                        </Tag>
                      </Space>
                    }
                    description={
                      <div>
                        <Text type="secondary">
                          大小: {(task.file.size / 1024 / 1024).toFixed(2)} MB
                        </Text>
                        {task.uploadInfo && (
                          <Text type="secondary" style={{ marginLeft: 16 }}>
                            分片: {task.currentChunk}/{task.uploadInfo.total_chunks}
                          </Text>
                        )}
                        <div style={{ marginTop: 8 }}>
                          <Progress 
                            percent={task.progress} 
                            status={task.status === 'error' ? 'exception' : 'active'}
                            size="small"
                          />
                        </div>
                        {task.error && (
                          <Alert 
                            message={task.error} 
                            type="error" 
                            style={{ marginTop: 8 }}
                          />
                        )}
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </div>
        )}
      </Card>
    </div>
  );
};

export default VideoUpload;