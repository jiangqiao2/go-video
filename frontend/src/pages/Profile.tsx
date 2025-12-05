import React, { useEffect, useState } from 'react';
import { Layout, Typography, Button, Avatar, Tabs, Row, Col, Tag, App, Drawer, Form, Input, Space, Upload, Divider } from 'antd';
import { UserOutlined, MessageOutlined, EllipsisOutlined, PlusOutlined, CheckOutlined, EditOutlined, UploadOutlined, LockOutlined } from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { UserProfile, VideoDetail } from '@/types/api';
import apiService from '@/services/api';
import VideoCard from '@/components/common/VideoCard';
import { useAuthStore } from '@/store/auth';

const { Content } = Layout;
const { Title, Text, Paragraph } = Typography;

const Profile: React.FC = () => {
    const { user_uuid } = useParams();
    const { message } = App.useApp();
    const navigate = useNavigate();
    const currentUserUuid = useAuthStore((state) => state.user?.user_uuid);
    const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
    const refreshUserInfo = useAuthStore((state) => state.refreshUserInfo);

    const [loading, setLoading] = useState(true);
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [videos, setVideos] = useState<VideoDetail[]>([]);
    const [following, setFollowing] = useState(false);
    const [followLoading, setFollowLoading] = useState(false);
    const [editVisible, setEditVisible] = useState(false);
    const [savingProfile, setSavingProfile] = useState(false);
    const [savingPassword, setSavingPassword] = useState(false);
    const [avatarUploading, setAvatarUploading] = useState(false);
    const [profileForm] = Form.useForm();
    const [pwdForm] = Form.useForm();

    // 判断是否是查看自己的主页
    const isOwnProfile = Boolean(isAuthenticated && currentUserUuid && currentUserUuid === user_uuid);

    useEffect(() => {
        const run = async () => {
            if (!user_uuid) return;
            setLoading(true);
            try {
                const profileData = await apiService.getUserProfile(user_uuid);
                setProfile(profileData);
                setFollowing(!!profileData.is_followed);
            } catch (error) {
                try {
                    const basic = await apiService.getUserBasicInfo(user_uuid);
                    let relation: any = null;
                    try { relation = await apiService.getUserRelation(user_uuid); } catch {}
                    setProfile({
                        ...basic,
                        follower_count: relation?.follower_count ?? 0,
                        following_count: relation?.following_count ?? 0,
                        is_followed: !!relation?.is_followed,
                    });
                    setFollowing(!!relation?.is_followed);
                } catch {
                    message.error('获取用户信息失败');
                }
            }
            try {
                const videosData = await apiService.listVideosByUser(user_uuid, { page: 1, size: 20 });
                setVideos(videosData.videos);
            } catch {}
            setLoading(false);
        };
        run();
    }, [user_uuid]);

    const handleFollow = async () => {
        if (!isAuthenticated) {
            message.warning('请先登录');
            navigate('/login');
            return;
        }
        if (!profile || !user_uuid) return;

        setFollowLoading(true);
        try {
            if (following) {
                await apiService.unfollowUser(user_uuid);
            } else {
                await apiService.followUser(user_uuid);
            }
            const r = await apiService.getUserRelation(user_uuid);
            setFollowing(!!r.is_followed);
            message.success(r.is_followed ? '关注成功' : '已取消关注');
            setProfile(prev => prev ? ({
                ...prev,
                follower_count: r.follower_count,
                following_count: r.following_count,
            }) : prev);
        } catch (error) {
            console.error(error);
            message.error('操作失败');
        } finally {
            setFollowLoading(false);
        }
    };

    const handleEditProfile = () => {
        if (!profile) return;
            profileForm.setFieldsValue({
                nickname: profile.nickname || '',
            });
            setEditVisible(true);
        };

    const handleUploadAvatar = async (file: File) => {
        try {
            setAvatarUploading(true);
            const res = await apiService.uploadImage({ file, category: 'avatar' });
            profileForm.setFieldsValue({ avatar_url: res.url });
            setProfile(prev => prev ? { ...prev, avatar_url: res.url } : prev);
            message.success('头像上传成功');
        } catch (e: any) {
            const msg = e?.response?.data?.message || e?.message || '头像上传失败';
            message.error(msg);
        } finally {
            setAvatarUploading(false);
        }
    };

    const handleSaveProfile = async () => {
        try {
            const values = await profileForm.validateFields();
            setSavingProfile(true);
            const payload: any = {
                nickname: values.nickname,
                avatar_url: values.avatar_url || profile?.avatar_url,
            };
            const res = await apiService.saveUserInfo(payload);
            await refreshUserInfo();
            setProfile(prev => prev ? { ...prev, nickname: res.nickname, avatar_url: res.avatar_url } : prev);
            message.success('资料已更新');
            setEditVisible(false);
        } catch (error: any) {
            if (error?.errorFields) return;
            const msg = error?.response?.data?.message || error?.message || '更新失败';
            message.error(msg);
        } finally {
            setSavingProfile(false);
        }
    };

    const handleChangePassword = async () => {
        try {
            const values = await pwdForm.validateFields();
            setSavingPassword(true);
            await apiService.changePassword({
                old_password: values.old_password,
                new_password: values.new_password,
            });
            message.success('密码已更新');
            pwdForm.resetFields();
        } catch (error: any) {
            if (error?.errorFields) return;
            const msg = error?.response?.data?.message || error?.message || '修改密码失败';
            message.error(msg);
        } finally {
            setSavingPassword(false);
        }
    };

    if (loading) {
        return <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}><Content style={{ padding: 50, textAlign: 'center' }}>Loading...</Content></Layout>;
    }

    if (!profile) {
        return <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}><Content style={{ padding: 50, textAlign: 'center' }}>User not found</Content></Layout>;
    }

    return (
        <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}>
            <div style={{ height: 32 }} />

            <Content style={{ maxWidth: 1200, margin: '0 auto', width: '100%', padding: '0 16px' }}>
                {/* User Info Header */}
                <div style={{
                    position: 'relative',
                    padding: '0 24px 24px',
                    background: '#fff',
                    display: 'flex',
                    alignItems: 'flex-end',
                    gap: 24
                }}>
                    {/* Avatar */}
                    <div style={{ position: 'relative' }}>
                        <Avatar
                            size={84}
                            src={profile.avatar_url}
                            icon={<UserOutlined />}
                            style={{
                                border: '4px solid #fff',
                                boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
                            }}
                        />
                    </div>

                    {/* Info */}
                    <div style={{ flex: 1, paddingBottom: 12, color: '#1f2329' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                            <Title level={3} style={{ margin: 0 }}>{profile.nickname || '创作者'}</Title>
                            <Tag color="#f50">Lv6</Tag>
                            <Tag color="#ff4d4f">年度大会员</Tag>
                        </div>
                        <Paragraph style={{ margin: '8px 0 0', color: '#666', maxWidth: 600 }} ellipsis={{ rows: 2 }}>
                            {profile.description || '这个人很懒，什么都没留下'}
                        </Paragraph>
                    </div>

                    {/* Actions - 区分自己和他人 */}
                    <div style={{ paddingBottom: 12, display: 'flex', gap: 12 }}>
                        {isOwnProfile ? (
                            // 自己的主页 - 显示编辑资料
                            <>
                                <Button
                                    type="primary"
                                    icon={<EditOutlined />}
                                    size="large"
                                    onClick={handleEditProfile}
                                    style={{ width: 120, borderRadius: 6 }}
                                >
                                    编辑资料
                                </Button>
                                <Button
                                    icon={<EllipsisOutlined />}
                                    size="large"
                                    ghost
                                    style={{ color: '#fff', borderColor: 'rgba(255,255,255,0.6)' }}
                                />
                            </>
                        ) : (
                            // 别人的主页 - 显示关注按钮
                            <>
                                <Button
                                    type={following ? 'default' : 'primary'}
                                    icon={following ? <CheckOutlined /> : <PlusOutlined />}
                                    size="large"
                                    onClick={handleFollow}
                                    loading={followLoading}
                                    style={{ width: 120, borderRadius: 6 }}
                                >
                                    {following ? '已关注' : '关注'}
                                </Button>
                                <Button
                                    icon={<MessageOutlined />}
                                    size="large"
                                    ghost
                                    style={{ color: '#fff', borderColor: 'rgba(255,255,255,0.6)' }}
                                >
                                    发消息
                                </Button>
                                <Button
                                    icon={<EllipsisOutlined />}
                                    size="large"
                                    ghost
                                    style={{ color: '#fff', borderColor: 'rgba(255,255,255,0.6)' }}
                                />
                            </>
                        )}
                    </div>
                </div>

                {/* Stats Bar */}
                <div style={{ background: '#fff', borderRadius: '0 0 12px 12px', padding: '16px 24px', display: 'flex', gap: 40, boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
                    <div style={{ textAlign: 'center' }}>
                        <Text type="secondary">关注数</Text>
                        <div style={{ fontSize: 18, fontWeight: 'bold' }}>{profile.following_count}</div>
                    </div>
                    <div style={{ textAlign: 'center' }}>
                        <Text type="secondary">粉丝数</Text>
                        <div style={{ fontSize: 18, fontWeight: 'bold' }}>{profile.follower_count}</div>
                    </div>
                    <div style={{ textAlign: 'center' }}>
                        <Text type="secondary">获赞数</Text>
                        <div style={{ fontSize: 18, fontWeight: 'bold' }}>999+</div>
                    </div>
                </div>

                {/* Content Tabs */}
                <div style={{ marginTop: 24, background: '#fff', borderRadius: 12, padding: 24, minHeight: 400 }}>
                    <Tabs defaultActiveKey="1" items={[
                        {
                            key: '1',
                            label: '主页',
                            children: (
                                <Row gutter={[24, 24]}>
                                    {videos.map(video => (
                                        <Col key={video.video_uuid} xs={24} sm={12} md={8} lg={6}>
                                            <VideoCard
                                                video={video}
                                                onClick={(v) => window.location.href = `/watch/${v.video_uuid}`}
                                                uploaderName={profile.nickname || '创作者'}
                                                uploaderAvatar={profile.avatar_url}
                                            />
                                        </Col>
                                    ))}
                                    {videos.length === 0 && <EmptyState />}
                                </Row>
                            )
                        },
                        {
                            key: '2',
                            label: '动态',
                            children: <EmptyState />
                        },
                        {
                            key: '3',
                            label: '投稿',
                            children: (
                                <Row gutter={[24, 24]}>
                                    {videos.map(video => (
                                        <Col key={video.video_uuid} xs={24} sm={12} md={8} lg={6}>
                                            <VideoCard
                                                video={video}
                                                onClick={(v) => window.location.href = `/watch/${v.video_uuid}`}
                                                uploaderName={profile.nickname || '创作者'}
                                                uploaderAvatar={profile.avatar_url}
                                            />
                                        </Col>
                                    ))}
                                    {videos.length === 0 && <EmptyState />}
                                </Row>
                            )
                        },
                        {
                            key: '4',
                            label: '合集和列表',
                            children: <EmptyState />
                        }
                    ]} />
                </div>
            </Content>

            <Drawer
                title="编辑资料"
                width={480}
                open={editVisible}
                onClose={() => setEditVisible(false)}
                destroyOnClose
            >
                <Form layout="vertical" form={profileForm} initialValues={{ nickname: profile?.nickname }}>
                    <Form.Item label="头像">
                        <Space align="center" size="middle">
                            <Avatar size={64} src={profile?.avatar_url} icon={<UserOutlined />} />
                            <Upload
                                accept="image/*"
                                showUploadList={false}
                                beforeUpload={(file) => {
                                    handleUploadAvatar(file);
                                    return false;
                                }}
                            >
                                <Button icon={<UploadOutlined />} loading={avatarUploading}>更换头像</Button>
                            </Upload>
                        </Space>
                    </Form.Item>
                    <Form.Item label="账号">
                        <Input value={profile?.account || '—'} disabled />
                    </Form.Item>
                    <Form.Item
                        label="用户名"
                        name="nickname"
                        rules={[{ required: true, message: '请输入用户名' }]}
                    >
                        <Input placeholder="用户名" />
                    </Form.Item>
                    <Form.Item name="avatar_url" hidden>
                        <Input />
                    </Form.Item>
                    <Button type="primary" block onClick={handleSaveProfile} loading={savingProfile}>
                        保存资料
                    </Button>
                </Form>

                <Divider />

                <Form layout="vertical" form={pwdForm}>
                    <Form.Item
                        label="当前密码"
                        name="old_password"
                        rules={[{ required: true, message: '请输入当前密码' }]}
                    >
                        <Input.Password prefix={<LockOutlined />} placeholder="当前密码" />
                    </Form.Item>
                    <Form.Item
                        label="新密码"
                        name="new_password"
                        rules={[{ required: true, message: '请输入新密码' }, { min: 8, message: '至少8个字符' }]}
                    >
                        <Input.Password prefix={<LockOutlined />} placeholder="新密码" />
                    </Form.Item>
                    <Form.Item
                        label="确认新密码"
                        name="confirm_password"
                        dependencies={['new_password']}
                        rules={[
                            { required: true, message: '请确认新密码' },
                            ({ getFieldValue }) => ({
                                validator(_, value) {
                                    if (!value || getFieldValue('new_password') === value) {
                                        return Promise.resolve();
                                    }
                                    return Promise.reject(new Error('两次输入的密码不一致'));
                                },
                            }),
                        ]}
                    >
                        <Input.Password prefix={<LockOutlined />} placeholder="再次输入新密码" />
                    </Form.Item>
                    <Button block onClick={handleChangePassword} loading={savingPassword}>
                        修改密码
                    </Button>
                </Form>
            </Drawer>
        </Layout>
    );
};

const EmptyState = () => (
    <div style={{ padding: 40, textAlign: 'center', color: '#999' }}>
        暂无内容
    </div>
);

export default Profile;
