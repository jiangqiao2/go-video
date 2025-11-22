import React from 'react';
import { Card, Typography, Space } from 'antd';
import { PlayCircleOutlined, UserOutlined } from '@ant-design/icons';
import { VideoDetail } from '@/types/api';
import dayjs from 'dayjs';

const { Text, Title } = Typography;

interface VideoCardProps {
  video: VideoDetail;
  onClick: (video: VideoDetail) => void;
}

const VideoCard: React.FC<VideoCardProps> = ({ video, onClick }) => {
  // 格式化时长 (目前后端没返回时长，先mock或留空)
  

  return (
    <Card
      hoverable
      style={{ width: '100%', borderRadius: 8, overflow: 'hidden', border: 'none', boxShadow: 'none' }}
      bodyStyle={{ padding: '10px 0 0 0' }}
      cover={
        <div style={{ position: 'relative', paddingTop: '56.25%', backgroundColor: '#f0f0f0', borderRadius: 6, overflow: 'hidden' }}>
          {video.cover_url ? (
            <img
              alt={video.title}
              src={video.cover_url}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: '100%',
                objectFit: 'cover',
              }}
            />
          ) : (
            <div
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#ccc',
              }}
            >
              <PlayCircleOutlined style={{ fontSize: 32 }} />
            </div>
          )}
          {/* 视频时长遮罩 */}
          <div
            style={{
              position: 'absolute',
              bottom: 6,
              right: 6,
              backgroundColor: 'rgba(0,0,0,0.6)',
              color: '#fff',
              padding: '0 4px',
              borderRadius: 4,
              fontSize: 12,
            }}
          >
            {/* TODO: Replace with actual duration when available */}
            03:24
          </div>
        </div>
      }
      onClick={() => onClick(video)}
    >
      <div style={{ padding: '0 4px' }}>
        <Title
          level={5}
          ellipsis={{ rows: 2 }}
          style={{ marginBottom: 4, fontSize: 15, lineHeight: '22px', height: 44 }}
        >
          {video.title}
        </Title>
        <Space align="center" style={{ width: '100%', justifyContent: 'space-between', color: '#9499a0', fontSize: 12 }}>
          <Space size={4}>
            <UserOutlined />
            <Text type="secondary" style={{ fontSize: 12 }}>UP主</Text>
          </Space>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {dayjs(video.published_at).format('MM-DD')}
          </Text>
        </Space>
      </div>
    </Card>
  );
};

export default VideoCard;
