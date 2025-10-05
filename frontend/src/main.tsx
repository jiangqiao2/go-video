import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import 'antd/dist/reset.css';
import router from './router';

// 配置Ant Design主题
const theme = {
  token: {
    colorPrimary: '#1890ff',
    borderRadius: 6,
  },
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN} theme={theme}>
      <RouterProvider router={router} />
    </ConfigProvider>
  </React.StrictMode>
);