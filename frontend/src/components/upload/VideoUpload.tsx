import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Upload, Button, Progress, Card, message, Typography, Space, Alert, Tag, Form, Input, Select, Image } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { useAuthStore } from '@/store/auth';
import apiService from '@/services/api';
import { useVideoStatusSubscription } from '@/hooks/useVideoStatusSubscription';
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
  cover_url?: string;
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
  const [publishLoading, setPublishLoading] = useState(false);
  const [step, setStep] = useState<'select' | 'edit'>('select');
  const [publishForm] = Form.useForm<PublishFormValues>();
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const abortControllersRef = useRef<Map<string, AbortController>>(new Map());
  const chunkUuidRef = useRef<Map<string, Record<number, string>>>(new Map());
  const lastCoverTaskIdRef = useRef<string | null>(null);
  const [coverPreviewUrl, setCoverPreviewUrl] = useState<string | undefined>();
  const [coverKey, setCoverKey] = useState<string | undefined>();
  const [coverUploading, setCoverUploading] = useState(false);
  const [tagOptions, setTagOptions] = useState<{ label: string; value: string }[]>([]);
  const currentTask = useMemo(() => uploadTasks[0], [uploadTasks]);

  const handleVideoStatusEvent = useCallback(
    (video: VideoDetail) => {
      let notifyType: 'Published' | 'Failed' | null = null;
      setUploadTasks((prev) => {
        let matched = false;
        const next = prev.map((task) => {
          const uploadUuid = task.uploadInfo?.upload_video_uuid || task.publishedVideo?.upload_video_uuid;
          if (!uploadUuid || uploadUuid !== video.upload_video_uuid) {
            return task;
          }
          matched = true;
          return {
            ...task,
            publishedVideo: video,
          };
        });
        if (!matched) {
          return prev;
        }
        if (video.status === 'Published' || video.status === 'Failed') {
          notifyType = video.status;
        }
        return next;
      });

      if (notifyType === 'Published') {
        message.success(`视频《${video.title}》已发布，可以前往管理页查看`);
      } else if (notifyType === 'Failed') {
        const reason = video.error_message ? `：${video.error_message}` : '';
        message.error(`视频《${video.title}》发布失败${reason}`);
      }
    },
    [setUploadTasks],
  );

  useVideoStatusSubscription(handleVideoStatusEvent, !!user);

  useEffect(() => {
    let canceled = false;
    apiService
      .listTags()
      .then((res) => {
        const opts = (res.list || []).map((t) => ({ label: t.name, value: t.name }));
        if (!canceled) setTagOptions(opts);
      })
      .catch((err) => {
        console.warn('加载标签列表失败', err);
      });
    return () => {
      canceled = true;
    };
  }, []);

  useEffect(() => {
    if (!currentTask) return;
    publishForm.setFieldsValue({
      title: getDefaultTitle(currentTask.file.name),
      description: '',
      tags: currentTask.publishedVideo?.tags ?? [],
      cover_url: coverKey,
    });
  }, [currentTask, publishForm, coverKey]);

  const handlePublishSubmit = async () => {
    try {
      const values = await publishForm.validateFields();
      if (!currentTask || !currentTask.uploadInfo) {
        message.error('上传任务未准备好');
        return;
      }
      setPublishLoading(true);
      const publishedVideo = await apiService.publishVideo({
        upload_video_uuid: currentTask.uploadInfo.upload_video_uuid,
        title: values.title,
        description: values.description,
        tags: values.tags || [],
        cover_url: coverKey,
      });

      setUploadTasks((prev) =>
        prev.map((task) =>
          task.id === currentTask.id ? { ...task, publishedVideo } : task,
        ),
      );

      message.success('视频发布成功，已进入转码中');
    } catch (error: any) {
      if (error?.errorFields) {
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
    setStep('edit');

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
        await pollUploadStatus(taskId, uploadInfo.upload_video_uuid);
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

  const pollUploadStatus = async (taskId: string, uploadVideoUuid: string, intervalMs = 3000, timeoutMs = 300000) => {
    const start = Date.now();
    for (; ;) {
      const now = Date.now();
      if (now - start > timeoutMs) {
        throw new Error('合并超时');
      }
      const res = await apiService.getUploadStatus({ upload_video_uuid: uploadVideoUuid, user_uuid: user!.user_uuid });
      if (res.status === 'Success') {
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
        break;
      }
      if (res.status === 'Failed') {
        setUploadTasks((prev) =>
          prev.map((task) => (task.id === taskId ? { ...task, status: 'error', error: '合并失败' } : task)),
        );
        message.error('合并失败');
        break;
      }
      await new Promise((r) => setTimeout(r, intervalMs));
    }
    abortControllersRef.current.delete(taskId);
  };

  const mergeChunks = async (taskId: string, uploadVideoUuid: string) => {
    try {
      await apiService.mergeChunks({ upload_video_uuid: uploadVideoUuid, user_uuid: user!.user_uuid });
      setUploadTasks((prev) =>
        prev.map((task) => (task.id === taskId ? { ...task, status: 'uploading', progress: Math.max(task.progress, 95) } : task)),
      );
      await pollUploadStatus(taskId, uploadVideoUuid);
    } catch (error: any) {
      throw new Error(error?.message ? `合并文件失败: ${error.message}` : '合并文件失败');
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
    message.info('已暂停上传');
  };

  const resumeUpload = async (taskId: string) => {
    const task = uploadTasks.find((t) => t.id === taskId);
    if (!task || !task.uploadInfo) {
      message.error('上传任务未准备好');
      return;
    }
    const uploadedChunkSet = new Set<number>(task.uploadedChunks || []);
    const totalChunks = Math.max(1, Math.ceil(task.file.size / CHUNK_SIZE));
    const abortController = new AbortController();
    abortControllersRef.current.set(taskId, abortController);
    setUploadTasks((prev) => prev.map((t) => (t.id === taskId ? { ...t, status: 'uploading' } : t)));
    try {
      await uploadChunks(taskId, task.file, task.uploadInfo, uploadedChunkSet, totalChunks, abortController.signal);
    } catch (error: any) {
      if (isAbortError(error)) {
        return;
      }
      setUploadTasks((prev) => prev.map((t) => (t.id === taskId ? { ...t, status: 'error', error: error?.message || '上传失败' } : t)));
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
    setStep('select');
    setCoverPreviewUrl(undefined);
    setCoverKey(undefined);
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

  

  const uploadCoverBlob = async (blob: Blob, suggestedName: string) => {
    setCoverUploading(true);
    try {
      const coverFileName = suggestedName.toLowerCase().endsWith('.jpg') || suggestedName.toLowerCase().endsWith('.png')
        ? suggestedName
        : `${suggestedName}.jpg`;
      const presign = await apiService.presignImage({ file_name: coverFileName, category: 'cover', expires_seconds: 900 });
      await fetch(presign.put_url, { method: 'PUT', headers: { 'Content-Type': 'image/jpeg' }, body: blob });
      setCoverKey(presign.key);
      const localUrl = URL.createObjectURL(blob);
      setCoverPreviewUrl(localUrl);
      publishForm.setFieldsValue({ cover_url: presign.key });
      message.success('封面上传成功');
    } catch (e: any) {
      message.error(e?.message || '封面上传失败');
    } finally {
      setCoverUploading(false);
    }
  };

  const handleCoverFileSelect = async (file: File) => {
    const isImage = /^image\/(png|jpeg|jpg)$/i.test(file.type);
    if (!isImage) {
      message.error('只支持 PNG 或 JPG 图片');
      return Upload.LIST_IGNORE;
    }
    const blob = file;
    await uploadCoverBlob(blob, file.name);
    return false;
  };

  const extractFirstFrame = async (file: File) => {
    try {
      const url = URL.createObjectURL(file);
      const video = document.createElement('video');
      video.src = url;
      video.muted = true;
      await new Promise((resolve, reject) => {
        const onLoaded = () => resolve(undefined);
        const onError = () => reject(new Error('视频预加载失败'));
        video.addEventListener('loadeddata', onLoaded, { once: true });
        video.addEventListener('error', onError, { once: true });
      });
      try { video.currentTime = 0.1; } catch {}
      await new Promise((resolve) => setTimeout(resolve, 150));
      const canvas = document.createElement('canvas');
      canvas.width = video.videoWidth || 640;
      canvas.height = video.videoHeight || 360;
      const ctx = canvas.getContext('2d');
      if (!ctx) throw new Error('Canvas 不支持');
      ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
      const blob: Blob = await new Promise((resolve) => canvas.toBlob((b) => resolve(b as Blob), 'image/jpeg', 0.85));
      URL.revokeObjectURL(url);
      if (!blob) throw new Error('封面生成失败');
      const base = getDefaultTitle(file.name) + '-cover';
      await uploadCoverBlob(blob, `${base}.jpg`);
    } catch {}
  };

  useEffect(() => {
    if (!currentTask) return;
    if (lastCoverTaskIdRef.current === currentTask.id) return;
    lastCoverTaskIdRef.current = currentTask.id;
    extractFirstFrame(currentTask.file);
  }, [currentTask?.id]);

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Title level={3}>视频上传</Title>
        {step === 'select' && (
          <Dragger name="file" multiple={false} beforeUpload={handleFileSelect} showUploadList={false} style={{ marginBottom: 24 }}>
            <p className="ant-upload-drag-icon">
              <UploadOutlined style={{ fontSize: 48, color: '#1890ff' }} />
            </p>
            <p className="ant-upload-text">点击或拖拽视频文件到此区域上传</p>
            <p className="ant-upload-hint">支持 MP4, AVI, MOV, WMV, FLV, MKV 格式，单个文件最大 5GB</p>
          </Dragger>
        )}
        {step === 'edit' && currentTask && (
          <div>
            <Space direction="vertical" style={{ width: '100%' }} size={16}>
              <Card bordered={false} style={{ background: '#fafafa' }}>
                <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
                  <div>
                    <Text strong>{currentTask.file.name}</Text>
                    <Tag color={getStatusColor(currentTask.status)} style={{ marginLeft: 8 }}>{getStatusText(currentTask.status)}</Tag>
                  </div>
                  <Space>
                    {currentTask.status === 'uploading' && currentTask.progress < 95 && (
                      <Button onClick={() => pauseUpload(currentTask.id)}>暂停</Button>
                    )}
                    {currentTask.status === 'paused' && (
                      <Button type="primary" onClick={() => resumeUpload(currentTask.id)}>继续</Button>
                    )}
                    <Button danger onClick={() => removeTask(currentTask.id)}>重新选择</Button>
                  </Space>
                </Space>
                <div style={{ marginTop: 12 }}>
                  <Progress percent={currentTask.progress} status={currentTask.status === 'error' ? 'exception' : 'active'} />
                </div>
                {currentTask.error && <Alert message={currentTask.error} type="error" style={{ marginTop: 8 }} />}
              </Card>

              <Form form={publishForm} layout="vertical" onFinish={handlePublishSubmit}>
                <Form.Item label="封面" name="cover_url">
                  <Space align="start">
                    {coverPreviewUrl ? (
                      <Image src={coverPreviewUrl} width={200} height={112} style={{ objectFit: 'cover' }} />
                    ) : (
                      <div style={{ width: 200, height: 112, background: '#f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                        <Text type="secondary">暂无封面</Text>
                      </div>
                    )}
                    <Upload accept="image/png,image/jpeg" showUploadList={false} beforeUpload={handleCoverFileSelect}>
                      <Button loading={coverUploading} disabled={coverUploading}>选择封面</Button>
                    </Upload>
                  </Space>
                </Form.Item>
                <Form.Item label="标题" name="title" rules={[{ required: true, message: '请输入视频标题' }, { max: 120, message: '标题不能超过120个字符' }]}>
                  <Input placeholder="请输入视频标题" />
                </Form.Item>
                <Form.Item label="简介" name="description" rules={[{ max: 2000, message: '简介不能超过2000个字符' }]}>
                  <Input.TextArea rows={4} placeholder="简单介绍一下您的视频" />
                </Form.Item>
                <Form.Item label="标签" name="tags">
                  <Select
                    mode="multiple"
                    style={{ width: '100%' }}
                    placeholder="从标签库选择"
                    options={tagOptions}
                  />
                </Form.Item>
                <Form.Item>
                  <Button type="primary" htmlType="submit" loading={publishLoading} disabled={!currentTask.uploadInfo}>发布</Button>
                </Form.Item>
              </Form>
            </Space>
          </div>
        )}
      </Card>
    </div>
  );
};

export default VideoUpload;
