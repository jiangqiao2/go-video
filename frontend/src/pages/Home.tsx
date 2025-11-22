import React, { useCallback, useEffect, useState } from 'react';
import { Layout, Row, Col, Button, Space, Avatar, Input, message, Modal, Empty, Spin } from 'antd';
import { UploadOutlined, UserOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import apiService from '@/services/api';
import { VideoDetail } from '@/types/api';
import { useAuthStore } from '@/store/auth';
import VideoCard from '@/components/common/VideoCard';
import VideoPlayer from '@/components/common/VideoPlayer';

const { Header, Content } = Layout;

const Home: React.FC = () => {
    const navigate = useNavigate();
    const { user, logout } = useAuthStore();
    const [videos, setVideos] = useState<VideoDetail[]>([]);
    const [loading, setLoading] = useState(false);
    const [previewVideo, setPreviewVideo] = useState<VideoDetail | null>(null);
    const [previewVisible, setPreviewVisible] = useState(false);

    const fetchVideos = useCallback(async () => {
        setLoading(true);
        try {
            // 目前复用 listUserVideos，后续应改为 listAllVideos 或 feed 流接口
            const response = await apiService.listUserVideos({
                page: 1,
                size: 20,
                status: 'Published', // 只展示已发布的视频
            });
            setVideos(response.videos);
        } catch (error: any) {
            console.error('加载视频列表失败', error);
            message.error('加载视频列表失败');
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchVideos();
    }, [fetchVideos]);

    const handleLogout = () => {
        logout();
        message.success('已退出登录');
        navigate('/login');
    };

    const handleVideoClick = (video: VideoDetail) => {
        setPreviewVideo(video);
        setPreviewVisible(true);
    };

    return (
        <Layout style={{ minHeight: '100vh', backgroundColor: '#fff' }}>
            {/* 顶部导航栏 */}
            <Header
                style={{
                    backgroundColor: '#fff',
                    boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '0 24px',
                    position: 'sticky',
                    top: 0,
                    zIndex: 100,
                    height: 64,
                }}
            >
                {/* Logo / Brand */}
                <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }} onClick={() => window.location.reload()}>
                    <div style={{
                        width: 40,
                        height: 40,
                        backgroundColor: '#fb7299',
                        borderRadius: 8,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        marginRight: 12,
                        color: '#fff',
                        fontWeight: 'bold',
                        fontSize: 18
                    }}>
                        Go
                    </div>
                    <span style={{ fontSize: 18, fontWeight: 600, color: '#18191c' }}>GoVideo</span>
                </div>

                {/* Search Bar */}
                <div style={{ flex: 1, maxWidth: 500, margin: '0 24px' }}>
                    <Input
                        placeholder="搜索视频..."
                        prefix={<SearchOutlined style={{ color: '#ccc' }} />}
                        style={{ borderRadius: 20, backgroundColor: '#f1f2f3', border: 'none', padding: '6px 12px' }}
                    />
                </div>

                {/* User Actions */}
                <Space size="middle">
                    <Button
                        type="primary"
                        icon={<UploadOutlined />}
                        style={{ backgroundColor: '#fb7299', borderColor: '#fb7299', borderRadius: 6 }}
                        onClick={() => navigate('/upload')}
                    >
                        投稿
                    </Button>

                    {user ? (
                        <Space>
                            <Button type="text" icon={<ReloadOutlined />} onClick={fetchVideos} />
                            <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }} onClick={() => navigate('/videos')}>
                                <Avatar style={{ backgroundColor: '#00a1d6' }} icon={<UserOutlined />} />
                                <span style={{ marginLeft: 8, fontSize: 14, color: '#18191c' }}>{user.account}</span>
                            </div>
                            <Button type="link" onClick={handleLogout} style={{ color: '#9499a0' }}>退出</Button>
                        </Space>
                    ) : (
                        <Button type="primary" ghost onClick={() => navigate('/login')}>登录</Button>
                    )}
                </Space>
            </Header>

            {/* Main Content */}
            <Content style={{ padding: '24px', maxWidth: 1600, margin: '0 auto', width: '100%' }}>
                {loading && videos.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: 50 }}>
                        <Spin size="large" />
                    </div>
                ) : (
                    <>
                        {videos.length > 0 ? (
                            <Row gutter={[20, 24]}>
                                {videos.map((video) => (
                                    <Col xs={24} sm={12} md={8} lg={6} xl={4} key={video.video_uuid}>
                                        <VideoCard video={video} onClick={handleVideoClick} />
                                    </Col>
                                ))}
                            </Row>
                        ) : (
                            <Empty description="暂无视频，快去投稿吧" style={{ marginTop: 100 }} />
                        )}
                    </>
                )}
            </Content>

            {/* Video Player Modal */}
            <Modal
                open={previewVisible}
                onCancel={() => setPreviewVisible(false)}
                footer={null}
                width={1000}
                destroyOnClose
                centered
                title={previewVideo?.title}
                bodyStyle={{ padding: 0, backgroundColor: '#000' }}
            >
                {previewVideo && previewVideo.video_url && (
                    <VideoPlayer src={previewVideo.video_url} autoPlay />
                )}
            </Modal>
        </Layout>
    );
};

export default Home;
