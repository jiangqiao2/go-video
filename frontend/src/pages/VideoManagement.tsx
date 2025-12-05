import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Typography, Button, Space, App, List, Tag, Empty } from 'antd';
import {
  VideoCameraOutlined,
  EditOutlined,
  PlayCircleOutlined,
  LikeOutlined,
  MessageOutlined,
} from '@ant-design/icons';
import apiService from '@/services/api';
import { VideoDetail } from '@/types/api';
import { useAuthStore } from '@/store/auth';
import { useVideoStatusSubscription } from '@/hooks/useVideoStatusSubscription';
import { formatPublishedTime } from '@/utils/date';
import CreatorLayout from '@/components/layout/CreatorLayout';

const { Text, Paragraph } = Typography;

type VideoStatusKey = 'all' | 'processing' | 'published' | 'failed' | 'draft';
const statusValueMap: Record<VideoStatusKey, string> = {
  all: '',
  processing: 'Processing',
  published: 'Published',
  failed: 'Failed',
  draft: 'Draft',
};

const statusMetaMap: Record<
  string,
  { color: string; text: string; description?: string }
> = {
  Draft: { color: 'default', text: '草稿', description: '待发布或者等待转码任务创建' },
  Processing: { color: 'processing', text: '转码中', description: '转码任务正在处理视频' },
  Published: { color: 'success', text: '已发布', description: '视频已转码完成并可播放' },
  Failed: { color: 'error', text: '转码失败', description: '转码过程中出现异常' },
};

const defaultPageSize = 8;

