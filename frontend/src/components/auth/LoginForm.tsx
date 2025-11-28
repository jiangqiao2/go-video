import React, { useState } from 'react';
import { Form, Input, Button, Typography, Checkbox, App } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/auth';

const { Title, Text } = Typography;

interface LoginFormData {
  account: string;
  password: string;
  remember: boolean;
}

const LoginForm: React.FC = () => {
  const [form] = Form.useForm();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuthStore();
  const redirectPath =
    (location.state as { from?: { pathname?: string } })?.from?.pathname ?? '/';

  const onFinish = async (values: LoginFormData) => {
    setLoading(true);
    try {
      await login({
        account: values.account,
        password: values.password,
      });

      message.success('登录成功');
      navigate(redirectPath, { replace: true });
    } catch (error: any) {
      console.error('Login error:', error);
      const errorMessage = error.response?.data?.message || '登录失败，请检查用户名和密码';
      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="scale-in retro-card" style={{
      background: 'rgba(26, 10, 31, 0.95)',
      backdropFilter: 'blur(15px)',
      WebkitBackdropFilter: 'blur(15px)',
      borderRadius: 12,
      padding: 48,
      width: '100%',
      maxWidth: 480,
      boxShadow: '0 0 40px rgba(255, 105, 180, 0.4), 0 20px 60px rgba(0, 0, 0, 0.6), inset 0 0 60px rgba(255, 105, 180, 0.1)',
      border: '3px solid rgba(255, 105, 180, 0.4)',
      position: 'relative',
      overflow: 'hidden',
    }}>
      {/* 装饰性霓虹边框 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: 3,
        background: 'linear-gradient(90deg, #ff69b4 0%, #ff1493 25%, #ba55d3 50%, #ff1493 75%, #ff69b4 100%)',
        backgroundSize: '200% 100%',
        animation: 'neon-flow 3s linear infinite',
        boxShadow: '0 0 15px rgba(255, 105, 180, 0.8)',
      }} />

      <div style={{
        position: 'absolute',
        bottom: 0,
        left: 0,
        right: 0,
        height: 3,
        background: 'linear-gradient(90deg, #ff69b4 0%, #ff1493 25%, #ba55d3 50%, #ff1493 75%, #ff69b4 100%)',
        backgroundSize: '200% 100%',
        animation: 'neon-flow 3s linear infinite reverse',
        boxShadow: '0 0 15px rgba(255, 105, 180, 0.8)',
      }} />

      {/* Logo和标题 */}
      <div style={{ textAlign: 'center', marginBottom: 40 }}>
        <div className="retro-pulse" style={{
          width: 80,
          height: 80,
          background: 'linear-gradient(135deg, #ff69b4 0%, #ff1493 100%)',
          borderRadius: 12,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          margin: '0 auto 24px',
          color: '#fff',
          fontWeight: 'bold',
          fontSize: 32,
          fontFamily: '"Press Start 2P", cursive, monospace',
          boxShadow: '0 0 30px rgba(255, 105, 180, 0.8), 0 0 60px rgba(255, 105, 180, 0.4), inset 0 0 20px rgba(255, 255, 255, 0.2)',
          border: '3px solid rgba(255, 255, 255, 0.3)',
          textShadow: '0 0 10px rgba(255, 255, 255, 0.8), 2px 2px 4px rgba(0, 0, 0, 0.8)',
        }}>
          Go
        </div>
        <Title level={2} style={{
          color: '#ff69b4',
          marginBottom: 12,
          fontSize: 32,
          fontWeight: 700,
          fontFamily: '"Press Start 2P", cursive, monospace',
          textShadow: '0 0 15px rgba(255, 105, 180, 0.8), 3px 3px 6px rgba(0, 0, 0, 0.8)',
          letterSpacing: '2px',
        }}>
          WELCOME
        </Title>
        <Text style={{
          fontSize: 14,
          color: 'rgba(255, 182, 193, 0.8)',
          fontFamily: 'monospace',
          textTransform: 'uppercase',
          letterSpacing: '1px',
        }}>
          [ Login to Go Video System ]
        </Text>
      </div>

      <Form
        form={form}
        name="login"
        onFinish={onFinish}
        autoComplete="off"
        layout="vertical"
        initialValues={{ remember: true }}
      >
        <Form.Item
          name="account"
          rules={[{ required: true, message: '请输入用户名' }]}
          style={{ marginBottom: 24 }}
        >
          <Input
            prefix={<UserOutlined style={{ color: '#ff69b4', fontSize: 18 }} />}
            placeholder="USERNAME"
            size="large"
            style={{
              borderRadius: 6,
              border: '2px solid rgba(255, 105, 180, 0.5)',
              background: 'rgba(0, 0, 0, 0.4)',
              padding: '14px 16px',
              fontSize: 16,
              color: '#ffb6c1',
              fontFamily: 'monospace',
              transition: 'all 0.3s ease',
              boxShadow: 'inset 0 2px 8px rgba(0, 0, 0, 0.4)',
            }}
            onFocus={(e) => {
              e.target.style.borderColor = '#ff69b4';
              e.target.style.boxShadow = '0 0 15px rgba(255, 105, 180, 0.6), inset 0 2px 8px rgba(0, 0, 0, 0.4)';
              e.target.style.background = 'rgba(0, 0, 0, 0.6)';
            }}
            onBlur={(e) => {
              e.target.style.borderColor = 'rgba(255, 105, 180, 0.5)';
              e.target.style.boxShadow = 'inset 0 2px 8px rgba(0, 0, 0, 0.4)';
              e.target.style.background = 'rgba(0, 0, 0, 0.4)';
            }}
          />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[{ required: true, message: '请输入密码' }]}
          style={{ marginBottom: 24 }}
        >
          <Input.Password
            prefix={<LockOutlined style={{ color: '#ff69b4', fontSize: 18 }} />}
            placeholder="PASSWORD"
            size="large"
            style={{
              borderRadius: 6,
              border: '2px solid rgba(255, 105, 180, 0.5)',
              background: 'rgba(0, 0, 0, 0.4)',
              padding: '14px 16px',
              fontSize: 16,
              color: '#ffb6c1',
              fontFamily: 'monospace',
              transition: 'all 0.3s ease',
              boxShadow: 'inset 0 2px 8px rgba(0, 0, 0, 0.4)',
            }}
            onFocus={(e) => {
              e.currentTarget.style.borderColor = '#ff69b4';
              e.currentTarget.style.boxShadow = '0 0 15px rgba(255, 105, 180, 0.6), inset 0 2px 8px rgba(0, 0, 0, 0.4)';
              e.currentTarget.style.background = 'rgba(0, 0, 0, 0.6)';
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = 'rgba(255, 105, 180, 0.5)';
              e.currentTarget.style.boxShadow = 'inset 0 2px 8px rgba(0, 0, 0, 0.4)';
              e.currentTarget.style.background = 'rgba(0, 0, 0, 0.4)';
            }}
          />
        </Form.Item>

        <Form.Item>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Form.Item name="remember" valuePropName="checked" noStyle>
              <Checkbox style={{
                fontSize: 14,
                color: 'rgba(255, 182, 193, 0.8)',
                fontFamily: 'monospace',
              }}>
                Remember Me
              </Checkbox>
            </Form.Item>
            <Link to="/forgot-password" style={{
              color: '#ff69b4',
              fontSize: 14,
              fontWeight: 600,
              fontFamily: 'monospace',
              textShadow: '0 0 5px rgba(255, 105, 180, 0.5)',
              transition: 'all 0.3s ease',
            }}>
              Forgot?
            </Link>
          </div>
        </Form.Item>

        <Form.Item style={{ marginBottom: 24 }}>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            size="large"
            block
            className="retro-button-primary"
            style={{
              height: 54,
              fontSize: 18,
              fontWeight: 700,
              fontFamily: '"Press Start 2P", cursive, monospace',
              borderRadius: 8,
              border: '3px solid #ff69b4',
              background: 'linear-gradient(135deg, #ff1493 0%, #ff69b4 100%)',
              boxShadow: '0 0 20px rgba(255, 105, 180, 0.6), 0 4px 15px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.3)',
              textShadow: '2px 2px 4px rgba(0, 0, 0, 0.8)',
              textTransform: 'uppercase',
              letterSpacing: '2px',
              transition: 'all 0.3s ease',
            }}
          >
            {loading ? 'Loading...' : '>> Login <<'}
          </Button>
        </Form.Item>

        <div style={{ textAlign: 'center', paddingTop: 16, borderTop: '1px solid rgba(255, 105, 180, 0.3)' }}>
          <Text style={{
            fontSize: 14,
            color: 'rgba(255, 182, 193, 0.7)',
            fontFamily: 'monospace',
          }}>
            No Account Yet?{' '}
            <Link to="/register" style={{
              marginLeft: 8,
              color: '#ff69b4',
              fontWeight: 700,
              textShadow: '0 0 5px rgba(255, 105, 180, 0.5)',
              transition: 'all 0.3s ease',
              textDecoration: 'underline',
            }}>
              [ Register Now ]
            </Link>
          </Text>
        </div>
      </Form>

      {/* 底部装饰光晕 */}
      <div style={{
        position: 'absolute',
        bottom: -60,
        left: '50%',
        transform: 'translateX(-50%)',
        width: 250,
        height: 120,
        background: 'radial-gradient(circle, rgba(255, 105, 180, 0.3) 0%, transparent 70%)',
        filter: 'blur(40px)',
        pointerEvents: 'none',
      }} />
    </div>
  );
};

export default LoginForm;
