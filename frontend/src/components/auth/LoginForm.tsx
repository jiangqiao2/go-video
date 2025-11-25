import React, { useState } from 'react';
import { Form, Input, Button, Card, message, Typography, Checkbox } from 'antd';
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
    <div className="scale-in" style={{
      background: 'rgba(255, 255, 255, 0.95)',
      backdropFilter: 'blur(20px)',
      WebkitBackdropFilter: 'blur(20px)',
      borderRadius: 20,
      padding: 48,
      width: '100%',
      maxWidth: 450,
      boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3), 0 0 40px rgba(102, 126, 234, 0.2)',
      border: '1px solid rgba(255, 255, 255, 0.5)',
      position: 'relative',
      overflow: 'hidden',
    }}>
      {/* 装饰性渐变条 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: 4,
        background: 'linear-gradient(90deg, #667eea 0%, #764ba2 50%, #f093fb 100%)',
        backgroundSize: '200% 100%',
        animation: 'gradient-shift 3s ease infinite',
      }} />

      {/* Logo和标题 */}
      <div style={{ textAlign: 'center', marginBottom: 40 }}>
        <div className="float" style={{
          width: 64,
          height: 64,
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          borderRadius: 16,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          margin: '0 auto 20px',
          color: '#fff',
          fontWeight: 'bold',
          fontSize: 28,
          boxShadow: '0 10px 30px rgba(102, 126, 234, 0.4)',
        }}>
          Go
        </div>
        <Title level={2} style={{
          color: '#18191c',
          marginBottom: 8,
          fontSize: 28,
          fontWeight: 700,
        }}>
          欢迎回来
        </Title>
        <Text type="secondary" style={{ fontSize: 15 }}>
          登录您的 GoVideo 账户
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
        >
          <Input
            prefix={<UserOutlined style={{ color: '#9499a0', fontSize: 16 }} />}
            placeholder="用户名"
            size="large"
            style={{
              borderRadius: 10,
              border: '2px solid #e8e8e8',
              padding: '12px 16px',
              fontSize: 15,
              transition: 'all 0.3s ease',
            }}
            onFocus={(e) => {
              e.target.style.borderColor = '#667eea';
              e.target.style.boxShadow = '0 0 0 3px rgba(102, 126, 234, 0.1)';
            }}
            onBlur={(e) => {
              e.target.style.borderColor = '#e8e8e8';
              e.target.style.boxShadow = 'none';
            }}
          />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[{ required: true, message: '请输入密码' }]}
        >
          <Input.Password
            prefix={<LockOutlined style={{ color: '#9499a0', fontSize: 16 }} />}
            placeholder="密码"
            size="large"
            style={{
              borderRadius: 10,
              border: '2px solid #e8e8e8',
              padding: '12px 16px',
              fontSize: 15,
              transition: 'all 0.3s ease',
            }}
            onFocus={(e) => {
              e.currentTarget.style.borderColor = '#667eea';
              e.currentTarget.style.boxShadow = '0 0 0 3px rgba(102, 126, 234, 0.1)';
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = '#e8e8e8';
              e.currentTarget.style.boxShadow = 'none';
            }}
          />
        </Form.Item>

        <Form.Item>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Form.Item name="remember" valuePropName="checked" noStyle>
              <Checkbox style={{ fontSize: 14 }}>记住我</Checkbox>
            </Form.Item>
            <Link to="/forgot-password" style={{
              color: '#667eea',
              fontSize: 14,
              fontWeight: 500,
              transition: 'all 0.3s ease',
            }}>
              忘记密码？
            </Link>
          </div>
        </Form.Item>

        <Form.Item style={{ marginBottom: 20 }}>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            size="large"
            block
            className="gradient-button"
            style={{
              height: 50,
              fontSize: 16,
              fontWeight: 600,
              borderRadius: 10,
              border: 'none',
              boxShadow: '0 4px 15px rgba(102, 126, 234, 0.4)',
            }}
          >
            {loading ? '登录中...' : '登录'}
          </Button>
        </Form.Item>

        <div style={{ textAlign: 'center' }}>
          <Text type="secondary" style={{ fontSize: 14 }}>
            还没有账户？
            <Link to="/register" style={{
              marginLeft: 6,
              color: '#667eea',
              fontWeight: 600,
              transition: 'all 0.3s ease',
            }}>
              立即注册
            </Link>
          </Text>
        </div>
      </Form>

      {/* 底部装饰光晕 */}
      <div style={{
        position: 'absolute',
        bottom: -50,
        left: '50%',
        transform: 'translateX(-50%)',
        width: 200,
        height: 100,
        background: 'radial-gradient(circle, rgba(102, 126, 234, 0.2) 0%, transparent 70%)',
        filter: 'blur(30px)',
        pointerEvents: 'none',
      }} />
    </div>
  );
};

export default LoginForm;
