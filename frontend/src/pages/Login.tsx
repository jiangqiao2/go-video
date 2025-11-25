import React, { useEffect, useState } from 'react';
import { Layout } from 'antd';
import LoginForm from '@/components/auth/LoginForm';

const { Content } = Layout;

const Login: React.FC = () => {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  return (
    <Layout style={{ minHeight: '100vh', position: 'relative', overflow: 'hidden' }}>
      {/* 动态渐变背景 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%)',
        backgroundSize: '400% 400%',
        animation: 'gradient-shift 15s ease infinite',
        zIndex: 0,
      }} />

      {/* 网格背景层 */}
      <div className="grid-background" style={{ zIndex: 1, opacity: 0.3 }} />

      {/* 浮动装饰圆圈 */}
      <div style={{
        position: 'absolute',
        top: '10%',
        left: '5%',
        width: '300px',
        height: '300px',
        borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(255,255,255,0.2) 0%, transparent 70%)',
        filter: 'blur(40px)',
        animation: 'float 6s ease-in-out infinite',
        zIndex: 1,
      }} />

      <div style={{
        position: 'absolute',
        bottom: '15%',
        right: '10%',
        width: '400px',
        height: '400px',
        borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(255,255,255,0.15) 0%, transparent 70%)',
        filter: 'blur(50px)',
        animation: 'float 8s ease-in-out infinite 1s',
        zIndex: 1,
      }} />

      {/* 顶部导航栏 - 玻璃态效果 */}
      <div className={mounted ? 'fade-in' : ''} style={{
        height: 70,
        background: 'rgba(255, 255, 255, 0.1)',
        backdropFilter: 'blur(20px)',
        WebkitBackdropFilter: 'blur(20px)',
        borderBottom: '1px solid rgba(255, 255, 255, 0.2)',
        boxShadow: '0 4px 30px rgba(0, 0, 0, 0.1)',
        display: 'flex',
        alignItems: 'center',
        padding: '0 48px',
        justifyContent: 'space-between',
        position: 'relative',
        zIndex: 10,
      }}>
        <div
          className="hover-scale"
          style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}
          onClick={() => window.location.href = '/'}
        >
          <div style={{
            width: 42,
            height: 42,
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            borderRadius: 10,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginRight: 12,
            color: '#fff',
            fontWeight: 'bold',
            fontSize: 18,
            boxShadow: '0 4px 15px rgba(102, 126, 234, 0.4)',
            transition: 'all 0.3s ease',
          }}>
            Go
          </div>
          <span className="gradient-text" style={{
            fontSize: 22,
            fontWeight: 700,
            color: '#fff',
            textShadow: '0 2px 10px rgba(0,0,0,0.2)',
          }}>
            GoVideo
          </span>
        </div>
        <a
          href="/"
          className="hover-glow"
          style={{
            color: '#fff',
            fontSize: 15,
            fontWeight: 500,
            padding: '8px 20px',
            borderRadius: 20,
            background: 'rgba(255, 255, 255, 0.15)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(255, 255, 255, 0.3)',
            textDecoration: 'none',
            transition: 'all 0.3s ease',
          }}
        >
          返回首页
        </a>
      </div>

      <Content className={mounted ? 'fade-in-up' : ''} style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        padding: '80px 20px',
        position: 'relative',
        zIndex: 10,
        animationDelay: '0.2s',
      }}>
        <LoginForm />
      </Content>

      {/* 底部装饰文字 */}
      <div className={mounted ? 'fade-in' : ''} style={{
        position: 'absolute',
        bottom: 30,
        left: 0,
        right: 0,
        textAlign: 'center',
        color: 'rgba(255, 255, 255, 0.6)',
        fontSize: 13,
        zIndex: 10,
        animationDelay: '0.4s',
      }}>
        <p style={{ margin: 0 }}>© 2024 GoVideo - 下一代视频分享平台</p>
      </div>
    </Layout>
  );
};

export default Login;
