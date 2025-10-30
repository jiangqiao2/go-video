import React, { useState } from 'react';
import { Form, Input, Button, Card, message, Typography } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/auth';

const { Title, Text } = Typography;

interface RegisterFormData {
  account: string;
  password: string;
  confirmPassword: string;
}

const RegisterForm: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { register } = useAuthStore();
  const redirectPath =
    (location.state as { from?: { pathname?: string } })?.from?.pathname ?? '/upload';

  const onFinish = async (values: RegisterFormData) => {
    if (values.password !== values.confirmPassword) {
      message.error('两次输入的密码不一致');
      return;
    }

    setLoading(true);
    try {
      await register({
        account: values.account,
        password: values.password,
      });
      
      message.success('注册成功');
      navigate(redirectPath, { replace: true });
    } catch (error: any) {
      console.error('Registration error:', error);
      const errorMessage = error.response?.data?.message || '注册失败，请重试';
      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card
      style={{
        width: 400,
        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)',
        borderRadius: '12px',
      }}
    >
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ color: '#1890ff', marginBottom: 8 }}>
          视频平台
        </Title>
        <Text type="secondary">创建您的账户</Text>
      </div>

      <Form form={form} name="register" onFinish={onFinish} autoComplete="off" layout="vertical">
        <Form.Item
          name="account"
          label="用户名"
          rules={[
            { required: true, message: '请输入用户名' },
            { min: 3, message: '用户名至少3个字符' },
            { max: 50, message: '用户名最多50个字符' },
            { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' },
          ]}
        >
          <Input prefix={<UserOutlined />} placeholder="请输入用户名" size="large" />
        </Form.Item>

        <Form.Item
          name="password"
          label="密码"
          rules={[
            { required: true, message: '请输入密码' },
            { min: 8, message: '密码至少8个字符' },
            { max: 100, message: '密码最多100个字符' },
            {
              pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]/,
              message: '密码必须包含大小写字母和数字',
            },
          ]}
        >
          <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" size="large" />
        </Form.Item>

        <Form.Item
          name="confirmPassword"
          label="确认密码"
          dependencies={['password']}
          rules={[
            { required: true, message: '请确认密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('两次输入的密码不一致'));
              },
            }),
          ]}
        >
          <Input.Password prefix={<LockOutlined />} placeholder="请再次输入密码" size="large" />
        </Form.Item>

        <Form.Item style={{ marginBottom: 16 }}>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            size="large"
            block
            style={{
              height: 48,
              fontSize: 16,
              fontWeight: 500,
            }}
          >
            注册
          </Button>
        </Form.Item>

        <div style={{ textAlign: 'center' }}>
          <Text type="secondary">
            已有账户？
            <Link to="/login" style={{ marginLeft: 4 }}>
              立即登录
            </Link>
          </Text>
        </div>
      </Form>
    </Card>
  );
};

export default RegisterForm;
