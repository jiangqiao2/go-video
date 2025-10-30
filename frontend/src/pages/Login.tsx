import React from 'react';
import { Layout } from 'antd';
import LoginForm from '@/components/auth/LoginForm';

const { Content } = Layout;

const Login: React.FC = () => {
  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#f0f2f5' }}>
      <Content style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center',
        padding: '50px 0'
      }}>
        <LoginForm />
      </Content>
    </Layout>
  );
};

export default Login;
