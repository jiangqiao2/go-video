import React, { useEffect, useState } from 'react';
import { Layout, Button, Typography, Spin, Empty, Space, Avatar, message, Tag, Dropdown } from 'antd';
import { ArrowLeftOutlined, UserOutlined, PlusOutlined, CheckOutlined, MenuOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import apiService from '@/services/api';
import { VideoDetail } from '@/types/api';
import VideoPlayer from '@/components/common/VideoPlayer';
import { useAuthStore } from '@/store/auth';
import dayjs from 'dayjs';

const { Content } = Layout;
const { Title, Text, Paragraph } = Typography;

const Watch: React.FC = () => {
  const navigate = useNavigate();
  const { video_uuid } = useParams();
  const [loading, setLoading] = useState(true);
  const [video, setVideo] = useState<VideoDetail | null>(null);
  const [following, setFollowing] = useState(false);
  const [followLoading, setFollowLoading] = useState(false);
  const currentUserUuid = useAuthStore((s) => s.user?.user_uuid) || localStorage.getItem('user_uuid') || '';
  const isOwner = !!video?.user_uuid && video.user_uuid === currentUserUuid;

  useEffect(() => {
    const run = async () => {
      setLoading(true);
      try {
        const res = await apiService.listPublicVideos({ page: 1, size: 100, status: 'Published' });
        const found = res.videos.find(v => v.video_uuid === video_uuid);
        setVideo(found || null);

        if (found && found.user_uuid) {
          try {
            const profile = await apiService.getUserProfile(found.user_uuid);
            setFollowing(profile.is_followed);
          } catch (e) {
            console.error('Failed to fetch uploader profile', e);
          }
        }
      } finally {
        setLoading(false);
      }
    };
    run();
  }, [video_uuid]);

  const handleFollow = async () => {
    if (!video?.user_uuid) return;
    if (isOwner) {
      message.warning('不能关注自己');
      return;
    }
    setFollowLoading(true);
    try {
      if (following) {
        await apiService.unfollowUser(video.user_uuid);
      } else {
        await apiService.followUser(video.user_uuid);
      }
      setFollowing(!following);
      message.success(following ? '已取消关注' : '关注成功');
    } catch (error: any) {
      const msg = error?.response?.data?.message || '操作失败';
      if (typeof msg === 'string' && msg.includes('不能关注自己')) {
        message.warning('不能关注自己');
      } else {
        message.error(msg);
      }
    } finally {
      setFollowLoading(false);
    }
  };

  const handleUnfollow = async () => {
    if (!video?.user_uuid) return;
    setFollowLoading(true);
    try {
      await apiService.unfollowUser(video.user_uuid);
      setFollowing(false);
      message.success('已取消关注');
    } catch (error: any) {
      const msg = error?.response?.data?.message || '操作失败';
      message.error(msg);
    } finally {
      setFollowLoading(false);
    }
  };

  return (
    <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}>
      <Content style={{ padding: '24px 16px' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <Space size="middle" style={{ marginBottom: 16 }}>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/')}>返回首页</Button>
            <Title level={4} style={{ margin: 0 }}>{video?.title || '播放页面'}</Title>
          </Space>

          {loading ? (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 480 }}>
              <Spin size="large" />
            </div>
          ) : video && video.video_url ? (
            <>
              <div style={{
                position: 'relative',
                width: '100%',
                paddingTop: '56.25%',
                background: '#000',
                borderRadius: 12,
                overflow: 'hidden',
                boxShadow: '0 8px 24px rgba(0,0,0,0.15)'
              }}>
                <div style={{ position: 'absolute', inset: 0 }}>
                  <VideoPlayer src={video.video_url} autoPlay />
                </div>
              </div>

              {/* Uploader Info & Description */}
              <div style={{ marginTop: 24, background: '#fff', padding: 24, borderRadius: 12, boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 16, cursor: 'pointer' }} onClick={() => window.location.href = `/user/${video.user_uuid}`}>
                    <Avatar
                      size={48}
                      src={video.uploader_avatar_url}
                      icon={<UserOutlined />}
                    />
                    <div>
                      <Title level={5} style={{ margin: 0 }}>{video.uploader_account}</Title>
                      <Text type="secondary" style={{ fontSize: 12 }}>{dayjs(video.published_at).format('YYYY-MM-DD HH:mm')}</Text>
                    </div>
                  </div>

                  {!isOwner && (
                    following ? (
                      <Dropdown
                        trigger={['click']}
                        menu={{
                          items: [
                            { key: 'unfollow', label: '取消关注' },
                          ],
                          onClick: async ({ key }) => {
                            if (key === 'unfollow') {
                              await handleUnfollow();
                            }
                          },
                        }}
                      >
                        <Button type="default" icon={<MenuOutlined />} loading={followLoading}>
                          已关注
                        </Button>
                      </Dropdown>
                    ) : (
                      <Button
                        type="primary"
                        icon={<PlusOutlined />}
                        onClick={handleFollow}
                        loading={followLoading}
                      >
                        关注
                      </Button>
                    )
                  )}
                </div>

                <div style={{ paddingTop: 16, borderTop: '1px solid #f0f0f0' }}>
                  <Title level={5} style={{ fontSize: 16, marginBottom: 8 }}>{video.title}</Title>
                  <Paragraph style={{ color: '#666', whiteSpace: 'pre-wrap' }}>
                    {video.description || '暂无简介'}
                  </Paragraph>
                  <div style={{ marginTop: 12 }}>
                    <Space size={[8, 8]} wrap>
                      {video.tags?.map(tag => (
                        <Tag key={tag} color="blue">#{tag}</Tag>
                      ))}
                    </Space>
                  </div>
                </div>
              </div>
            </>
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 480 }}>
              <Empty description="暂无可播放的视频" />
            </div>
          )}
        </div>
      </Content>
    </Layout>
  );
};

export default Watch;
