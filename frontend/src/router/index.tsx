import React from 'react';
import { createBrowserRouter, Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/auth';
import Login from '@/pages/Login';
import Register from '@/pages/Register';
import Upload from '@/pages/Upload';
import VideoManagement from '@/pages/VideoManagement';
import Home from '@/pages/Home';
import Watch from '@/pages/Watch';
import Profile from '@/pages/Profile';
import Notifications from '@/pages/Notifications';
// 受保护的路由组件
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <>{children}</>;
};

// 公开路由组件：
// - 登录页/注册页：已登录用户重定向到首页
// - 首页：所有用户都可访问
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, logout } = useAuthStore();
  const location = useLocation();
  const token = localStorage.getItem('access_token');

  if (isAuthenticated && !token) {
    logout();
    return <>{children}</>;
  }

  const path = location.pathname;
  if (isAuthenticated && (path === '/login' || path === '/register')) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};

export const router = createBrowserRouter([
  {
    path: '/',
    element: (
      <PublicRoute>
        <Home />
      </PublicRoute>
    ),
  },
  {
    path: '/login',
    element: (
      <PublicRoute>
        <Login />
      </PublicRoute>
    ),
  },
  {
    path: '/register',
    element: (
      <PublicRoute>
        <Register />
      </PublicRoute>
    ),
  },
  {
    path: '/upload',
    element: (
      <ProtectedRoute>
        <Upload />
      </ProtectedRoute>
    ),
  },
  {
    path: '/videos',
    element: (
      <ProtectedRoute>
        <VideoManagement />
      </ProtectedRoute>
    ),
  },
  {
    path: '/watch/:video_uuid',
    element: (
      <PublicRoute>
        <Watch />
      </PublicRoute>
    ),
  },
  {
    path: '/user/:user_uuid',
    element: (
      <PublicRoute>
        <Profile />
      </PublicRoute>
    ),
  },
  {
    path: '/notifications/:tab?',
    element: (
      <ProtectedRoute>
        <Notifications />
      </ProtectedRoute>
    ),
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

export default router;
