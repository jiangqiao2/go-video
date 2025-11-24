import React, { useEffect, useMemo, useState } from 'react';
import { Card, Typography, Space } from 'antd';
import { PlayCircleOutlined, UserOutlined } from '@ant-design/icons';
import { VideoDetail } from '@/types/api';
import dayjs from 'dayjs';

const { Text, Title } = Typography;

interface VideoCardProps {
  video: VideoDetail;
  onClick: (video: VideoDetail) => void;
  uploaderName?: string;
  uploaderAvatar?: string;
}

const formatDuration = (seconds: number) => {
  if (!isFinite(seconds) || seconds <= 0) return '--:--';
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

const VideoCard: React.FC<VideoCardProps> = ({ video, onClick, uploaderName, uploaderAvatar }) => {
  const [durationText, setDurationText] = useState<string>('--:--');

  const videoSrc = useMemo(() => video.video_url || '', [video.video_url]);

  useEffect(() => {
    if (!videoSrc) {
      setDurationText('--:--');
      return;
    }
    const el = document.createElement('video');
    el.preload = 'metadata';
    el.src = videoSrc;
    const onLoaded = () => {
      setDurationText(formatDuration(el.duration));
    };
    const onError = () => {
      setDurationText('--:--');
    };
    el.addEventListener('loadedmetadata', onLoaded);
    el.addEventListener('error', onError);
    return () => {
      el.removeEventListener('loadedmetadata', onLoaded);
      el.removeEventListener('error', onError);
      el.src = '';
    };
  }, [videoSrc]);

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
            {durationText}
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
          <Space size={6}>
            {uploaderAvatar ? (
              <img src={uploaderAvatar} alt={uploaderName || 'UP主'} style={{ width: 18, height: 18, borderRadius: '50%' }} />
            ) : (
              <UserOutlined />
            )}
            <Text type="secondary" style={{ fontSize: 12 }}>{uploaderName || 'UP主'}</Text>
          </Space>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {dayjs(video.published_at).format('MM-DD')}
          </Text>
        </Space>
        {video.tags?.length > 0 && (
          <div style={{ marginTop: 6 }}>
            <Space size={[4, 4]} wrap>
              {video.tags.slice(0, 3).map((tag) => (
                <span key={tag} style={{ backgroundColor: '#f1f2f3', color: '#6b6b6b', padding: '0 6px', borderRadius: 4, fontSize: 12 }}>
                  {tag}
                </span>
              ))}
            </Space>
          </div>
        )}
      </div>
    </Card>
  );
};

export default VideoCard;
