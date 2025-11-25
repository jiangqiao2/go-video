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
    <Layout style={{ minHeight: '100vh', position: 'relative', overflow: 'hidden', background: '#0a0a0a' }}>
      {/* 复古渐变背景 - 暗紫红色调 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: 'linear-gradient(135deg, #1a0a1f 0%, #2d1b3d 25%, #4a1942 50%, #2d1b3d 75%, #1a0a1f 100%)',
        backgroundSize: '400% 400%',
        animation: 'vintage-gradient 20s ease infinite',
        zIndex: 0,
      }} />

      {/* VHS 扫描线效果 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: 'repeating-linear-gradient(0deg, transparent 0px, rgba(0, 0, 0, 0.3) 1px, transparent 2px, transparent 4px)',
        pointerEvents: 'none',
        zIndex: 5,
        opacity: 0.3,
      }} />

      {/* 移动的扫描线 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: '100%',
        background: 'linear-gradient(to bottom, transparent 0%, rgba(255, 255, 255, 0.05) 50%, transparent 100%)',
        animation: 'scan 8s linear infinite',
        pointerEvents: 'none',
        zIndex: 4,
      }} />

      {/* 像素网格背景 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundImage: `
          linear-gradient(rgba(255, 182, 193, 0.03) 1px, transparent 1px),
          linear-gradient(90deg, rgba(255, 182, 193, 0.03) 1px, transparent 1px)
        `,
        backgroundSize: '20px 20px',
        zIndex: 1,
        opacity: 0.4,
      }} />

      {/* 复古发光圆圈 */}
      <div style={{
        position: 'absolute',
        top: '15%',
        left: '10%',
        width: '400px',
        height: '400px',
        borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(255, 105, 180, 0.15) 0%, transparent 70%)',
        filter: 'blur(60px)',
        animation: 'retro-float 10s ease-in-out infinite',
        zIndex: 1,
      }} />

      <div style={{
        position: 'absolute',
        bottom: '10%',
        right: '15%',
        width: '350px',
        height: '350px',
        borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(138, 43, 226, 0.2) 0%, transparent 70%)',
        filter: 'blur(50px)',
        animation: 'retro-float 12s ease-in-out infinite 2s',
        zIndex: 1,
      }} />

      {/* 顶部导航栏 - 复古风格 */}
      <div className={mounted ? 'fade-in' : ''} style={{
        height: 70,
        background: 'rgba(26, 10, 31, 0.85)',
        backdropFilter: 'blur(10px)',
        WebkitBackdropFilter: 'blur(10px)',
        borderBottom: '2px solid rgba(255, 105, 180, 0.3)',
        boxShadow: '0 4px 20px rgba(255, 105, 180, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.1)',
        display: 'flex',
        alignItems: 'center',
        padding: '0 48px',
        justifyContent: 'space-between',
        position: 'relative',
        zIndex: 10,
      }}>
        <div
          className="hover-glitch"
          style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}
          onClick={() => window.location.href = '/'}
        >
          <div style={{
            width: 50,
            height: 50,
            background: 'linear-gradient(135deg, #ff69b4 0%, #ff1493 100%)',
            borderRadius: 8,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginRight: 14,
            color: '#fff',
            fontWeight: 'bold',
            fontSize: 22,
            fontFamily: '"Press Start 2P", cursive, monospace',
            boxShadow: '0 0 20px rgba(255, 105, 180, 0.6), inset 0 0 10px rgba(255, 255, 255, 0.2)',
            border: '2px solid rgba(255, 255, 255, 0.3)',
            transition: 'all 0.3s ease',
            textShadow: '2px 2px 4px rgba(0, 0, 0, 0.5)',
          }}>
            Go
          </div>
          <span style={{
            fontSize: 26,
            fontWeight: 700,
            fontFamily: '"Press Start 2P", cursive, monospace',
            color: '#ff69b4',
            textShadow: '0 0 10px rgba(255, 105, 180, 0.8), 2px 2px 4px rgba(0, 0, 0, 0.7)',
            letterSpacing: '2px',
          }}>
            GoVideo
          </span>
        </div>
        <a
          href="/"
          className="retro-button"
          style={{
            color: '#ff69b4',
            fontSize: 14,
            fontWeight: 600,
            padding: '10px 24px',
            borderRadius: 6,
            background: 'rgba(255, 105, 180, 0.1)',
            backdropFilter: 'blur(5px)',
            border: '2px solid rgba(255, 105, 180, 0.5)',
            textDecoration: 'none',
            transition: 'all 0.3s ease',
            textTransform: 'uppercase',
            letterSpacing: '1px',
            fontFamily: 'monospace',
            boxShadow: '0 0 10px rgba(255, 105, 180, 0.3)',
          }}
        >
          ← Return Home
        </a>
      </div>

      <Content className={mounted ? 'fade-in-up' : ''} style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        padding: '60px 20px',
        position: 'relative',
        zIndex: 10,
        animationDelay: '0.2s',
      }}>
        <LoginForm />
      </Content>

      {/* 底部复古文字 */}
      <div className={mounted ? 'fade-in typewriter' : ''} style={{
        position: 'absolute',
        bottom: 30,
        left: 0,
        right: 0,
        textAlign: 'center',
        color: 'rgba(255, 105, 180, 0.6)',
        fontSize: 14,
        zIndex: 10,
        animationDelay: '0.5s',
        fontFamily: '"Courier New", monospace',
        textShadow: '0 0 5px rgba(255, 105, 180, 0.5)',
      }}>
        <p style={{ margin: 0 }}>[ © 1984-2024 GoVideo Corporation - Retro Video Platform ]</p>
        <p style={{ margin: 0, fontSize: 12, marginTop: 4, opacity: 0.5 }}>* INSERT COIN TO CONTINUE *</p>
      </div>

      {/* VHS 噪点效果 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: 'url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAABZJREFUeNpi2r9//38gYGAEESAAEGAAasgJOgzOKCoAAAAASUVORK5CYII=")',
        opacity: 0.03,
        pointerEvents: 'none',
        zIndex: 6,
        animation: 'grain 0.3s steps(1) infinite',
      }} />
    </Layout>
  );
};

export default Login;
