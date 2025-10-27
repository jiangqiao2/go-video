import React from 'react';
import { Layout, Card, Typography } from 'antd';
import LoginForm from '@/components/auth/LoginForm';

const { Content } = Layout;
const { Title, Text } = Typography;

const Login: React.FC = () => {
  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#f0f2f5' }}>
      <Content style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center',
        padding: '50px 0'
      }}>
        <Card 
          style={{ 
            width: 400, 
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
            borderRadius: 8
          }}
        >
          <div style={{ textAlign: 'center', marginBottom: 24 }}>
            <Title level={2} style={{ marginBottom: 8 }}>
              用户登录
            </Title>
            <Text type="secondary">
              登录您的账户访问视频上传服务
            </Text>
          </div>
          
          <LoginForm />
        </Card>
      </Content>
    </Layout>
  );
};

export default Login;
