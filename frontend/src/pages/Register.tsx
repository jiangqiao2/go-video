import React from 'react';
import { Layout } from 'antd';
import RegisterForm from '@/components/auth/RegisterForm';

const { Content } = Layout;

const Register: React.FC = () => {
  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#f0f2f5' }}>
      <Content style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center',
        padding: '50px 0'
      }}>
        <RegisterForm />
      </Content>
    </Layout>
  );
};

export default Register;
