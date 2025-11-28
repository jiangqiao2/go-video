import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import 'antd/dist/reset.css';
import './index.css';
import router from './router';

// 配置Ant Design主题 - 更科技感的配色
const theme = {
  token: {
    colorPrimary: '#00a1d6',
    colorSuccess: '#00d4aa',
    colorWarning: '#ff8f16',
    colorError: '#ff6b9d',
    borderRadius: 8,
    colorBgContainer: '#ffffff',
    fontSize: 14,
  },
  components: {
    Button: {
      borderRadius: 6,
      controlHeight: 40,
    },
    Card: {
      borderRadius: 12,
    },
    Input: {
      borderRadius: 6,
      controlHeight: 40,
    },
  },
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN} theme={theme}>
      <AntdApp>
        <RouterProvider router={router} />
      </AntdApp>
    </ConfigProvider>
  </React.StrictMode>
);
