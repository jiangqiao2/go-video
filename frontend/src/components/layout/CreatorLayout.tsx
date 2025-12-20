import React, { useMemo, useState } from 'react';
import { Layout, Menu, Button, Avatar, Space, Typography, Dropdown, Badge } from 'antd';
import {
  HomeOutlined,
  FileTextOutlined,
  UploadOutlined,
  UserOutlined,
  AppstoreOutlined,
  LogoutOutlined,
  SettingOutlined,
  BellOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/auth';
import NotificationDropdown from '@/components/common/NotificationDropdown';
import { useNotificationStore } from '@/store/notifications';
import { useNotificationStream } from '@/hooks/useNotificationStream';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

interface CreatorLayoutProps {
  children: React.ReactNode;
  activeKey?: 'home' | 'upload' | 'videos';
}

const CreatorLayout: React.FC<CreatorLayoutProps> = ({ children, activeKey }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const [collapsed, setCollapsed] = useState(false);
  const [notificationOpen, setNotificationOpen] = useState(false);
  const { unreadCount, fetchNotifications } = useNotificationStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const userMenu = {
    items: [
      {
        key: 'profile',
        icon: <UserOutlined />,
        label: '个人中心',
      },
      {
        key: 'settings',
        icon: <SettingOutlined />,
        label: '创作设置',
      },
      {
        type: 'divider',
      },
      {
        key: 'logout',
        icon: <LogoutOutlined />,
        label: '退出登录',
        onClick: handleLogout,
      },
    ] as any,
  };

  const menuItems = [
    {
      key: 'main',
      icon: <HomeOutlined />,
      label: '主站',
      onClick: () => window.open('/', '_blank'),
    },
    {
      key: 'content',
      icon: <AppstoreOutlined />,
      label: '内容管理',
      children: [
        {
          key: 'videos',
          label: '稿件管理',
          icon: <FileTextOutlined />,
          onClick: () => navigate('/videos'),
        },
      ],
    },
  ];

  const currentKey = useMemo(() => {
    if (activeKey) return activeKey;
    if (location.pathname.startsWith('/videos')) return 'videos';
    return 'main';
  }, [activeKey, location.pathname]);

  const openKeys = currentKey === 'videos' ? ['content'] : [];

  // 在创作中心布局中同样保持通知 SSE 流；NotificationStream 是单例，不会重复建立连接。
  useNotificationStream(!!user);

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={(value) => setCollapsed(value)}
        width={240}
        style={{
          background: '#fff',
          borderRight: '1px solid #f0f0f0',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 100,
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #f0f0f0',
            padding: '0 12px',
          }}
        >
          <Button
            type="primary"
            icon={<UploadOutlined />}
            size="large"
            onClick={() => navigate('/upload')}
            style={{
              width: collapsed ? 48 : '100%',
              padding: collapsed ? 0 : undefined,
              transition: 'all 0.2s',
              overflow: 'hidden',
            }}
          >
            {collapsed ? '' : '发布投稿'}
          </Button>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[currentKey]}
          defaultOpenKeys={openKeys}
          style={{ borderRight: 0, paddingTop: 12 }}
          items={menuItems}
        />
      </Sider>
      <Layout style={{ marginLeft: collapsed ? 80 : 240, transition: 'all 0.2s' }}>
        <Header
          style={{
            padding: '0 24px',
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid #f0f0f0',
            position: 'sticky',
            top: 0,
            zIndex: 99,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Text strong style={{ fontSize: 18, color: '#00a1d6' }}>
              创作中心
            </Text>
            <Text type="secondary" style={{ marginLeft: 12, fontSize: 12 }}>
              创作服务平台
            </Text>
          </div>
          <Space size="large">
            {user && (
              <>
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
                    style={{
                      width: 40,
                      height: 40,
                      borderRadius: 20,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <Badge
                      count={unreadCount}
                      size="small"
                      overflowCount={99}
                      offset={[-2, 2]}
                    >
                      <BellOutlined />
                    </Badge>
                  </Button>
                </Dropdown>
                <Dropdown menu={userMenu} placement="bottomRight">
                  <div
                    style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}
                    onClick={() => navigate(`/user/${user.user_uuid}`)}
                  >
                    <Avatar src={user.avatar_url} icon={<UserOutlined />} />
                  </div>
                </Dropdown>
              </>
            )}
          </Space>
        </Header>
        <Content style={{ margin: '24px 24px 0', overflow: 'initial' }}>{children}</Content>
      </Layout>
    </Layout>
  );
};

export default CreatorLayout;
