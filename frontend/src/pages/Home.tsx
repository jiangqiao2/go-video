import React, { useCallback, useEffect, useState } from 'react';
import { Layout, Row, Col, Button, Space, Avatar, Input, Spin, Badge, App, Dropdown } from 'antd';
import { UploadOutlined, UserOutlined, SearchOutlined, ReloadOutlined, BellOutlined, FireOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import apiService from '@/services/api';
import { VideoDetail } from '@/types/api';
import { useAuthStore } from '@/store/auth';
import VideoCard from '@/components/common/VideoCard';
import NotificationDropdown from '@/components/common/NotificationDropdown';
import { useNotificationStore } from '@/store/notifications';
import { useNotificationStream } from '@/hooks/useNotificationStream';


const { Header, Content } = Layout;

const Home: React.FC = () => {
    const navigate = useNavigate();
    const { message } = App.useApp();
    const { user, logout, refreshUserInfo } = useAuthStore();
    const [videos, setVideos] = useState<VideoDetail[]>([]);
    const [loading, setLoading] = useState(false);
    const [notificationOpen, setNotificationOpen] = useState(false);

    const [mounted, setMounted] = useState(false);

    useEffect(() => {
        setMounted(true);
    }, []);

    const normalizeAvatarUrl = (url?: string) => {
        if (!url) return undefined;
        if (url.startsWith('http')) return url;
        const base = import.meta.env.VITE_ASSET_BASE || window.location.origin;
        return `${base}/${url.replace(/^\/+/, '')}`;
    };

    const fetchVideos = useCallback(async () => {
        setLoading(true);
        try {
            const response = await apiService.listPublicVideos({
                page: 1,
                size: 20,
            });
            const enriched = await apiService.attachUploaderBasicInfo(response.videos);
            setVideos(enriched);
        } catch (error: any) {
            console.error('加载视频列表失败', error);
            message.error('加载视频列表失败');
        } finally {
            setLoading(false);
        }
    }, [message]);

    const handleLoadVideos = useCallback(() => {
        fetchVideos();
    }, [fetchVideos]);

    useEffect(() => {
        handleLoadVideos();
    }, [handleLoadVideos]);

    useEffect(() => {
        if (user?.user_uuid) {
            refreshUserInfo().catch(() => { });
        }
    }, [user?.user_uuid, refreshUserInfo]);

    // 通知相关：在打开下拉时加载一次
    const { unreadCount, fetchNotifications } = useNotificationStore();

    const handleLogout = () => {
        logout();
        message.success('已退出登录');
        navigate('/login');
    };

    const handleVideoClick = (video: VideoDetail) => {
        window.open(`/watch/${video.video_uuid}`, '_blank');
    };

    // 建立通知 SSE 流：有新通知时自动刷新未读列表
    useNotificationStream(!!user);

    return (
        <Layout style={{ minHeight: '100vh', background: 'linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)' }}>
            {/* 装饰性背景网格 */}
            <div className="grid-background" style={{ opacity: 0.3 }} />

            {/* 顶部导航栏 - 玻璃态设计 */}
            <Header
                className={mounted ? 'fade-in' : ''}
                style={{
                    background: 'rgba(255, 255, 255, 0.8)',
                    backdropFilter: 'blur(20px)',
                    WebkitBackdropFilter: 'blur(20px)',
                    boxShadow: '0 4px 30px rgba(0, 0, 0, 0.1)',
                    borderBottom: '1px solid rgba(255, 255, 255, 0.5)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '0 32px',
                    position: 'sticky',
                    top: 0,
                    zIndex: 100,
                    height: 70,
                }}
            >
                {/* Logo / Brand */}
                <div
                    className="hover-scale"
                    style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}
                    onClick={() => window.location.reload()}
                >
                    <div style={{
                        width: 46,
                        height: 46,
                        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                        borderRadius: 12,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        marginRight: 12,
                        color: '#fff',
                        fontWeight: 'bold',
                        fontSize: 20,
                        boxShadow: '0 4px 15px rgba(102, 126, 234, 0.3)',
                        transition: 'all 0.3s ease',
                    }}>
                        Go
                    </div>
                    <span style={{
                        fontSize: 22,
                        fontWeight: 700,
                        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                        WebkitBackgroundClip: 'text',
                        WebkitTextFillColor: 'transparent',
                        backgroundClip: 'text',
                    }}>
                        GoVideo
                    </span>
                </div>

                {/* Search Bar */}
                <div style={{ flex: 1, maxWidth: 560, margin: '0 32px' }}>
                    <Input
                        placeholder="搜索你感兴趣的视频..."
                        prefix={<SearchOutlined style={{ color: '#9499a0', fontSize: 16 }} />}
                        suffix={
                            <FireOutlined style={{ color: '#ff6b9d', fontSize: 16 }} title="热门搜索" />
                        }
                        style={{
                            borderRadius: 24,
                            background: 'rgba(255, 255, 255, 0.9)',
                            border: '2px solid transparent',
                            padding: '10px 20px',
                            fontSize: 15,
                            transition: 'all 0.3s ease',
                        }}
                        onFocus={(e) => {
                            e.target.style.borderColor = '#667eea';
                            e.target.style.boxShadow = '0 0 0 3px rgba(102, 126, 234, 0.1)';
                            e.target.style.background = '#fff';
                        }}
                        onBlur={(e) => {
                            e.target.style.borderColor = 'transparent';
                            e.target.style.boxShadow = 'none';
                            e.target.style.background = 'rgba(255, 255, 255, 0.9)';
                        }}
                    />
                </div>

                {/* User Actions */}
                <Space size="middle">
                    <Button
                        type="primary"
                        icon={<UploadOutlined />}
                        className="gradient-button hover-lift"
                        style={{
                            borderRadius: 8,
                            height: 42,
                            paddingLeft: 24,
                            paddingRight: 24,
                            fontWeight: 600,
                            fontSize: 15,
                            border: 'none',
                            boxShadow: '0 4px 15px rgba(102, 126, 234, 0.3)',
                        }}
                        onClick={() => navigate('/upload')}
                    >
                        投稿
                    </Button>

                    {user ? (
                        <Space size="middle">
                            <Dropdown
                                trigger={['click']}
                                open={notificationOpen}
                                onOpenChange={async (open) => {
                                    setNotificationOpen(open);
                                    if (open) {
                                        await fetchNotifications();
                                    }
                                }}
                                dropdownRender={() => <NotificationDropdown />}
                                placement="bottomRight"
                            >
                                <Button
                                    type="text"
                                    className="hover-scale"
                                    style={{
                                        width: 42,
                                        height: 42,
                                        borderRadius: 10,
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                    }}
                                >
                                    <Badge
                                        count={unreadCount}
                                        size="small"
                                        offset={[-2, 2]}
                                        overflowCount={99}
                                    >
                                        <BellOutlined style={{ fontSize: 18 }} />
                                    </Badge>
                                </Button>
                            </Dropdown>
                            <Button
                                type="text"
                                icon={<ReloadOutlined style={{ fontSize: 18 }} />}
                                onClick={handleLoadVideos}
                                className="hover-scale"
                                style={{
                                    width: 42,
                                    height: 42,
                                    borderRadius: 10,
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                }}
                            />
                            <div
                                className="hover-scale"
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    cursor: 'pointer',
                                    padding: '6px 12px',
                                    borderRadius: 20,
                                    background: 'rgba(102, 126, 234, 0.1)',
                                    transition: 'all 0.3s ease',
                                }}
                                onClick={() => navigate(`/user/${user.user_uuid}`)}
                            >
                                <Avatar
                                    style={{
                                        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                                        boxShadow: '0 2px 8px rgba(102, 126, 234, 0.3)',
                                    }}
                                    src={user.avatar_url}
                                    icon={<UserOutlined />}
                                />
                                <span style={{ marginLeft: 10, fontSize: 15, fontWeight: 600, color: '#18191c' }}>
                                    {user.nickname || '个人中心'}
                                </span>
                            </div>
                            <Button
                                type="link"
                                onClick={handleLogout}
                                style={{ color: '#9499a0', fontSize: 14 }}
                            >
                                退出
                            </Button>
                        </Space>
                    ) : (
                        <Button
                            type="primary"
                            ghost
                            onClick={() => navigate('/login')}
                            className="hover-lift"
                            style={{
                                borderRadius: 8,
                                height: 42,
                                paddingLeft: 24,
                                paddingRight: 24,
                                fontWeight: 600,
                                borderWidth: 2,
                                borderColor: '#667eea',
                                color: '#667eea',
                            }}
                        >
                            登录
                        </Button>
                    )}
                </Space>
            </Header>

            {/* Main Content */}
            <Content
                className={mounted ? 'fade-in-up' : ''}
                style={{
                    padding: '32px',
                    maxWidth: 1600,
                    margin: '0 auto',
                    width: '100%',
                    animationDelay: '0.2s',
                }}
            >
                {loading && videos.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: 80 }}>
                        <Spin size="large" />
                        <p style={{ marginTop: 20, color: '#9499a0', fontSize: 15 }}>加载中...</p>
                    </div>
                ) : videos.length > 0 ? (
                    <>
                        <Row gutter={[24, 28]}>
                            {videos.map((video, index) => (
                                <Col
                                    xs={24}
                                    sm={12}
                                    md={8}
                                    lg={6}
                                    xl={4}
                                    key={video.video_uuid}
                                    className="fade-in"
                                    style={{
                                        animationDelay: `${index * 0.05}s`,
                                    }}
                                >
                                    <VideoCard
                                        video={video}
                                        onClick={handleVideoClick}
                                        uploaderName={video.uploader_account || user?.nickname || '创作者'}
                                        uploaderAvatar={normalizeAvatarUrl(video.uploader_avatar_url) || user?.avatar_url}
                                    />
                                </Col>
                            ))}
                        </Row>
                    </>
                ) : (
                    <div style={{
                        textAlign: 'center',
                        padding: '120px 20px',
                        background: 'rgba(255, 255, 255, 0.6)',
                        backdropFilter: 'blur(10px)',
                        borderRadius: 16,
                        marginTop: 40,
                    }}>
                        <p style={{ fontSize: 18, color: '#18191c', marginBottom: 12 }}>暂无视频，快去投稿吧！</p>
                        <Button
                            type="primary"
                            size="large"
                            icon={<UploadOutlined />}
                            onClick={() => navigate('/upload')}
                            className="gradient-button hover-lift"
                            style={{
                                height: 48,
                                paddingLeft: 28,
                                paddingRight: 28,
                                fontSize: 16,
                                fontWeight: 600,
                                borderRadius: 10,
                                border: 'none',
                            }}
                        >
                            立即投稿
                        </Button>
                    </div>
                )}
            </Content>


        </Layout>
    );
};

export default Home;
