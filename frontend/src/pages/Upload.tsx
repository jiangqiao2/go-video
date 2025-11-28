import React from 'react';
import { Layout, Button, Space, Avatar, App } from 'antd';
import { UserOutlined, HomeOutlined, FileTextOutlined, UploadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import VideoUpload from '@/components/upload/VideoUpload';
import apiService from '@/services/api';
import { useAuthStore } from '@/store/auth';

const { Content, Header } = Layout;

const Upload: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const { user, logout, refreshUserInfo } = useAuthStore();
  const avatarInputRef = React.useRef<HTMLInputElement | null>(null);
  const [mounted, setMounted] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
    if (user) {
      refreshUserInfo().catch(console.error);
    }
  }, []);

  const handleLogout = () => {
    logout();
    message.success('已退出登录');
    navigate('/login', { replace: true });
  };

  const handleGoLogin = () => {
    navigate('/login');
  };

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
        <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
          <div
            className="hover-scale"
            style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}
            onClick={() => navigate('/')}
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
              创作中心
            </span>
          </div>

          {/* Navigation Links */}
          <Space size="small">
            <Button
              type="text"
              icon={<HomeOutlined />}
              className="hover-scale"
              onClick={() => navigate('/')}
              style={{
                borderRadius: 8,
                height: 38,
                fontWeight: 500,
                transition: 'all 0.3s ease',
              }}
            >
              主站
            </Button>
            <Button
              type="text"
              icon={<UploadOutlined />}
              className="hover-scale"
              onClick={() => navigate('/upload')}
              style={{
                borderRadius: 8,
                height: 38,
                fontWeight: 500,
                background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                color: '#667eea',
              }}
            >
              上传
            </Button>
            <Button
              type="text"
              icon={<FileTextOutlined />}
              className="hover-scale"
              onClick={() => navigate('/videos')}
              style={{
                borderRadius: 8,
                height: 38,
                fontWeight: 500,
                transition: 'all 0.3s ease',
              }}
            >
              管理
            </Button>
          </Space>
        </div>

        {/* User Actions */}
        <Space size="middle">
          {user ? (
            <>
              <div
                className="hover-scale"
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  padding: '6px 12px',
                  borderRadius: 20,
                  background: 'rgba(102, 126, 234, 0.1)',
                  transition: 'all 0.3s ease',
                }}
              >
                <Avatar
                  style={{
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    boxShadow: '0 2px 8px rgba(102, 126, 234, 0.3)',
                    cursor: 'pointer',
                  }}
                  src={user.avatar_url}
                  icon={<UserOutlined />}
                  onClick={() => avatarInputRef.current?.click()}
                />
                <input ref={avatarInputRef} type="file" accept="image/*" style={{ display: 'none' }} onChange={async (e) => {
                  const file = e.target.files?.[0];
                  if (!file) return;
                  try {
                    const res = await apiService.uploadImage({ file, category: 'avatar' });
                    await apiService.saveUserInfo({ avatar_url: res.url });
                    await refreshUserInfo();
                    message.success('头像更新成功');
                  } catch (err: any) {
                    message.error(err?.message || '头像更新失败');
                  } finally {
                    e.target.value = '';
                  }
                }} />
                <span style={{ marginLeft: 10, fontSize: 15, fontWeight: 600, color: '#18191c' }}>
                  {user.account}
                </span>
              </div>
              <Button
                type="link"
                onClick={handleLogout}
                style={{ color: '#9499a0', fontSize: 14 }}
              >
                退出
              </Button>
            </>
          ) : (
            <Button
              type="primary"
              ghost
              onClick={handleGoLogin}
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
          padding: 0,
          animationDelay: '0.2s',
        }}
      >
        <VideoUpload />
      </Content>
    </Layout>
  );
};

export default Upload;
