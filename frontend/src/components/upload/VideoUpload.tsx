import React, { useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
  Modal,
  Form,
  Input,
  Select,
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
import { UploadVideoInfo, VideoDetail } from '@/types/api';

const { Title, Text } = Typography;
const { Dragger } = Upload;

const CHUNK_SIZE = 1024 * 1024 * 5; // 5MB per chunk (MinIO ComposeObject minimum requirement)

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
  publishedVideo?: VideoDetail;
}

interface PublishFormValues {
  title: string;
  description?: string;
  tags: string[];
}

const createAbortError = () => {
  const abortError = new Error('上传已取消');
  (abortError as Error & { name: string }).name = 'AbortError';
  return abortError;
};

const isAbortError = (error: unknown, signal?: AbortSignal): boolean => {
  if (signal?.aborted) {
    return true;
  }
  if (!error || typeof error !== 'object') {
    return false;
  }
  const err = error as { name?: string; code?: string; message?: string };
  return (
    err.name === 'AbortError' ||
    err.name === 'CanceledError' ||
    err.code === 'ERR_CANCELED' ||
    err.message === 'canceled'
  );
};

const getDefaultTitle = (fileName: string) => {
  if (!fileName) {
    return '';
  }
  const segments = fileName.split('.');
  if (segments.length === 1) {
    return fileName;
  }
  segments.pop();
  return segments.join('.');
};

