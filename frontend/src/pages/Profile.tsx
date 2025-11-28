import React, { useEffect, useState } from 'react';
import { Layout, Typography, Button, Avatar, Tabs, Row, Col, Tag, App } from 'antd';
import { UserOutlined, MessageOutlined, EllipsisOutlined, PlusOutlined, CheckOutlined, EditOutlined } from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { UserProfile, VideoDetail } from '@/types/api';
import apiService from '@/services/api';
import VideoCard from '@/components/common/VideoCard';
import { useAuthStore } from '@/store/auth';

const { Content } = Layout;
const { Title, Text, Paragraph } = Typography;

const Profile: React.FC = () => {
    const { user_uuid } = useParams();
    const { message } = App.useApp();
    const navigate = useNavigate();
    const currentUserUuid = useAuthStore((state) => state.user?.user_uuid);
    const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

    const [loading, setLoading] = useState(true);
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [videos, setVideos] = useState<VideoDetail[]>([]);
    const [following, setFollowing] = useState(false);
    const [followLoading, setFollowLoading] = useState(false);

    // 判断是否是查看自己的主页
    const isOwnProfile = Boolean(isAuthenticated && currentUserUuid && currentUserUuid === user_uuid);

    useEffect(() => {
        const run = async () => {
            if (!user_uuid) return;
            setLoading(true);
            try {
                const profileData = await apiService.getUserProfile(user_uuid);
                setProfile(profileData);
                setFollowing(!!profileData.is_followed);
            } catch (error) {
                try {
                    const basic = await apiService.getUserBasicInfo(user_uuid);
                    let relation: any = null;
                    try { relation = await apiService.getUserRelation(user_uuid); } catch {}
                    setProfile({
                        ...basic,
                        follower_count: relation?.follower_count ?? 0,
                        following_count: relation?.following_count ?? 0,
                        is_followed: !!relation?.is_followed,
                    });
                    setFollowing(!!relation?.is_followed);
                } catch {
                    message.error('获取用户信息失败');
                }
            }
            try {
                const videosData = await apiService.listVideosByUser(user_uuid, { page: 1, size: 20 });
                setVideos(videosData.videos);
            } catch {}
            setLoading(false);
        };
        run();
    }, [user_uuid]);

    const handleFollow = async () => {
        if (!isAuthenticated) {
            message.warning('请先登录');
            navigate('/login');
            return;
        }
        if (!profile || !user_uuid) return;

        setFollowLoading(true);
        try {
            if (following) {
                await apiService.unfollowUser(user_uuid);
            } else {
                await apiService.followUser(user_uuid);
            }
            const r = await apiService.getUserRelation(user_uuid);
            setFollowing(!!r.is_followed);
            message.success(r.is_followed ? '关注成功' : '已取消关注');
            setProfile(prev => prev ? ({
                ...prev,
                follower_count: r.follower_count,
                following_count: r.following_count,
            }) : prev);
        } catch (error) {
            console.error(error);
            message.error('操作失败');
        } finally {
            setFollowLoading(false);
        }
    };

    const handleEditProfile = () => {
        message.info('编辑资料功能开发中...');
        // TODO: 导航到编辑资料页面
        // navigate('/settings/profile');
    };

    if (loading) {
        return <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}><Content style={{ padding: 50, textAlign: 'center' }}>Loading...</Content></Layout>;
    }

    if (!profile) {
        return <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}><Content style={{ padding: 50, textAlign: 'center' }}>User not found</Content></Layout>;
    }

    return (
        <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}>
            {/* Banner */}
            <div style={{
                height: 200,
                backgroundImage: `url(${profile.cover_url || 'https://picsum.photos/1200/200'})`,
                backgroundSize: 'cover',
                backgroundPosition: 'center',
                position: 'relative'
            }}>
                <div style={{
                    position: 'absolute',
                    bottom: 0,
                    left: 0,
                    right: 0,
                    height: '60%',
                    background: 'linear-gradient(to top, rgba(0,0,0,0.5), transparent)'
                }} />
            </div>

            <Content style={{ maxWidth: 1200, margin: '0 auto', width: '100%', padding: '0 16px' }}>
                {/* User Info Header */}
                <div style={{
                    position: 'relative',
                    marginTop: -20,
                    padding: '0 24px 24px',
                    background: 'transparent',
                    display: 'flex',
                    alignItems: 'flex-end',
                    gap: 24
                }}>
                    {/* Avatar */}
                    <div style={{ position: 'relative' }}>
                        <Avatar
                            size={84}
                            src={profile.avatar_url}
                            icon={<UserOutlined />}
                            style={{
                                border: '4px solid #fff',
                                boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
                            }}
                        />
                    </div>

                    {/* Info */}
                    <div style={{ flex: 1, paddingBottom: 12, color: '#fff', textShadow: '0 1px 2px rgba(0,0,0,0.5)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                            <Title level={3} style={{ margin: 0, color: '#fff' }}>{profile.nickname || profile.account}</Title>
                            <Tag color="#f50">Lv6</Tag>
                            <Tag color="#ff4d4f">年度大会员</Tag>
                        </div>
                        <Paragraph style={{ margin: '8px 0 0', color: 'rgba(255,255,255,0.9)', maxWidth: 600 }} ellipsis={{ rows: 2 }}>
                            {profile.description || '这个人很懒，什么都没留下'}
                        </Paragraph>
                    </div>

                    {/* Actions - 区分自己和他人 */}
                    <div style={{ paddingBottom: 12, display: 'flex', gap: 12 }}>
                        {isOwnProfile ? (
                            // 自己的主页 - 显示编辑资料
                            <>
                                <Button
                                    type="primary"
                                    icon={<EditOutlined />}
                                    size="large"
                                    onClick={handleEditProfile}
                                    style={{ width: 120, borderRadius: 6 }}
                                >
                                    编辑资料
                                </Button>
                                <Button
                                    icon={<EllipsisOutlined />}
                                    size="large"
                                    ghost
                                    style={{ color: '#fff', borderColor: 'rgba(255,255,255,0.6)' }}
                                />
                            </>
                        ) : (
                            // 别人的主页 - 显示关注按钮
                            <>
                                <Button
                                    type={following ? 'default' : 'primary'}
                                    icon={following ? <CheckOutlined /> : <PlusOutlined />}
                                    size="large"
                                    onClick={handleFollow}
                                    loading={followLoading}
                                    style={{ width: 120, borderRadius: 6 }}
                                >
                                    {following ? '已关注' : '关注'}
                                </Button>
                                <Button
                                    icon={<MessageOutlined />}
                                    size="large"
                                    ghost
                                    style={{ color: '#fff', borderColor: 'rgba(255,255,255,0.6)' }}
                                >
                                    发消息
                                </Button>
                                <Button
                                    icon={<EllipsisOutlined />}
                                    size="large"
                                    ghost
                                    style={{ color: '#fff', borderColor: 'rgba(255,255,255,0.6)' }}
                                />
                            </>
                        )}
                    </div>
                </div>

                {/* Stats Bar */}
                <div style={{ background: '#fff', borderRadius: '0 0 12px 12px', padding: '16px 24px', display: 'flex', gap: 40, boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
                    <div style={{ textAlign: 'center' }}>
                        <Text type="secondary">关注数</Text>
                        <div style={{ fontSize: 18, fontWeight: 'bold' }}>{profile.following_count}</div>
                    </div>
                    <div style={{ textAlign: 'center' }}>
                        <Text type="secondary">粉丝数</Text>
                        <div style={{ fontSize: 18, fontWeight: 'bold' }}>{profile.follower_count}</div>
                    </div>
                    <div style={{ textAlign: 'center' }}>
                        <Text type="secondary">获赞数</Text>
                        <div style={{ fontSize: 18, fontWeight: 'bold' }}>999+</div>
                    </div>
                    <div style={{ textAlign: 'center' }}>
                        <Text type="secondary">播放数</Text>
                        <div style={{ fontSize: 18, fontWeight: 'bold' }}>1.2w</div>
                    </div>
                </div>

                {/* Content Tabs */}
                <div style={{ marginTop: 24, background: '#fff', borderRadius: 12, padding: 24, minHeight: 400 }}>
                    <Tabs defaultActiveKey="1" items={[
                        {
                            key: '1',
                            label: '主页',
                            children: (
                                <Row gutter={[24, 24]}>
                                    {videos.map(video => (
                                        <Col key={video.video_uuid} xs={24} sm={12} md={8} lg={6}>
                                            <VideoCard
                                                video={video}
                                                onClick={(v) => window.location.href = `/watch/${v.video_uuid}`}
                                                uploaderName={profile.nickname || profile.account}
                                                uploaderAvatar={profile.avatar_url}
                                            />
                                        </Col>
                                    ))}
                                    {videos.length === 0 && <EmptyState />}
                                </Row>
                            )
                        },
                        {
                            key: '2',
                            label: '动态',
                            children: <EmptyState />
                        },
                        {
                            key: '3',
                            label: '投稿',
                            children: (
                                <Row gutter={[24, 24]}>
                                    {videos.map(video => (
                                        <Col key={video.video_uuid} xs={24} sm={12} md={8} lg={6}>
                                            <VideoCard
                                                video={video}
                                                onClick={(v) => window.location.href = `/watch/${v.video_uuid}`}
                                                uploaderName={profile.nickname || profile.account}
                                                uploaderAvatar={profile.avatar_url}
                                            />
                                        </Col>
                                    ))}
                                    {videos.length === 0 && <EmptyState />}
                                </Row>
                            )
                        },
                        {
                            key: '4',
                            label: '合集和列表',
                            children: <EmptyState />
                        }
                    ]} />
                </div>
            </Content>
        </Layout>
    );
};

const EmptyState = () => (
    <div style={{ padding: 40, textAlign: 'center', color: '#999' }}>
        暂无内容
    </div>
);

export default Profile;
