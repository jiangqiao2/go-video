import React, { useEffect, useState } from 'react';
import { Layout, Button, Typography, Spin, Empty, Space } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import apiService from '@/services/api';
import { VideoDetail } from '@/types/api';
import VideoPlayer from '@/components/common/VideoPlayer';

const { Content } = Layout;
const { Title } = Typography;

const Watch: React.FC = () => {
  const navigate = useNavigate();
  const { video_uuid } = useParams();
  const [loading, setLoading] = useState(true);
  const [video, setVideo] = useState<VideoDetail | null>(null);

  useEffect(() => {
    const run = async () => {
      setLoading(true);
      try {
        const res = await apiService.listPublicVideos({ page: 1, size: 100, status: 'Published' });
        const found = res.videos.find(v => v.video_uuid === video_uuid);
        setVideo(found || null);
      } finally {
        setLoading(false);
      }
    };
    run();
  }, [video_uuid]);

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
