import React, { useEffect, useMemo, useState } from 'react';
import {
  Layout,
  Input,
  List,
  Typography,
  Badge,
  Space,
  Button,
  Spin,
  Empty,
} from 'antd';
import {
  MessageOutlined,
  BellOutlined,
  LikeOutlined,
  NotificationOutlined,
  ArrowLeftOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useNotificationStore, mapNotificationCategory } from '@/store/notifications';
import type { NotificationItem as NotificationItemType } from '@/types/api';
import { formatPublishedTime } from '@/utils/date';

const { Header, Content } = Layout;
const { Text, Title, Paragraph } = Typography;

type TabKey = 'reply' | 'at' | 'likes' | 'system' | 'mine';

const TAB_KEYS: TabKey[] = ['reply', 'at', 'likes', 'system', 'mine'];

const TABS: { key: TabKey; label: string; icon: React.ReactNode }[] = [
  { key: 'reply', label: '回复我的', icon: <MessageOutlined /> },
  { key: 'at', label: '@我的', icon: <NotificationOutlined /> },
  { key: 'likes', label: '收到的赞', icon: <LikeOutlined /> },
  { key: 'system', label: '系统消息', icon: <NotificationOutlined /> },
  { key: 'mine', label: '我的消息', icon: <BellOutlined /> },
];

