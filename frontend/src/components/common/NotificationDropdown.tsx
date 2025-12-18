import React, { useMemo } from 'react';
import { Typography, Button, Space } from 'antd';
import { BellOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useNotificationStore, mapNotificationCategory } from '@/store/notifications';
import type { NotificationItem } from '@/types/api';

const { Text } = Typography;

interface NotificationDropdownProps {
  onItemClick?: () => void;
}

/**
 * 顶部铃铛展开的小菜单：
 * - 只负责展示各类消息是否有未读，以及未读数量（>99 显示 99+）
 * - 点击不同入口跳转到对应的通知中心页面
 */
const NotificationDropdown: React.FC<NotificationDropdownProps> = ({ onItemClick }) => {
  const navigate = useNavigate();
  const { items, markAllAsRead } = useNotificationStore();

  // 统计各分类未读数量
  const counts = useMemo(() => {
    const result = {
      reply: 0,
      at: 0,
      likes: 0,
      system: 0,
      mine: 0,
    };
    items.forEach((n: NotificationItem) => {
      if (n.is_read) return;
      const cat = mapNotificationCategory(n.type);
      if (cat === 'reply' || cat === 'comment') {
        result.reply += 1;
      } else if (cat === 'mention') {
        result.at += 1;
      } else if (cat === 'like') {
        result.likes += 1;
      } else if (cat === 'system') {
        result.system += 1;
      } else {
        result.mine += 1;
      }
    });
    return result;
  }, [items]);

  const formatCount = (n: number) => {
    if (n <= 0) return '';
    if (n > 99) return '99+';
    return String(n);
  };

  const go = (path: string) => {
    navigate(path);
    onItemClick?.();
  };

  return (
    <div
      style={{
        width: 360,
        maxHeight: 420,
        background: '#fff',
        borderRadius: 12,
        boxShadow: '0 8px 30px rgba(0,0,0,0.15)',
        padding: 12,
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '4px 8px 8px',
        }}
      >
        <Space>
          <BellOutlined style={{ color: '#ff6b9d' }} />
          <Text strong>消息中心</Text>
        </Space>
        <Button type="link" size="small" onClick={() => markAllAsRead()}>
          全部已读
        </Button>
      </div>

      {/* 菜单项：只负责跳转，不在这里展开列表 */}
      <div style={{ padding: '4px 4px 0' }}>
        <Button
          type="text"
          block
          onClick={() => go('/notifications/reply')}
          style={{ height: 40, textAlign: 'left' }}
        >
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <span>回复我的</span>
            {counts.reply > 0 && (
              <span style={{ color: '#ff4d4f', fontSize: 12 }}>{formatCount(counts.reply)}</span>
            )}
          </Space>
        </Button>

        <Button
          type="text"
          block
          onClick={() => go('/notifications/at')}
          style={{ height: 40, textAlign: 'left' }}
        >
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <span>@我的</span>
            {counts.at > 0 && (
              <span style={{ color: '#ff4d4f', fontSize: 12 }}>{formatCount(counts.at)}</span>
            )}
          </Space>
        </Button>

        <Button
          type="text"
          block
          onClick={() => go('/notifications/likes')}
          style={{ height: 40, textAlign: 'left' }}
        >
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <span>收到的赞</span>
            {counts.likes > 0 && (
              <span style={{ color: '#ff4d4f', fontSize: 12 }}>{formatCount(counts.likes)}</span>
            )}
          </Space>
        </Button>

        <Button
          type="text"
          block
          onClick={() => go('/notifications/system')}
          style={{ height: 40, textAlign: 'left' }}
        >
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <span>系统消息</span>
            {counts.system > 0 && (
              <span style={{ color: '#ff4d4f', fontSize: 12 }}>{formatCount(counts.system)}</span>
            )}
          </Space>
        </Button>

        <Button
          type="text"
          block
          onClick={() => go('/notifications/mine')}
          style={{ height: 40, textAlign: 'left' }}
        >
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <span>我的消息</span>
            {counts.mine > 0 && (
              <span style={{ color: '#ff4d4f', fontSize: 12 }}>{formatCount(counts.mine)}</span>
            )}
          </Space>
        </Button>
      </div>
    </div>
  );
};

export default NotificationDropdown;
