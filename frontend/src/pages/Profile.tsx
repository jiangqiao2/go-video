import React, { useEffect, useState } from 'react';
import { Layout, Typography, Button, Avatar, Tabs, Row, Col, message, Tag } from 'antd';
import { UserOutlined, MessageOutlined, EllipsisOutlined, PlusOutlined, CheckOutlined } from '@ant-design/icons';
import { useParams } from 'react-router-dom';
import { UserProfile, VideoDetail } from '@/types/api';
import apiService from '@/services/api';
import VideoCard from '@/components/common/VideoCard';

const { Content } = Layout;
const { Title, Text, Paragraph } = Typography;

const Profile: React.FC = () => {
    const { user_uuid } = useParams();
    const [loading, setLoading] = useState(true);
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [videos, setVideos] = useState<VideoDetail[]>([]);
    const [following, setFollowing] = useState(false);

    useEffect(() => {
        const fetchProfile = async () => {
            if (!user_uuid) return;
            setLoading(true);
            try {
                // 并行获取用户信息和视频列表
                // 注意：如果后端尚未实现 listVideosByUser，这里可能会报错，可以暂时用 listPublicVideos 代替测试
                const [profileData, videosData] = await Promise.all([
                    apiService.getUserProfile(user_uuid),
                    apiService.listVideosByUser(user_uuid, { page: 1, size: 20 })
                ]);

                setProfile(profileData);
                setFollowing(profileData.is_followed);
                setVideos(videosData.videos);
            } catch (error) {
                console.error(error);
                // 如果是 404，可能是用户不存在
                message.error('获取用户信息失败');
            } finally {
                setLoading(false);
            }
        };

        fetchProfile();
    }, [user_uuid]);

    const handleFollow = async () => {
        if (!profile || !user_uuid) return;
        try {
            if (following) {
                await apiService.unfollowUser(user_uuid);
            } else {
                await apiService.followUser(user_uuid);
            }
            setFollowing(!following);
            message.success(following ? '已取消关注' : '关注成功');

            // 更新本地粉丝数显示
            setProfile(prev => prev ? ({
                ...prev,
                follower_count: following ? prev.follower_count - 1 : prev.follower_count + 1
            }) : null);
        } catch (error) {
            console.error(error);
            message.error('操作失败');
        }
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
                backgroundImage: `url(${profile.cover_url})`,
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
                    background: 'transparent', // Make it blend or use glassmorphism if desired
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
                        {/* Verified Badge could go here */}
                    </div>

                    {/* Info */}
                    <div style={{ flex: 1, paddingBottom: 12, color: '#fff', textShadow: '0 1px 2px rgba(0,0,0,0.5)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                            <Title level={3} style={{ margin: 0, color: '#fff' }}>{profile.nickname || profile.account}</Title>
                            <Tag color="#f50">Lv6</Tag>
                            <Tag color="#ff4d4f">年度大会员</Tag>
                        </div>
                        <Paragraph style={{ margin: '8px 0 0', color: 'rgba(255,255,255,0.9)', maxWidth: 600 }} ellipsis={{ rows: 2 }}>
                            {profile.description}
                        </Paragraph>
                    </div>

                    {/* Actions */}
                    <div style={{ paddingBottom: 12, display: 'flex', gap: 12 }}>
                        <Button
                            type={following ? 'default' : 'primary'}
                            icon={following ? <CheckOutlined /> : <PlusOutlined />}
                            size="large"
                            onClick={handleFollow}
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
                            children: <EmptyState />
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
