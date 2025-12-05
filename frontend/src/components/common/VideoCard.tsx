import React, { useMemo, useState } from 'react';
import { Typography, Space } from 'antd';
import { PlayCircleOutlined, UserOutlined } from '@ant-design/icons';
import { VideoDetail } from '@/types/api';
import { formatPublishedTime } from '@/utils/date';

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
  const [isHovered, setIsHovered] = useState(false);
  const durationText = useMemo(
    () => formatDuration(video.duration_seconds ?? 0),
    [video.duration_seconds],
  );

  return (
    <div
      className="hover-lift"
      style={{
        width: '100%',
        borderRadius: 16,
        overflow: 'hidden',
        background: 'rgba(255, 255, 255, 0.9)',
        backdropFilter: 'blur(10px)',
        WebkitBackdropFilter: 'blur(10px)',
        border: '1px solid rgba(255, 255, 255, 0.5)',
        boxShadow: '0 4px 20px rgba(0, 0, 0, 0.08)',
        cursor: 'pointer',
        transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
      }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={() => onClick(video)}
    >
      {/* Video Cover */}
      <div style={{
        position: 'relative',
        paddingTop: '56.25%',
        backgroundColor: '#f0f0f0',
        overflow: 'hidden',
      }}>
        {video.cover_url ? (
          <>
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
                transition: 'transform 0.4s ease',
                transform: isHovered ? 'scale(1.1)' : 'scale(1)',
              }}
            />
            {/* 悬停遮罩 */}
            <div style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: 'linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,0.6) 100%)',
              opacity: isHovered ? 1 : 0.7,
              transition: 'opacity 0.3s ease',
            }} />
          </>
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
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              color: '#fff',
            }}
          >
            <PlayCircleOutlined style={{ fontSize: 48, opacity: 0.6 }} />
          </div>
        )}

        {/* 播放按钮 - 悬停时显示 */}
        <div style={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: `translate(-50%, -50%) scale(${isHovered ? 1 : 0.8})`,
          opacity: isHovered ? 1 : 0,
          transition: 'all 0.3s ease',
          pointerEvents: 'none',
        }}>
          <div style={{
            width: 64,
            height: 64,
            borderRadius: '50%',
            background: 'rgba(255, 255, 255, 0.95)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 8px 24px rgba(0, 0, 0, 0.3)',
          }}>
            <PlayCircleOutlined style={{
              fontSize: 36,
              color: '#667eea',
            }} />
          </div>
        </div>

        {/* 视频时长标签 */}
        <div
          style={{
            position: 'absolute',
            bottom: 10,
            right: 10,
            background: 'rgba(0, 0, 0, 0.75)',
            backdropFilter: 'blur(10px)',
            color: '#fff',
            padding: '4px 8px',
            borderRadius: 6,
            fontSize: 12,
            fontWeight: 600,
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)',
          }}
        >
          {durationText}
        </div>

      </div>

      {/* Card Content */}
      <div style={{ padding: '14px 16px' }}>
        <Title
          level={5}
          ellipsis={{ rows: 2 }}
          style={{
            marginBottom: 10,
            fontSize: 15,
            lineHeight: '22px',
            height: 44,
            fontWeight: 600,
            color: '#18191c',
          }}
        >
          {video.title}
        </Title>

        {/* 上传者信息 */}
        <Space align="center" style={{ width: '100%', justifyContent: 'space-between', marginBottom: 10 }}>
          <div
            onClick={(e) => {
              e.stopPropagation();
              if (video.user_uuid) {
                window.location.href = `/user/${video.user_uuid}`;
              }
            }}
            style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}
          >
            {uploaderAvatar ? (
              <img
                src={uploaderAvatar}
                alt={uploaderName || 'UP主'}
                style={{
                  width: 24,
                  height: 24,
                  borderRadius: '50%',
                  objectFit: 'cover',
                  border: '2px solid rgba(102, 126, 234, 0.2)',
                }}
                onError={(e) => {
                  const s = e.currentTarget.src;
                  const fixed = s.replace(/(\/storage\/image\/)(?:storage\/image\/)+/g, '$1');
                  if (fixed !== s) {
                    e.currentTarget.src = fixed;
                  }
                }}
              />
            ) : (
              <div style={{
                width: 24,
                height: 24,
                borderRadius: '50%',
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#fff',
                fontSize: 12,
              }}>
                <UserOutlined />
              </div>
            )}
            <Text
              type="secondary"
              style={{ fontSize: 13, fontWeight: 500 }}
              ellipsis
            >
              {uploaderName || 'UP主'}
            </Text>
          </div>

          <Text type="secondary" style={{ fontSize: 12 }}>
            {formatPublishedTime(video.published_at, { fallbackFormat: 'YYYY-MM-DD' })}
          </Text>
        </Space>

        {/* 标签 */}
        {video.tags?.length > 0 && (
          <div style={{ marginTop: 8 }}>
            <Space size={[6, 6]} wrap>
              {video.tags.slice(0, 3).map((tag) => (
                <span
                  key={tag}
                  style={{
                    background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                    color: '#667eea',
                    padding: '2px 10px',
                    borderRadius: 12,
                    fontSize: 12,
                    fontWeight: 500,
                    border: '1px solid rgba(102, 126, 234, 0.2)',
                    transition: 'all 0.3s ease',
                  }}
                  className="hover-scale"
                >
                  #{tag}
                </span>
              ))}
            </Space>
          </div>
        )}
      </div>

      {/* 底部渐变装饰 */}
      <div style={{
        height: 3,
        background: 'linear-gradient(90deg, #667eea 0%, #764ba2 50%, #f093fb 100%)',
        backgroundSize: '200% 100%',
        opacity: isHovered ? 1 : 0,
        transition: 'opacity 0.3s ease',
        animation: isHovered ? 'gradient-shift 3s ease infinite' : 'none',
      }} />
    </div>
  );
};

export default VideoCard;