const VideoManagement: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const { user } = useAuthStore();
  const [videos, setVideos] = useState<VideoDetail[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultPageSize);
  const [total, setTotal] = useState(0);
  const [statusKey, setStatusKey] = useState<VideoStatusKey>('all');

  const currentStatusValue = useMemo(() => {
    return statusValueMap[statusKey] ?? '';
  }, [statusKey]);

  const fetchVideos = useCallback(async () => {
    setLoading(true);
    try {
      const response = await apiService.listUserVideos({
        page,
        size: pageSize,
        status: currentStatusValue || undefined,
      });
      setVideos(response.videos);
      setTotal(response.total);
      if (response.page !== page) {
        setPage(response.page);
      }
      if (response.size !== pageSize) {
        setPageSize(response.size);
      }
    } catch (error: any) {
      console.error('加载视频列表失败', error);
      const errorMessage =
        error?.response?.data?.message ||
        error?.message ||
        '获取视频列表失败，请稍后重试';
      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, currentStatusValue]);

  const handleVideoStatusEvent = useCallback(
    (video: VideoDetail) => {
      let shouldRefresh = false;
      setVideos((prev) => {
        const index = prev.findIndex((item) => item.video_uuid === video.video_uuid);
        if (index === -1) {
          return prev;
        }
        const next = [...prev];
        if (currentStatusValue && video.status !== currentStatusValue) {
          next.splice(index, 1);
          shouldRefresh = true;
          return next;
        }
        next[index] = video;
        return next;
      });
      if (shouldRefresh) {
        fetchVideos();
      }
    },
    [currentStatusValue, fetchVideos],
  );

  useVideoStatusSubscription(handleVideoStatusEvent, !!user);

  useEffect(() => {
    fetchVideos();
  }, [fetchVideos]);

  const handleStatusChange = (key: VideoStatusKey) => {
    setStatusKey(key);
    setPage(1);
  };

  const handlePaginationChange = (current: number, size: number) => {
    if (size !== pageSize) {
      setPage(1);
      setPageSize(size);
    } else {
      setPage(current);
    }
  };

  const handleNavigateWatch = (video: VideoDetail) => {
    if (!video?.video_uuid || video.status !== 'Published') return;
    navigate(`/watch/${video.video_uuid}`);
  };

  const statusCounts = useMemo(() => {
    const init = { all: videos.length, processing: 0, published: 0, failed: 0 };
    videos.forEach((v) => {
      const s = (v.status || '').toLowerCase();
      if (s === 'processing') init.processing += 1;
      if (s === 'published') init.published += 1;
      if (s === 'failed') init.failed += 1;
    });
    return init;
  }, [videos]);

  const renderStatus = (status: string) => {
    const meta = statusMetaMap[status] || { color: 'default', text: status || '未知状态' };
    return (
      <Space align="center" size={8}>
        <Tag color={meta.color}>{meta.text}</Tag>
        {meta.description && (
          <Text type="secondary" style={{ fontSize: 12 }}>
            {meta.description}
          </Text>
        )}
      </Space>
    );
  };

  const formatDuration = (seconds?: number) => {
    if (!seconds || !isFinite(seconds) || seconds <= 0) return '--:--';
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    const mm = String(mins).padStart(2, '0');
    const ss = String(secs).padStart(2, '0');
    if (hrs > 0) {
      const hh = String(hrs).padStart(2, '0');
      return `${hh}:${mm}:${ss}`;
    }
    return `${mm}:${ss}`;
  };

  return (
    <CreatorLayout activeKey="videos">
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div
          style={{
            background: '#fff',
            borderRadius: 12,
            padding: '20px 24px',
            boxShadow: 'var(--shadow-sm)',
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'center',
          }}
        >
          <div>
            <Space size={20} wrap>
              <Space size={12}>
                <Text strong>视频管理</Text>
              </Space>
            </Space>
            <div style={{ marginTop: 8 }}>
              <Space size={24} wrap>
                {[
                  { key: 'all' as VideoStatusKey, label: '全部稿件', count: statusCounts.all },
                  { key: 'processing' as VideoStatusKey, label: '进行中', count: statusCounts.processing },
                  { key: 'published' as VideoStatusKey, label: '已通过', count: statusCounts.published },
                  { key: 'failed' as VideoStatusKey, label: '未通过', count: statusCounts.failed },
                ].map((item) => {
                  const active = statusKey === item.key;
                  return (
                    <Button
                      key={item.key}
                      type={active ? 'primary' : 'default'}
                      size="small"
                      ghost={false}
                      onClick={() => handleStatusChange(item.key)}
                    >
                      {item.label} {item.count}
                    </Button>
                  );
                })}
              </Space>
            </div>
          </div>
        </div>

        <div
          style={{
            background: '#fff',
            borderRadius: 12,
            padding: '12px 24px 16px',
            boxShadow: 'var(--shadow-sm)',
          }}
        >
          <List
            itemLayout="vertical"
            dataSource={videos}
            loading={loading}
            locale={{
              emptyText: (
                <Empty
                  image={<VideoCameraOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />}
                  description="还没有视频，快去上传一个吧～"
                />
              ),
            }}
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              onChange: handlePaginationChange,
              pageSizeOptions: ['4', '8', '12', '16'],
            }}
            renderItem={(video) => {
              const stats = [
                { label: '点赞', value: video.like_count ?? 0, icon: <LikeOutlined /> },
                { label: '评论', value: video.comment_count ?? 0, icon: <MessageOutlined /> },
              ];
              const cover = video.cover_url || 'https://picsum.photos/seed/video-cover/320/180';
              const canNavigate = video.status === 'Published' && !!video.video_uuid;
              return (
                <List.Item style={{ padding: '12px 0' }}>
                  <div
                    style={{
                      display: 'grid',
                      gridTemplateColumns: '220px 1fr 180px',
                      gap: 16,
                      padding: 16,
                      border: '1px solid #f0f0f0',
                      borderRadius: 12,
                      boxShadow: '0 4px 12px rgba(0,0,0,0.04)',
                      background: '#fff',
                      cursor: canNavigate ? 'pointer' : 'default',
                    }}
                    role={canNavigate ? 'button' : undefined}
                    tabIndex={canNavigate ? 0 : -1}
                    onClick={() => canNavigate && handleNavigateWatch(video)}
                  >
                    <div style={{ position: 'relative', borderRadius: 8, overflow: 'hidden' }}>
                      <img
                        src={cover}
                        alt={video.title}
                        style={{ width: '100%', height: 120, objectFit: 'cover', display: 'block' }}
                      />
                      <div
                        style={{
                          position: 'absolute',
                          top: 8,
                          left: 8,
                          background: 'rgba(0,0,0,0.55)',
                          color: '#fff',
                          padding: '2px 8px',
                          borderRadius: 6,
                          fontSize: 12,
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: 6,
                        }}
                      >
                        <PlayCircleOutlined />
                        <span>{formatDuration(video.duration_seconds)}</span>
                      </div>
                    </div>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                      <Space size={8} align="center">
                        <Text strong style={{ fontSize: 16 }}>{video.title || '未命名稿件'}</Text>
                        {renderStatus(video.status)}
                      </Space>
                      <Space size={12} wrap style={{ fontSize: 12, color: '#6c6f73' }}>
                        <span>创建：{formatPublishedTime(video.created_at)}</span>
                        {video.status === 'Published' && video.published_at && (
                          <span>发布：{formatPublishedTime(video.published_at)}</span>
                        )}
                      </Space>
                      {video.description && (
                        <Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 0 }}>
                          {video.description}
                        </Paragraph>
                      )}
                      <Space size={20} wrap style={{ color: '#6c6f73' }}>
                        {stats.map((item) => (
                          <Space size={6} key={item.label}>
                            {item.icon}
                            {item.value}
                          </Space>
                        ))}
                      </Space>
                    </div>

                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 8 }}>
                      <Button
                        icon={<EditOutlined />}
                        size="small"
                        disabled={video.status === 'Processing'}
                        onClick={(e) => e.stopPropagation()}
                      >
                        编辑
                      </Button>
                    </div>
                  </div>
                </List.Item>
              );
            }}
          />
        </div>
      </Space>
    </CreatorLayout>
  );
};

export default VideoManagement;
