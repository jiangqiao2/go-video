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
    <Card
      style={{
        width: 400,
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.08)',
        borderRadius: '8px',
        border: 'none',
      }}
    >
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        <div style={{
          width: 48,
          height: 48,
          backgroundColor: '#fb7299',
          borderRadius: 12,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          margin: '0 auto 16px',
          color: '#fff',
          fontWeight: 'bold',
          fontSize: 24
        }}>
          Go
        </div>
        <Title level={3} style={{ color: '#18191c', marginBottom: 8 }}>
          登录 GoVideo
        </Title>
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
            prefix={<UserOutlined style={{ color: '#bfbfbf' }} />}
            placeholder="用户名"
            size="large"
            style={{ borderRadius: 4 }}
          />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[{ required: true, message: '请输入密码' }]}
        >
          <Input.Password
            prefix={<LockOutlined style={{ color: '#bfbfbf' }} />}
            placeholder="密码"
            size="large"
            style={{ borderRadius: 4 }}
          />
        </Form.Item>

        <Form.Item>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Form.Item name="remember" valuePropName="checked" noStyle>
              <Checkbox>记住我</Checkbox>
            </Form.Item>
            <Link to="/forgot-password" style={{ color: '#fb7299' }}>忘记密码？</Link>
          </div>
        </Form.Item>

        <Form.Item style={{ marginBottom: 16 }}>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            size="large"
            block
            style={{
              height: 44,
              fontSize: 16,
              fontWeight: 500,
              backgroundColor: '#fb7299',
              borderColor: '#fb7299',
              borderRadius: 4,
            }}
          >
            登录
          </Button>
        </Form.Item>

        <div style={{ textAlign: 'center' }}>
          <Text type="secondary">
            还没有账户？
            <Link to="/register" style={{ marginLeft: 4, color: '#00a1d6' }}>
              立即注册
            </Link>
          </Text>
        </div>
      </Form>
    </Card>
  );
};

export default LoginForm;