const VideoUpload: React.FC = () => {
  const [uploadTasks, setUploadTasks] = useState<UploadTask[]>([]);
  const [publishModalVisible, setPublishModalVisible] = useState(false);
  const [publishLoading, setPublishLoading] = useState(false);
  const [currentPublishTask, setCurrentPublishTask] = useState<UploadTask | null>(null);
  const [publishForm] = Form.useForm<PublishFormValues>();
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const abortControllersRef = useRef<Map<string, AbortController>>(new Map());
  const chunkUuidRef = useRef<Map<string, Record<number, string>>>(new Map());

  const closePublishModal = () => {
    setPublishModalVisible(false);
    setPublishLoading(false);
    setCurrentPublishTask(null);
    publishForm.resetFields();
  };

  const openPublishModal = (task: UploadTask) => {
    if (!task.uploadInfo) {
      message.warning('上传信息尚未准备完成，请稍后重试');
      return;
    }
    setCurrentPublishTask(task);
    publishForm.resetFields();
    publishForm.setFieldsValue({
      title: getDefaultTitle(task.file.name),
      description: '',
      tags: task.publishedVideo?.tags ?? [],
    });
    setPublishModalVisible(true);
  };

  const handlePublishSubmit = async () => {
    try {
      const values = await publishForm.validateFields();

      if (!currentPublishTask || !currentPublishTask.uploadInfo) {
        message.error('未找到对应的上传任务，请刷新后重试');
        return;
      }

      setPublishLoading(true);
      const publishedVideo = await apiService.publishVideo({
        upload_video_uuid: currentPublishTask.uploadInfo.upload_video_uuid,
        title: values.title,
        description: values.description,
        tags: values.tags || [],
      });

      setUploadTasks((prev) =>
        prev.map((task) =>
          task.id === currentPublishTask.id ? { ...task, publishedVideo } : task,
        ),
      );

      message.success('视频发布成功');
      closePublishModal();
    } catch (error: any) {
      if (error?.errorFields) {
        // 表单校验错误由 Ant Design 统一提示
        return;
      }
      console.error('Publish video failed:', error);
      const errorMessage = error?.response?.data?.message || '发布失败，请稍后重试';
      message.error(errorMessage);
    } finally {
      setPublishLoading(false);
    }
  };

  const handleFileSelect = async (file: File) => {
    if (!user) {
      message.error('请先登录');
      navigate('/login', { replace: true });
      return false;
    }

    const allowedTypes = ['video/mp4', 'video/avi', 'video/mov', 'video/wmv', 'video/flv', 'video/x-matroska'];
    if (!allowedTypes.includes(file.type)) {
      message.error('只支持视频文件格式（MP4, AVI, MOV, WMV, FLV, MKV）');
      return false;
    }

    const maxSize = 5 * 1024 * 1024 * 1024;
    if (file.size > maxSize) {
      message.error('文件大小不能超过5GB');
      return false;
    }

    const taskId = generateUUID();
    const totalChunks = Math.max(1, Math.ceil(file.size / CHUNK_SIZE));

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
      const totalChunks = Math.max(1, Math.ceil(file.size / CHUNK_SIZE));
      const fileHash = await calculateFileHash(file);

      const uploadInfo = await apiService.initVideoUpload({
        file_name: file.name,
        file_size: file.size,
        total_chunks: totalChunks,
        user_uuid: user!.user_uuid,
        file_hash: fileHash,
      });

      // 检查上传视频的状态
      if (uploadInfo.status === 'Success') {
        // 如果已经上传完成，直接标记为完成
        setUploadTasks((prev) =>
          prev.map((task) =>
            task.id === taskId
              ? {
                  ...task,
                  uploadInfo,
                  progress: 100,
                  status: 'completed',
                }
              : task,
          ),
        );
        message.success('文件已存在，无需重复上传！');
        return;
      }

      if (uploadInfo.status === 'Failed') {
        // 如果之前上传失败，提示用户重新开始
        setUploadTasks((prev) =>
          prev.map((task) =>
            task.id === taskId
              ? { ...task, status: 'error', error: '之前的上传已失败，请重新上传' }
              : task,
          ),
        );
        message.error('之前的上传已失败，请重新上传');
        return;
      }

      if (uploadInfo.status === 'Merging') {
        // 如果正在合并中，提示用户等待
        setUploadTasks((prev) =>
          prev.map((task) =>
            task.id === taskId
              ? {
                  ...task,
                  uploadInfo,
                  progress: 95,
                  status: 'uploading',
                }
              : task,
          ),
        );
        message.info('文件正在合并中，请稍候...');
        return;
      }

      // 构建chunk UUID映射
      const chunkUuidMap: { [index: number]: string } = {};
      const uploadedChunkSet = new Set<number>();
      
      uploadInfo.upload_chunks?.forEach(chunk => {
        chunkUuidMap[chunk.chunk_index] = chunk.chunk_uuid;
        // 只有状态为Completed的分片才算已上传
        if (chunk.status === 'Completed') {
          uploadedChunkSet.add(chunk.chunk_index);
        }
      });
      chunkUuidRef.current.set(taskId, chunkUuidMap);

      const initialProgress = uploadedChunkSet.size
        ? Math.round((uploadedChunkSet.size / totalChunks) * 100)
        : 0;

      // 如果所有分片都已上传完成，直接进行合并
      if (uploadedChunkSet.size === totalChunks) {
        setUploadTasks((prev) =>
          prev.map((task) =>
            task.id === taskId
              ? {
                  ...task,
                  uploadInfo,
                  uploadedChunks: Array.from(uploadedChunkSet).sort((a, b) => a - b),
                  progress: 95,
                  currentChunk: uploadedChunkSet.size,
                  status: 'uploading',
                }
              : task,
          ),
        );
        
        message.info('所有分片已上传完成，开始合并文件...');
        await mergeChunks(taskId, uploadInfo.upload_video_uuid);
        return;
      }

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
      if (isAbortError(error)) {
        return;
      }
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
    // 如果已经有后端返回的chunk_uuid，直接使用
    if (map[index]) {
      return map[index];
    }
    // 如果没有，生成一个临时的（这种情况不应该发生，因为后端已经返回了所有chunk的UUID）
    console.warn(`Missing chunk UUID for index ${index}, generating fallback`);
    map[index] = `${uploadVideoUuid}-${index}`;
    chunkUuidRef.current.set(taskId, map);
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
    let successfulUploads = 0;
    const errors: string[] = [];
    const abortIfNeeded = () => {
      if (signal.aborted) {
        throw createAbortError();
      }
    };

    console.log(`开始上传分片，总分片数: ${totalChunks}, 已上传分片: ${uploadedChunkSet.size}`);

    for (let index = 0; index < totalChunks; index += 1) {
      if (uploadedChunkSet.has(index)) {
        successfulUploads += 1;
        console.log(`分片 ${index} 已存在，跳过上传`);
        continue;
      }

      abortIfNeeded();

      try {
        console.log(`开始上传分片 ${index}/${totalChunks - 1}`);
        
        const start = index * CHUNK_SIZE;
        const end = Math.min(start + CHUNK_SIZE, file.size);
        const chunk = file.slice(start, end);
        const chunkArrayBuffer = await chunk.arrayBuffer();
        abortIfNeeded();
        const chunkHash = await calculateChunkHash(chunkArrayBuffer);

        const chunkUUID = getChunkUUID(taskId, uploadVideoUuid, index);

        console.log(`分片 ${index} 信息:`, {
          chunkUUID,
          chunkSize: chunk.size,
          chunkIndex: index,
          chunkHash: chunkHash.substring(0, 8) + '...',
        });

        // 发送分片上传请求
        await apiService.uploadChunk(
          {
            chunk_uuid: chunkUUID,
            user_uuid: user!.user_uuid,
            upload_video_uuid: uploadVideoUuid,
            chunk_size: chunk.size,
            chunk_index: index,
            chunk_data: chunkArrayBuffer,
            chunk_hash: chunkHash,
          },
          { signal },
        );

        console.log(`分片 ${index} 上传成功`);
        
        uploadedChunkSet.add(index);
        successfulUploads += 1;
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
      } catch (error: any) {
        if (isAbortError(error, signal)) {
          throw createAbortError();
        }
        const errorMessage = `分片 ${index} 上传失败: ${error?.message || '未知错误'}`;
        errors.push(errorMessage);
        console.error(`分片 ${index} 上传失败:`, error);
        console.error('错误详情:', {
          message: error?.message,
          status: error?.response?.status,
          statusText: error?.response?.statusText,
          data: error?.response?.data,
        });
      }
    }

    console.log(`分片上传完成，成功: ${successfulUploads}/${totalChunks}`);

    // 验证所有分片是否都成功上传
    if (successfulUploads !== totalChunks) {
      const errorMessage = `上传失败：成功上传 ${successfulUploads}/${totalChunks} 个分片。错误详情：${errors.join('; ')}`;
      console.error('分片上传验证失败:', errorMessage);
      throw new Error(errorMessage);
    }

    console.log('所有分片上传成功，开始合并文件');
    // 只有所有分片都成功上传后才调用合并接口
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
    if (!task) {
      message.error('无法恢复上传：任务信息不完整');
      return;
    }

    // 清理之前的缓存数据
    abortControllersRef.current.delete(taskId);
    chunkUuidRef.current.delete(taskId);

    setUploadTasks((prev) =>
      prev.map((item) =>
        item.id === taskId ? { ...item, status: 'uploading', error: undefined } : item,
      ),
    );

    try {
      // 重新计算文件信息
      const totalChunks = Math.max(1, Math.ceil(task.file.size / CHUNK_SIZE));
      const fileHash = await calculateFileHash(task.file);

      // 重新调用初始化接口获取最新的分片状态
      const uploadInfo = await apiService.initVideoUpload({
        file_name: task.file.name,
        file_size: task.file.size,
        total_chunks: totalChunks,
        user_uuid: user!.user_uuid,
        file_hash: fileHash,
      });

      // 检查上传视频的状态
      if (uploadInfo.status === 'Success') {
        // 如果已经上传完成，直接标记为完成
        setUploadTasks((prev) =>
          prev.map((item) =>
            item.id === taskId
              ? {
                  ...item,
                  uploadInfo,
                  progress: 100,
                  status: 'completed',
                }
              : item,
          ),
        );
        message.success('文件已上传完成！');
        return;
      }

      if (uploadInfo.status === 'Failed') {
        // 如果之前上传失败，提示用户重新开始
        setUploadTasks((prev) =>
          prev.map((item) =>
            item.id === taskId
              ? { ...item, status: 'error', error: '之前的上传已失败，请重新上传' }
              : item,
          ),
        );
        message.error('之前的上传已失败，请重新上传');
        return;
      }

      if (uploadInfo.status === 'Merging') {
        // 如果正在合并中，提示用户等待
        setUploadTasks((prev) =>
          prev.map((item) =>
            item.id === taskId
              ? {
                  ...item,
                  uploadInfo,
                  progress: 95,
                  status: 'uploading',
                }
              : item,
          ),
        );
        message.info('文件正在合并中，请稍候...');
        return;
      }

      // 重新构建chunk UUID映射和已上传分片集合
      const chunkUuidMap: { [index: number]: string } = {};
      const uploadedChunkSet = new Set<number>();
      
      uploadInfo.upload_chunks?.forEach(chunk => {
        chunkUuidMap[chunk.chunk_index] = chunk.chunk_uuid;
        // 只有状态为Completed的分片才算已上传
        if (chunk.status === 'Completed') {
          uploadedChunkSet.add(chunk.chunk_index);
        }
      });
      chunkUuidRef.current.set(taskId, chunkUuidMap);

      const currentProgress = uploadedChunkSet.size
        ? Math.round((uploadedChunkSet.size / totalChunks) * 100)
        : 0;

      // 如果所有分片都已上传完成，直接进行合并
      if (uploadedChunkSet.size === totalChunks) {
        setUploadTasks((prev) =>
          prev.map((item) =>
            item.id === taskId
              ? {
                  ...item,
                  uploadInfo,
                  uploadedChunks: Array.from(uploadedChunkSet).sort((a, b) => a - b),
                  progress: 95,
                  currentChunk: uploadedChunkSet.size,
                  status: 'uploading',
                }
              : item,
          ),
        );
        
        message.info('所有分片已上传完成，开始合并文件...');
        await mergeChunks(taskId, uploadInfo.upload_video_uuid);
        return;
      }

      // 更新任务状态
      setUploadTasks((prev) =>
        prev.map((item) =>
          item.id === taskId
            ? {
                ...item,
                uploadInfo,
                uploadedChunks: Array.from(uploadedChunkSet).sort((a, b) => a - b),
                progress: currentProgress,
                currentChunk: uploadedChunkSet.size,
                totalChunks,
                status: 'uploading',
              }
            : item,
        ),
      );

      const abortController = new AbortController();
      abortControllersRef.current.set(taskId, abortController);

      // 继续上传剩余分片
      await uploadChunks(
        taskId,
        task.file,
        uploadInfo,
        uploadedChunkSet,
        totalChunks,
        abortController.signal,
      );
    } catch (error: any) {
      if (isAbortError(error)) {
        message.info('上传已暂停');
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
          <p className="ant-upload-hint">支持 MP4, AVI, MOV, WMV, FLV, MKV 格式，单个文件最大 5GB</p>
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
                    task.status === 'completed' && !task.publishedVideo && (
                      <Button
                        key="publish"
                        type="primary"
                        size="small"
                        onClick={() => openPublishModal(task)}
                      >
                        发布
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
                        {task.publishedVideo && (
                          <div style={{ marginTop: 12 }}>
                            <Tag color="gold">已发布</Tag>
                            <Text style={{ marginLeft: 8 }}>
                              标题: {task.publishedVideo.title}
                            </Text>
                            {task.publishedVideo.tags?.length > 0 && (
                              <div style={{ marginTop: 4 }}>
                                {task.publishedVideo.tags.map((publishTag) => (
                                  <Tag key={publishTag} color="blue">
                                    {publishTag}
                                  </Tag>
                                ))}
                              </div>
                            )}
                          </div>
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
      <Modal
        title="发布视频"
        open={publishModalVisible}
        onCancel={closePublishModal}
        onOk={handlePublishSubmit}
        okText="发布"
        cancelText="取消"
        confirmLoading={publishLoading}
        destroyOnClose
      >
        <Form form={publishForm} layout="vertical">
          <Form.Item
            label="标题"
            name="title"
            rules={[
              { required: true, message: '请输入视频标题' },
              { max: 120, message: '标题不能超过120个字符' },
            ]}
          >
            <Input placeholder="请输入视频标题" />
          </Form.Item>
          <Form.Item
            label="简介"
            name="description"
            rules={[{ max: 2000, message: '简介不能超过2000个字符' }]}
          >
            <Input.TextArea rows={4} placeholder="简单介绍一下您的视频" />
          </Form.Item>
          <Form.Item label="标签" name="tags" extra="输入后回车可快速创建标签">
            <Select
              mode="tags"
              style={{ width: '100%' }}
              placeholder="例如：教程, 游戏, 音乐"
              tokenSeparators={[',', ' ']}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default VideoUpload;
