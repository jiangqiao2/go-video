import React from 'react';
import { Typography, Row, Col, Card, Space, Tag } from 'antd';
import { FileOutlined, VideoCameraOutlined, PictureOutlined, PlayCircleOutlined } from '@ant-design/icons';
import VideoUpload from '@/components/upload/VideoUpload';
import CreatorLayout from '@/components/layout/CreatorLayout';

const { Text, Title } = Typography;

const Upload: React.FC = () => {
  return (
    <CreatorLayout activeKey="upload">
      <div style={{
        background: 'linear-gradient(135deg, #f0f7ff 0%, #fff 60%)',
        padding: '24px 32px',
        minHeight: 520,
        borderRadius: 12,
        boxShadow: 'var(--shadow-sm)',
      }}>
        <div style={{ marginBottom: 16 }}>
          <Title level={4} style={{ margin: 0 }}>投稿中心</Title>
          <Text type="secondary">按照指引上传素材，自动转码生成多端可播放的稿件</Text>
        </div>
        <div className="fade-in">
          <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
            {[{
              title: '视频大小',
              desc: '单个文件≤2G，时长≤10小时；超大文件请分片上传',
              icon: <FileOutlined style={{ color: '#1890ff', fontSize: 20 }} />,
            }, {
              title: '视频格式',
              desc: '推荐 MP4 / MOV / MKV，转码更快、画质更稳定',
              icon: <VideoCameraOutlined style={{ color: '#1890ff', fontSize: 20 }} />,
            }, {
              title: '分辨率与帧率',
              desc: '推荐 1080P / 4K，高帧率更清晰流畅',
              icon: <PictureOutlined style={{ color: '#1890ff', fontSize: 20 }} />,
            }].map((item) => (
              <Col span={8} key={item.title}>
                <Card bordered={false} style={{ minHeight: 120 }}>
                  <Space align="start">
                    <div style={{
                      width: 40, height: 40, background: '#e6f7ff', borderRadius: 8,
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                    }}>
                      {item.icon}
                    </div>
                    <div>
                      <div style={{ fontWeight: 600, marginBottom: 6 }}>{item.title}</div>
                      <Text type="secondary" style={{ fontSize: 12, lineHeight: 1.6 }}>{item.desc}</Text>
                    </div>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>

          <Card bordered={false} style={{ boxShadow: 'var(--shadow-sm)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
              <Space>
                <PlayCircleOutlined style={{ color: '#00a1d6' }} />
                <Text strong>上传视频</Text>
                <Tag color="blue">自动转码</Tag>
                <Tag color="green">HLS</Tag>
              </Space>
              <Text type="secondary" style={{ fontSize: 12 }}>上传后可在「稿件管理」查看进度</Text>
            </div>
            <VideoUpload />
          </Card>
        </div>
      </div>
    </CreatorLayout>
  );
};

export default Upload;
