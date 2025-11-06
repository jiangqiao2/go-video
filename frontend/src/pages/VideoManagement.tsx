import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Layout,
  Typography,
  Button,
  Space,
  Avatar,
  message,
  Tabs,
  List,
  Card,
  Tag,
  Empty,
} from 'antd';
import {
  UserOutlined,
  ReloadOutlined,
  VideoCameraOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import apiService from '@/services/api';
import { VideoDetail } from '@/types/api';
import { useAuthStore } from '@/store/auth';
import { useVideoStatusSubscription } from '@/hooks/useVideoStatusSubscription';

const { Header, Content } = Layout;
const { Title, Text, Paragraph } = Typography;

type VideoStatusKey = 'all' | 'processing' | 'published' | 'failed' | 'draft';

const statusTabs: Array<{ key: VideoStatusKey; label: string; status: string }> = [
  { key: 'all', label: '全部视频', status: '' },
  { key: 'processing', label: '转码中', status: 'Processing' },
  { key: 'published', label: '已发布', status: 'Published' },
  { key: 'failed', label: '转码失败', status: 'Failed' },
  { key: 'draft', label: '草稿', status: 'Draft' },
];

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
  const { user, logout } = useAuthStore();
  const [videos, setVideos] = useState<VideoDetail[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultPageSize);
  const [total, setTotal] = useState(0);
  const [statusKey, setStatusKey] = useState<VideoStatusKey>('all');

  const currentStatusValue = useMemo(() => {
    const tab = statusTabs.find((item) => item.key === statusKey);
    return tab?.status ?? '';
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

  const handleStatusChange = (key: string) => {
    setStatusKey(key as VideoStatusKey);
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

  const handleLogout = () => {
    logout();
    message.success('已退出登录');
    navigate('/login', { replace: true });
  };

  const handleNavigateUpload = () => {
    navigate('/upload');
  };

  const renderStatus = (status: string) => {
    const meta = statusMetaMap[status] || { color: 'default', text: status || '未知状态' };
    return (
      <Space direction="vertical" size={4}>
        <Tag color={meta.color}>{meta.text}</Tag>
        {meta.description && (
          <Text type="secondary" style={{ fontSize: 12 }}>
            {meta.description}
          </Text>
        )}
      </Space>
    );
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header
        style={{
          backgroundColor: '#fff',
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
          <Space size="large">
            <Title level={3} style={{ margin: 0, color: '#1890ff' }}>
              我的创作中心
            </Title>
            <Button type="link" icon={<UploadOutlined />} onClick={handleNavigateUpload}>
              上传视频
            </Button>
          </Space>
          <Space size="middle">
            <Button icon={<ReloadOutlined />} onClick={fetchVideos} disabled={loading}>
              刷新
            </Button>
            {user ? (
              <>
                <Avatar size="small" icon={<UserOutlined />} />
                <Text>{user.account}</Text>
                <Button size="small" onClick={handleLogout}>
                  退出登录
                </Button>
              </>
            ) : (
              <Button type="primary" size="small" onClick={() => navigate('/login')}>
                登录
              </Button>
            )}
          </Space>
        </div>
      </Header>
      <Content style={{ padding: 24 }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Space direction="horizontal" align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
            <Title level={3} style={{ margin: 0 }}>
              视频管理
            </Title>
            <Text type="secondary">随时关注视频转码进度与发布状态</Text>
          </Space>
          <Tabs
            activeKey={statusKey}
            onChange={handleStatusChange}
            items={statusTabs.map((tab) => ({
              key: tab.key,
              label: tab.label,
            }))}
          />
          <List
            grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 4 }}
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
            renderItem={(video) => (
              <List.Item>
                <Card
                  hoverable
                  title={
                    <Space direction="vertical" size={4}>
                      <Text strong>{video.title}</Text>
                      {renderStatus(video.status)}
                    </Space>
                  }
                  actions={[
                    <Button type="link" key="upload" onClick={handleNavigateUpload}>
                      再上传一个
                    </Button>,
                  ]}
                >
                  <Space direction="vertical" size="small" style={{ width: '100%' }}>
                    {video.description && (
                      <Paragraph ellipsis={{ rows: 3 }} style={{ marginBottom: 0 }}>
                        {video.description}
                      </Paragraph>
                    )}
                    <Text type="secondary">
                      视频UUID：{video.video_uuid}
                    </Text>
                    {video.transcode_task_uuid && (
                      <Text type="secondary">转码任务：{video.transcode_task_uuid}</Text>
                    )}
                    {video.status === 'Failed' && video.error_message && (
                      <Text type="danger">错误信息：{video.error_message}</Text>
                    )}
                    {video.status === 'Published' && video.video_url && (
                      <Text>
                        播放地址：<a href={video.video_url} target="_blank" rel="noreferrer">{video.video_url}</a>
                      </Text>
                    )}
                    <Text type="secondary">
                      {video.published_at
                        ? `发布时间：${dayjs(video.published_at).format('YYYY-MM-DD HH:mm')}`
                        : '发布时间：--'}
                    </Text>
                    {video.tags?.length > 0 && (
                      <Space size={[4, 4]} wrap>
                        {video.tags.map((tag) => (
                          <Tag color="blue" key={tag}>
                            {tag}
                          </Tag>
                        ))}
                      </Space>
                    )}
                  </Space>
                </Card>
              </List.Item>
            )}
          />
        </Space>
      </Content>
    </Layout>
  );
};

export default VideoManagement;
