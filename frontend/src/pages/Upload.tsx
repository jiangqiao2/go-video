import React from 'react';
import { Layout, Typography } from 'antd';
import VideoUpload from '@/components/upload/VideoUpload';

const { Content, Header } = Layout;
const { Title } = Typography;

const Upload: React.FC = () => {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ 
        backgroundColor: '#fff', 
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
        padding: '0 24px',
        display: 'flex',
        alignItems: 'center'
      }}>
        <Title level={3} style={{ margin: 0, color: '#1890ff' }}>
          视频上传系统
        </Title>
      </Header>
      <Content>
        <VideoUpload />
      </Content>
    </Layout>
  );
};

export default Upload;