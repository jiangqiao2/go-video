import React from 'react';
import { Layout } from 'antd';
import RegisterForm from '@/components/auth/RegisterForm';

const { Content } = Layout;

const Register: React.FC = () => {
  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#f0f2f5' }}>
      {/* 简易头部 */}
      <div style={{
        height: 60,
        backgroundColor: '#fff',
        boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
        display: 'flex',
        alignItems: 'center',
        padding: '0 40px',
        justifyContent: 'space-between'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }} onClick={() => window.location.href = '/'}>
          <div style={{
            width: 32,
            height: 32,
            backgroundColor: '#fb7299',
            borderRadius: 8,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginRight: 8,
            color: '#fff',
            fontWeight: 'bold',
            fontSize: 16
          }}>
            Go
          </div>
          <span style={{ fontSize: 18, fontWeight: 600, color: '#18191c' }}>GoVideo</span>
        </div>
        <a href="/" style={{ color: '#61666d', fontSize: 14 }}>返回首页</a>
      </div>

      <Content style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        padding: '50px 0',
        backgroundImage: 'url(https://s1.hdslb.com/bfs/static/jinkela/long/images/login_bg.png)',
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }}>
        <RegisterForm />
      </Content>
    </Layout>
  );
};

export default Register;
