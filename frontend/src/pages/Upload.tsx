import React from 'react';
import { Layout, Typography, Button, Space, Avatar, message } from 'antd';
import { UserOutlined, VideoCameraOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import VideoUpload from '@/components/upload/VideoUpload';
import { useAuthStore } from '@/store/auth';

const { Content, Header } = Layout;
const { Title, Text } = Typography;

const Upload: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    message.success('已退出登录');
    navigate('/login', { replace: true });
  };

  const handleGoLogin = () => {
    navigate('/login');
  };

  const handleGoManagement = () => {
    navigate('/videos');
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ 
        backgroundColor: '#fff', 
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
        padding: '0 24px',
        display: 'flex',
        alignItems: 'center'
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
          <Space size="large" align="center">
            <Title level={3} style={{ margin: 0, color: '#1890ff' }}>
              视频上传系统
            </Title>
            <Button type="link" icon={<VideoCameraOutlined />} onClick={handleGoManagement}>
              视频管理
            </Button>
          </Space>
          <Space size="middle">
            {user ? (
              <>
                <Avatar size="small" icon={<UserOutlined />} />
                <Text>{user.account}</Text>
                <Button size="small" onClick={handleLogout}>
                  退出登录
                </Button>
              </>
            ) : (
              <Button type="primary" size="small" onClick={handleGoLogin}>
                登录
              </Button>
            )}
          </Space>
        </div>
      </Header>
      <Content>
        <VideoUpload />
      </Content>
    </Layout>
  );
};

export default Upload;