const Notifications: React.FC = () => {
  const navigate = useNavigate();
  const params = useParams<{ tab?: string }>();
  const routeTab = (params.tab || 'reply') as TabKey;

  const {
    items,
    loading,
    fetchNotifications,
    markAsRead,
    markAllAsRead,
  } = useNotificationStore();

  const [keyword, setKeyword] = useState('');
  const [selectedId, setSelectedId] = useState<number | null>(null);

  const activeTab: TabKey = TAB_KEYS.includes(routeTab) ? routeTab : 'reply';

  // 首次进入或切换 tab 时拉取通知列表
  useEffect(() => {
    fetchNotifications().catch(() => undefined);
  }, [fetchNotifications, activeTab]);

  // 非法 tab 时重定向到默认的 reply
  useEffect(() => {
    if (!TAB_KEYS.includes(routeTab)) {
      navigate('/notifications/reply', { replace: true });
    }
  }, [routeTab, navigate]);

  // 各类别未读计数，用于左侧菜单徽标
  const categoryCounts = useMemo(() => {
    const res = {
      reply: 0,
      at: 0,
      likes: 0,
      system: 0,
      mine: 0,
    };
    items.forEach((n: NotificationItemType) => {
      if (n.is_read) return;
      const cat = mapNotificationCategory(n.type);
      if (cat === 'reply' || cat === 'comment') {
        res.reply += 1;
      } else if (cat === 'mention') {
        res.at += 1;
      } else if (cat === 'like') {
        res.likes += 1;
      } else if (cat === 'system') {
        res.system += 1;
      } else {
        res.mine += 1;
      }
    });
    return res;
  }, [items]);

  // 根据当前 tab 和搜索关键字过滤列表
  const filteredList = useMemo(() => {
    const byTab = items.filter((n) => {
      const cat = mapNotificationCategory(n.type);
      switch (activeTab) {
        case 'reply':
          return cat === 'reply' || cat === 'comment';
        case 'at':
          return cat === 'mention';
        case 'likes':
          return cat === 'like';
        case 'system':
          return cat === 'system';
        case 'mine':
          return cat === 'other';
        default:
          return true;
      }
    });

    if (!keyword.trim()) return byTab;

    const kw = keyword.trim().toLowerCase();
    return byTab.filter((n) => {
      const t = (n.title || '').toLowerCase();
      const c = (n.content || '').toLowerCase();
      return t.includes(kw) || c.includes(kw);
    });
  }, [items, activeTab, keyword]);

  const selectedItem = useMemo(
    () => items.find((n) => n.id === selectedId) || null,
    [items, selectedId],
  );

  const handleMenuClick = (key: TabKey) => {
    if (key === activeTab) return;
    setSelectedId(null);
    navigate(`/notifications/${key}`);
  };

  const handleItemClick = (item: NotificationItemType) => {
    setSelectedId(item.id);
    if (!item.is_read) {
      markAsRead(item.id).catch(() => undefined);
    }
  };

  const handleAllRead = () => {
    markAllAsRead().catch(() => undefined);
  };

  const renderListItem = (item: NotificationItemType) => {
    const isActive = item.id === selectedId;
    const isUnread = !item.is_read;
    const content =
      item.content && item.content.length > 40
        ? `${item.content.slice(0, 40)}...`
        : item.content;

    return (
      <div
        key={item.id}
        onClick={() => handleItemClick(item)}
        style={{
          padding: '10px 12px',
          borderRadius: 8,
          cursor: 'pointer',
          marginBottom: 4,
          background: isActive ? 'rgba(0, 161, 214, 0.08)' : 'transparent',
          border: isActive
            ? '1px solid rgba(0, 161, 214, 0.25)'
            : '1px solid transparent',
          display: 'flex',
          alignItems: 'flex-start',
          gap: 8,
        }}
      >
        <Badge dot={isUnread} offset={[-2, 2]} color="#ff4d4f">
          <BellOutlined style={{ color: '#00a1d6' }} />
        </Badge>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontSize: 14,
              fontWeight: isUnread ? 600 : 500,
              color: '#18191c',
              marginBottom: 4,
            }}
          >
            {item.title || '系统通知'}
          </div>
          {content && (
            <div
              style={{
                fontSize: 12,
                color: '#9499a0',
                marginBottom: 4,
              }}
            >
              {content}
            </div>
          )}
          <div style={{ fontSize: 12, color: '#c0c4cc' }}>
            {formatPublishedTime(item.created_at)}
          </div>
        </div>
      </div>
    );
  };

  return (
    <Layout
      style={{
        minHeight: '100vh',
        background: '#e5f5f9',
      }}
    >
      {/* 顶部条，参考 B 站消息中心 */}
      <Header
        style={{
          background: '#ffffff',
          borderBottom: '1px solid #e5e9ef',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 32px',
          height: 64,
        }}
      >
        <Space size={16}>
          <Button
            type="text"
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/')}
          />
          <Title level={4} style={{ margin: 0 }}>
            消息中心
          </Title>
        </Space>
        <Space>
          <Text type="secondary">我的消息</Text>
        </Space>
      </Header>

      <Content
        style={{
          padding: '24px 40px',
          maxWidth: 1200,
          margin: '0 auto',
          width: '100%',
        }}
      >
        <div
          style={{
            display: 'flex',
            background: '#ffffff',
            borderRadius: 12,
            boxShadow: '0 4px 16px rgba(0,0,0,0.06)',
            minHeight: 520,
            overflow: 'hidden',
          }}
        >
          {/* 左侧分类菜单 */}
          <div
            style={{
              width: 220,
              borderRight: '1px solid #f0f0f0',
              padding: '18px 0',
              background: '#f7f9fc',
            }}
          >
            <div
              style={{
                padding: '0 24px 12px',
                fontSize: 12,
                color: '#9499a0',
              }}
            >
              最近消息
            </div>
            {TABS.map((tab) => {
              const count = categoryCounts[tab.key];
              const active = activeTab === tab.key;
              return (
                <div
                  key={tab.key}
                  onClick={() => handleMenuClick(tab.key)}
                  style={{
                    padding: '10px 24px',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    background: active ? '#ffffff' : 'transparent',
                    boxShadow: active ? 'inset 3px 0 0 #00a1d6' : 'none',
                    color: active ? '#00a1d6' : '#18191c',
                    fontWeight: active ? 600 : 500,
                  }}
                >
                  <Space>
                    {tab.icon}
                    <span>{tab.label}</span>
                  </Space>
                  {count > 0 && (
                    <Badge
                      count={count}
                      size="small"
                      overflowCount={99}
                      style={{ backgroundColor: '#ff4d4f' }}
                    />
                  )}
                </div>
              );
            })}
          </div>

          {/* 右侧主区域 */}
          <div
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              padding: '18px 20px',
            }}
          >
            {/* 搜索 + 全部已读 */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 16,
              }}
            >
              <Input.Search
                placeholder="搜索我的消息"
                allowClear
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                style={{ maxWidth: 360 }}
              />
              <Button type="link" onClick={handleAllRead}>
                全部已读
              </Button>
            </div>

            {/* 列表 + 详情 两栏布局 */}
            <div
              style={{
                flex: 1,
                display: 'flex',
                minHeight: 420,
              }}
            >
              {/* 左侧列表 */}
              <div
                style={{
                  width: 340,
                  borderRight: '1px solid #f0f0f0',
                  paddingRight: 16,
                  marginRight: 16,
                  overflowY: 'auto',
                }}
              >
                {loading ? (
                  <div
                    style={{
                      paddingTop: 80,
                      textAlign: 'center',
                    }}
                  >
                    <Spin />
                  </div>
                ) : filteredList.length === 0 ? (
                  <div
                    style={{
                      paddingTop: 80,
                      textAlign: 'center',
                    }}
                  >
                    <Empty description="还没有相关消息" />
                  </div>
                ) : (
                  <List
                    dataSource={filteredList}
                    renderItem={renderListItem}
                    split={false}
                  />
                )}
              </div>

              {/* 右侧详情 / 空白插画 */}
              <div
                style={{
                  flex: 1,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                {selectedItem ? (
                  <div style={{ maxWidth: 520 }}>
                    <Space size="small">
                      <Badge
                        dot={!selectedItem.is_read}
                        color="#ff4d4f"
                      />
                      <Text strong style={{ fontSize: 16 }}>
                        {selectedItem.title || '系统通知'}
                      </Text>
                    </Space>
                    <div
                      style={{
                        marginTop: 8,
                        marginBottom: 16,
                        fontSize: 12,
                        color: '#9499a0',
                      }}
                    >
                      {formatPublishedTime(selectedItem.created_at)}
                    </div>
                    <Paragraph
                      style={{
                        fontSize: 14,
                        lineHeight: 1.8,
                        whiteSpace: 'pre-wrap',
                      }}
                    >
                      {selectedItem.content || '暂无内容'}
                    </Paragraph>
                  </div>
                ) : (
                  <div style={{ textAlign: 'center', color: '#9499a0' }}>
                    <div
                      style={{
                        width: 260,
                        height: 180,
                        borderRadius: 12,
                        border: '1px dashed #dcdfe6',
                        margin: '0 auto 16px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#c0c4cc',
                        fontSize: 13,
                      }}
                    >
                      (´･ω･`) 这里空空如也
                    </div>
                    <div>快找小伙伴聊聊天吧 (ﾟ ∀。)ノ♡</div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </Content>
    </Layout>
  );
};

export default Notifications;
