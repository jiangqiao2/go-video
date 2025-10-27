import React, { useRef, useState } from 'react';
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
  Tag,
} from 'antd';
import {
  UploadOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/store/auth';
import apiService from '@/services/api';
import { calculateFileHash, calculateChunkHash, generateUUID } from '@/utils/crypto';
import { UploadVideoInfo } from '@/types/api';

const { Title, Text } = Typography;
const { Dragger } = Upload;

const CHUNK_SIZE = 1024 * 1024 * 2; // 2MB per chunk

type UploadStatus = 'waiting' | 'uploading' | 'paused' | 'completed' | 'error';

interface UploadTask {
  id: string;
  file: File;
  uploadInfo?: UploadVideoInfo;
  totalChunks: number;
  uploadedChunks: number[];
  progress: number;
  status: UploadStatus;
  error?: string;
  currentChunk: number;
}

const VideoUpload: React.FC = () => {
  const [uploadTasks, setUploadTasks] = useState<UploadTask[]>([]);
  const { user } = useAuthStore();
  const abortControllersRef = useRef<Map<string, AbortController>>(new Map());
  const chunkUuidRef = useRef<Map<string, Record<number, string>>>(new Map());

  const handleFileSelect = async (file: File) => {
    if (!user) {
      message.error('请先登录');
      return false;
    }

    const allowedTypes = ['video/mp4', 'video/avi', 'video/mov', 'video/wmv', 'video/flv'];
    if (!allowedTypes.includes(file.type)) {
      message.error('只支持视频文件格式（MP4, AVI, MOV, WMV, FLV）');
      return false;
    }

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
      totalChunks,
      uploadedChunks: [],
      progress: 0,
      status: 'waiting',
      currentChunk: 0,
    };

    setUploadTasks((prev) => [...prev, newTask]);

    startUpload(taskId, file).catch((error) => {
      console.error('Upload start failed:', error);
    });

    return false;
  };

  const startUpload = async (taskId: string, file: File) => {
    setUploadTasks((prev) =>
      prev.map((task) => (task.id === taskId ? { ...task, status: 'uploading', error: undefined } : task)),
    );

    try {
      const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
      const fileHash = await calculateFileHash(file);

      const uploadInfo = await apiService.initVideoUpload({
        file_name: file.name,
        file_size: file.size,
        total_chunks: totalChunks,
        user_uuid: user!.user_uuid,
        file_hash: fileHash,
      });

      const uploadedChunkSet = new Set(uploadInfo.upload_chunks ?? []);

      chunkUuidRef.current.set(taskId, {});

      const initialProgress = uploadedChunkSet.size
        ? Math.round((uploadedChunkSet.size / totalChunks) * 100)
        : 0;

      setUploadTasks((prev) =>
        prev.map((task) =>
          task.id === taskId
            ? {
                ...task,
                uploadInfo,
                uploadedChunks: Array.from(uploadedChunkSet).sort((a, b) => a - b),
                progress: initialProgress,
                currentChunk: uploadedChunkSet.size,
                status: 'uploading',
              }
            : task,
        ),
      );

      const abortController = new AbortController();
      abortControllersRef.current.set(taskId, abortController);

      await uploadChunks(taskId, file, uploadInfo, uploadedChunkSet, totalChunks, abortController.signal);
    } catch (error: any) {
      setUploadTasks((prev) =>
        prev.map((task) =>
          task.id === taskId
            ? { ...task, status: 'error', error: error?.message ?? '上传失败' }
            : task,
        ),
      );
      message.error(error?.message ? `上传失败：${error.message}` : '上传失败');
    }
  };

  const getChunkUUID = (taskId: string, uploadVideoUuid: string, index: number) => {
    const map = chunkUuidRef.current.get(taskId) ?? {};
    if (!map[index]) {
      map[index] = `${uploadVideoUuid}-${index}`;
      chunkUuidRef.current.set(taskId, map);
    }
    return map[index];
  };

  const uploadChunks = async (
    taskId: string,
    file: File,
    uploadInfo: UploadVideoInfo,
    uploadedChunkSet: Set<number>,
    totalChunks: number,
    signal: AbortSignal,
  ) => {
    const uploadVideoUuid = uploadInfo.upload_video_uuid;

    for (let index = 0; index < totalChunks; index += 1) {
      if (uploadedChunkSet.has(index)) {
        continue;
      }

      if (signal.aborted) {
        throw new Error('上传已取消');
      }

      const start = index * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, end);
      const chunkArrayBuffer = await chunk.arrayBuffer();
      const chunkHash = await calculateChunkHash(chunkArrayBuffer);

      const chunkUUID = getChunkUUID(taskId, uploadVideoUuid, index);

      await apiService.uploadChunk({
        chunk_uuid: chunkUUID,
        user_uuid: user!.user_uuid,
        upload_video_uuid: uploadVideoUuid,
        chunk_size: chunk.size,
        chunk_index: index,
        chunk_data: chunkArrayBuffer,
        chunk_hash: chunkHash,
      });

      uploadedChunkSet.add(index);
      const progress = Math.round((uploadedChunkSet.size / totalChunks) * 100);

      setUploadTasks((prev) =>
        prev.map((task) =>
          task.id === taskId
            ? {
                ...task,
                progress,
                currentChunk: uploadedChunkSet.size,
                uploadedChunks: Array.from(uploadedChunkSet).sort((a, b) => a - b),
              }
            : task,
        ),
      );
    }

    await mergeChunks(taskId, uploadVideoUuid);
  };

  const mergeChunks = async (taskId: string, uploadVideoUuid: string) => {
    try {
      await apiService.mergeChunks({
        upload_video_uuid: uploadVideoUuid,
        user_uuid: user!.user_uuid,
      });

      setUploadTasks((prev) =>
        prev.map((task) =>
          task.id === taskId
            ? {
                ...task,
                status: 'completed',
                progress: 100,
              }
            : task,
        ),
      );

      message.success('视频上传完成！');
    } catch (error: any) {
      throw new Error(error?.message ? `合并文件失败: ${error.message}` : '合并文件失败');
    } finally {
      abortControllersRef.current.delete(taskId);
    }
  };

  const pauseUpload = (taskId: string) => {
    const abortController = abortControllersRef.current.get(taskId);
    if (abortController) {
      abortController.abort();
      abortControllersRef.current.delete(taskId);
    }

    setUploadTasks((prev) =>
      prev.map((task) => (task.id === taskId ? { ...task, status: 'paused' } : task)),
    );
  };

  const resumeUpload = async (taskId: string) => {
    const task = uploadTasks.find((item) => item.id === taskId);
    if (!task || !task.uploadInfo) {
      message.error('无法恢复上传：任务信息不完整');
      return;
    }

    const abortController = new AbortController();
    abortControllersRef.current.set(taskId, abortController);

    setUploadTasks((prev) =>
      prev.map((item) =>
        item.id === taskId ? { ...item, status: 'uploading', error: undefined } : item,
      ),
    );

    try {
      await uploadChunks(
        taskId,
        task.file,
        task.uploadInfo,
        new Set(task.uploadedChunks),
        task.totalChunks,
        abortController.signal,
      );
    } catch (error: any) {
      if (abortController.signal.aborted) {
        message.info('上传已取消');
        return;
      }
      setUploadTasks((prev) =>
        prev.map((item) =>
          item.id === taskId
            ? { ...item, status: 'error', error: error?.message ?? '上传失败' }
            : item,
        ),
      );
      message.error(error?.message ? `上传失败：${error.message}` : '上传失败');
    }
  };

  const removeTask = (taskId: string) => {
    const abortController = abortControllersRef.current.get(taskId);
    if (abortController) {
      abortController.abort();
      abortControllersRef.current.delete(taskId);
    }
    chunkUuidRef.current.delete(taskId);
    setUploadTasks((prev) => prev.filter((task) => task.id !== taskId));
  };

  const getStatusColor = (status: UploadStatus) => {
    switch (status) {
      case 'waiting':
        return 'default';
      case 'uploading':
        return 'processing';
      case 'paused':
        return 'warning';
      case 'completed':
        return 'success';
      case 'error':
        return 'error';
      default:
        return 'default';
    }
  };

  const getStatusText = (status: UploadStatus) => {
    switch (status) {
      case 'waiting':
        return '等待中';
      case 'uploading':
        return '上传中';
      case 'paused':
        return '已暂停';
      case 'completed':
        return '已完成';
      case 'error':
        return '上传失败';
      default:
        return '未知状态';
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
          <p className="ant-upload-hint">支持 MP4, AVI, MOV, WMV, FLV 格式，单个文件最大 5GB</p>
        </Dragger>

        {uploadTasks.length > 0 && (
          <>
            <Title level={4}>上传任务</Title>
            <List
              dataSource={uploadTasks}
              renderItem={(task) => (
                <List.Item
                  actions={[
                    task.status === 'uploading' && (
                      <Button key="pause" icon={<PauseCircleOutlined />} onClick={() => pauseUpload(task.id)} size="small">
                        暂停
                      </Button>
                    ),
                    task.status === 'paused' && (
                      <Button
                        key="resume"
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
                        key="delete"
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
                        <Tag color={getStatusColor(task.status)}>{getStatusText(task.status)}</Tag>
                      </Space>
                    }
                    description={
                      <div>
                        <Text type="secondary">
                          大小: {(task.file.size / 1024 / 1024).toFixed(2)} MB
                        </Text>
                        <Text type="secondary" style={{ marginLeft: 16 }}>
                          已上传分片: {task.uploadedChunks.length}/{task.totalChunks}
                        </Text>
                        <div style={{ marginTop: 8 }}>
                          <Progress
                            percent={task.progress}
                            status={task.status === 'error' ? 'exception' : 'active'}
                            size="small"
                          />
                        </div>
                        {task.error && (
                          <Alert message={task.error} type="error" style={{ marginTop: 8 }} />
                        )}
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </>
        )}
      </Card>
    </div>
  );
};

export default VideoUpload;
