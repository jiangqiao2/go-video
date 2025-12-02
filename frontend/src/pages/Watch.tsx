import React, { useCallback, useEffect, useState } from 'react';
import { Layout, Button, Typography, Spin, Empty, Space, Avatar, Tag, Dropdown, App, List, Tabs, Input, Pagination } from 'antd';
import { ArrowLeftOutlined, UserOutlined, PlusOutlined, MenuOutlined, LikeOutlined, MessageOutlined, FireOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import apiService from '@/services/api';
import { CommentItem, UserBasicInfo, VideoDetail } from '@/types/api';
import VideoPlayer from '@/components/common/VideoPlayer';
import { useAuthStore } from '@/store/auth';
import { formatPublishedTime } from '@/utils/date';

const { Content } = Layout;
const { Title, Text, Paragraph } = Typography;

type CommentWithReplies = CommentItem & {
  replies?: CommentItem[];
  repliesLoaded?: boolean;
  repliesLoading?: boolean;
  replyPage?: number;
  replyTotal?: number;
};

const Watch: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const { video_uuid } = useParams();
  const [loading, setLoading] = useState(true);
  const [video, setVideo] = useState<VideoDetail | null>(null);
  const [following, setFollowing] = useState(false);
  const [followLoading, setFollowLoading] = useState(false);
  const [liked, setLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(0);
  const [comments, setComments] = useState<CommentWithReplies[]>([]);
  const [commentTotal, setCommentTotal] = useState(0);
  const [commentLoading, setCommentLoading] = useState(false);
  const [commentSort, setCommentSort] = useState<'hot' | 'time'>('hot');
  const [commentInput, setCommentInput] = useState('');
  const [commentSubmitting, setCommentSubmitting] = useState(false);
  const [replyingTo, setReplyingTo] = useState<{ rootId: string; target: CommentWithReplies } | null>(null);
  const [replyContent, setReplyContent] = useState('');
  const [userInfoMap, setUserInfoMap] = useState<Record<string, UserBasicInfo>>({});
  const currentUserUuid = useAuthStore((s) => s.user?.user_uuid) || localStorage.getItem('user_uuid') || '';
  const isOwner = !!video?.user_uuid && video.user_uuid === currentUserUuid;
  const REPLY_PAGE_SIZE = 10;

  useEffect(() => {
    const run = async () => {
      setLoading(true);
      try {
        if (!video_uuid) return;
        const v = await apiService.getVideo(video_uuid);
        let enriched = v;
        if (v.user_uuid) {
          try {
            const info = await apiService.getUserBasicInfo(v.user_uuid);
            enriched = { ...v, uploader_account: info.nickname || '创作者', uploader_avatar_url: info.avatar_url } as VideoDetail;
            const r = await apiService.getUserRelation(v.user_uuid);
            setFollowing(!!r.is_followed);
          } catch {}
        }
        setVideo(enriched);
        setLiked(!!enriched.liked);
        setLikeCount(enriched.like_count ?? 0);
        setCommentTotal(enriched.comment_count ?? 0);
        await fetchComments(1, commentSort, enriched.video_uuid);
      } finally {
        setLoading(false);
      }
    };
    run();
  }, [video_uuid]);

  const handleFollow = async () => {
    if (!video?.user_uuid) return;
    if (isOwner) {
      message.warning('不能关注自己');
      return;
    }
    const isAuth = !!localStorage.getItem('access_token');
    if (!isAuth) {
      message.warning('请先登录');
      navigate('/login');
      return;
    }
    setFollowLoading(true);
    try {
      if (following) {
        await apiService.unfollowUser(video.user_uuid);
      } else {
        await apiService.followUser(video.user_uuid);
      }
      const r = await apiService.getUserRelation(video.user_uuid);
      setFollowing(!!r.is_followed);
      message.success(r.is_followed ? '关注成功' : '已取消关注');
    } catch (error: any) {
      const msg = error?.response?.data?.message || '操作失败';
      if (typeof msg === 'string' && msg.includes('不能关注自己')) {
        message.warning('不能关注自己');
      } else {
        message.error(msg);
      }
    } finally {
      setFollowLoading(false);
    }
  };

  const handleUnfollow = async () => {
    if (!video?.user_uuid) return;
    const isAuth = !!localStorage.getItem('access_token');
    if (!isAuth) {
      message.warning('请先登录');
      navigate('/login');
      return;
    }
    setFollowLoading(true);
    try {
      await apiService.unfollowUser(video.user_uuid);
      const r = await apiService.getUserRelation(video.user_uuid);
      setFollowing(!!r.is_followed);
      message.success(r.is_followed ? '关注成功' : '已取消关注');
    } catch (error: any) {
      const msg = error?.response?.data?.message || '操作失败';
      message.error(msg);
    } finally {
      setFollowLoading(false);
    }
  };

  const handleToggleLike = async () => {
    if (!video) return;
    const isAuth = !!localStorage.getItem('access_token');
    if (!isAuth) {
      message.warning('请先登录');
      navigate('/login');
      return;
    }
    try {
      const res = await apiService.toggleLike(video.video_uuid);
      if (typeof res.liked === 'boolean') setLiked(res.liked);
      if (typeof res.like_count === 'number') setLikeCount(res.like_count);
    } catch (error: any) {
      const msg = error?.response?.data?.message || error?.message || '操作失败';
      message.error(msg);
    }
  };

  const fetchUserInfos = useCallback(async (uuids: string[]) => {
    const missing = uuids.filter((u) => u && !userInfoMap[u]);
    if (missing.length === 0) return;
    const updates: Record<string, UserBasicInfo> = {};
    await Promise.all(
      missing.map(async (uuid) => {
        try {
          const info = await apiService.getUserBasicInfo(uuid);
          updates[uuid] = info;
        } catch {}
      })
    );
    if (Object.keys(updates).length > 0) {
      setUserInfoMap((prev) => ({ ...prev, ...updates }));
    }
  }, [userInfoMap]);

  const fetchComments = useCallback(async (page = 1, sort: 'hot' | 'time' = 'hot', targetVideoUuid?: string) => {
    if (!video_uuid && !targetVideoUuid) return;
    setCommentLoading(true);
    try {
      const videoId = targetVideoUuid || video_uuid!;
      const res = await apiService.listComments(videoId, { page, size: 20, sort_by: sort });
      const list = res.list.map((c) => {
        const rootId = c.root_uuid || c.comment_uuid;
        return { ...c, root_uuid: rootId, replies: [], repliesLoaded: false, replyPage: 1, replyTotal: c.reply_count ?? 0, repliesLoading: false };
      });
      setComments(list);
      setCommentTotal((prev) => (prev === 0 ? res.total : prev));
      setCommentSort(sort);
      await fetchUserInfos(list.map((c) => c.user_uuid));
    } catch (error: any) {
      const msg = error?.response?.data?.message || error?.message || '加载评论失败';
      message.error(msg);
    } finally {
      setCommentLoading(false);
    }
  }, [fetchUserInfos, message, video_uuid]);

  const ensureAuthed = () => {
    const isAuth = !!localStorage.getItem('access_token');
    if (!isAuth) {
      message.warning('请先登录');
      navigate('/login');
      return false;
    }
    return true;
  };

  const mergeRepliesChrono = (existing: CommentWithReplies[] = [], incoming: CommentWithReplies[] = []) => {
    const map = new Map<string, CommentWithReplies>();
    existing.forEach((item) => map.set(item.comment_uuid, item));
    incoming.forEach((item) => map.set(item.comment_uuid, item));
    return Array.from(map.values()).sort((a, b) => Number(a.created_at || 0) - Number(b.created_at || 0));
  };

  const handleReplyClick = (root: CommentWithReplies, target: CommentWithReplies) => {
    setReplyingTo({ rootId: root.root_uuid || root.comment_uuid, target });
    setReplyContent('');
  };

  const handleToggleLikeAny = async (target: CommentWithReplies, root: CommentWithReplies) => {
    if (!ensureAuthed()) return;
    try {
      const res = await apiService.toggleCommentLike(target.comment_uuid);
      setComments((prev) =>
        prev.map((c) => {
          if (c.comment_uuid !== root.comment_uuid) return c;
          if (target.comment_uuid === root.comment_uuid) {
            return { ...c, liked: !!res.liked, like_count: res.like_count ?? c.like_count };
          }
          const replies = (c.replies || []).map((r) =>
            r.comment_uuid === target.comment_uuid ? { ...r, liked: !!res.liked, like_count: res.like_count ?? r.like_count } : r
          );
          return { ...c, replies };
        })
      );
    } catch (error: any) {
      const msg = error?.response?.data?.message || error?.message || '点赞失败';
      message.error(msg);
    }
  };

  const fetchRepliesAny = async (root: CommentWithReplies, page = 1) => {
    if (!video_uuid) return;
    const rootId = root.root_uuid || root.comment_uuid;
    setComments((prev) => prev.map((c) => (c.comment_uuid === root.comment_uuid ? { ...c, repliesLoading: true } : c)));
    try {
      const res = await apiService.listComments(video_uuid, { root_uuid: rootId, page, size: REPLY_PAGE_SIZE, sort_by: 'time' });
      const children = res.list.map((child) => ({ ...child, root_uuid: rootId } as CommentWithReplies));
      setComments((prev) =>
        prev.map((c) => {
          if (c.comment_uuid !== root.comment_uuid) return c;
          return {
            ...c,
            replies: mergeRepliesChrono([], children),
            repliesLoaded: true,
            repliesLoading: false,
            replyPage: res.page,
            replyTotal: res.total,
          };
        })
      );
      await fetchUserInfos(res.list.map((c) => c.user_uuid));
    } catch (error: any) {
      const msg = error?.response?.data?.message || error?.message || '加载回复失败';
      message.error(msg);
      setComments((prev) => prev.map((c) => (c.comment_uuid === root.comment_uuid ? { ...c, repliesLoading: false } : c)));
    }
  };

  const handleSubmitComment = async () => {
    if (!video || !commentInput.trim()) {
      message.warning('请输入评论内容');
      return;
    }
    if (!ensureAuthed()) return;
    setCommentSubmitting(true);
    try {
      const res = await apiService.addComment({ video_uuid: video.video_uuid, content: commentInput.trim() });
      const rootId = res.root_uuid || res.comment_uuid;
      setComments((prev) => [{ ...res, root_uuid: rootId, replies: [], repliesLoaded: false, replyTotal: res.reply_count ?? 0, repliesLoading: false }, ...prev]);
      setCommentTotal((prev) => prev + 1);
      setCommentInput('');
      await fetchUserInfos([res.user_uuid]);
      // refresh from server to reflect sorting/hot counts without整页刷新
      fetchComments(1, commentSort, video.video_uuid);
      message.success('评论成功');
    } catch (error: any) {
      const msg = error?.response?.data?.message || error?.message || '评论失败';
      message.error(msg);
    } finally {
      setCommentSubmitting(false);
    }
  };

  const handleSubmitReply = async (parent: CommentWithReplies) => {
    if (!video || !replyContent.trim()) {
      message.warning('请输入回复内容');
      return;
    }
    if (!ensureAuthed()) return;
    setCommentSubmitting(true);
    try {
      const res = await apiService.addComment({ video_uuid: video.video_uuid, content: replyContent.trim(), parent_uuid: parent.comment_uuid });
      const rootId = parent.root_uuid || parent.comment_uuid;
      let targetPage = 1;
      setComments((prev) =>
        prev.map((c) => {
          if (c.root_uuid !== rootId && c.comment_uuid !== rootId) return c;
          const newReply = { ...res, root_uuid: rootId } as CommentWithReplies;
          const newTotal = (c.replyTotal ?? c.reply_count ?? 0) + 1;
          targetPage = Math.max(1, Math.ceil(newTotal / REPLY_PAGE_SIZE));
          const replies = mergeRepliesChrono(c.replies || [], [newReply]);
          return { ...c, replyTotal: newTotal, repliesLoaded: true, replies, replyPage: targetPage };
        })
      );
      setCommentTotal((prev) => prev + 1);
      setReplyContent('');
      setReplyingTo(null);
      await fetchUserInfos([res.user_uuid]);
      // refresh current root replies to ensure最新分页
      const rootObj = comments.find((c) => c.comment_uuid === rootId || c.root_uuid === rootId);
      if (rootObj) {
        fetchRepliesAny(rootObj as CommentWithReplies, targetPage);
      }
      message.success('回复成功');
    } catch (error: any) {
      const msg = error?.response?.data?.message || error?.message || '回复失败';
      message.error(msg);
    } finally {
      setCommentSubmitting(false);
    }
  };

  const renderUserName = (uuid: string) => {
    const info = userInfoMap[uuid];
    if (!info) return `用户${uuid.slice(0, 6)}`;
    return info.nickname || info.account || `用户${uuid.slice(0, 6)}`;
  };

  const renderAvatar = (uuid: string) => {
    const info = userInfoMap[uuid];
    if (info?.avatar_url) {
      return <Avatar src={info.avatar_url} />;
    }
    return <Avatar icon={<UserOutlined />} />;
  };

  const renderCommentItem = (root: CommentWithReplies) => {
    const isReplyingRoot = replyingTo && (replyingTo.rootId === (root.root_uuid || root.comment_uuid)) && replyingTo.target.comment_uuid === root.comment_uuid;
    const parentNameCache = new Map<string, string>();
    const getParentName = (parentUUID?: string) => {
      if (!parentUUID) return '';
      if (parentNameCache.has(parentUUID)) return parentNameCache.get(parentUUID) || '';
      if (parentUUID === root.comment_uuid) {
        const name = renderUserName(root.user_uuid);
        parentNameCache.set(parentUUID, name);
        return name;
      }
      const found = (root.replies || []).find((r) => r.comment_uuid === parentUUID);
      if (found) {
        const name = renderUserName(found.user_uuid);
        parentNameCache.set(parentUUID, name);
        return name;
      }
      return '';
    };

    return (
      <div key={root.comment_uuid} style={{ display: 'flex', width: '100%', gap: 12 }}>
        {renderAvatar(root.user_uuid)}
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Text strong>{renderUserName(root.user_uuid)}</Text>
            <Text type="secondary">{formatPublishedTime(root.created_at)}</Text>
          </div>
          <div style={{ marginTop: 4, whiteSpace: 'pre-wrap' }}>{root.content}</div>
          <Space size="large" style={{ marginTop: 8 }}>
            <span onClick={() => handleToggleLikeAny(root, root)} style={{ cursor: 'pointer' }}>
              <LikeOutlined style={{ color: root.liked ? '#1677ff' : undefined, marginRight: 4 }} />
              {root.like_count ?? 0}
            </span>
            <span onClick={() => handleReplyClick(root, root)} style={{ cursor: 'pointer' }}>
              <MessageOutlined style={{ marginRight: 4 }} />
              {root.replyTotal ?? root.reply_count ?? 0} 回复
            </span>
          </Space>

          {isReplyingRoot && (
            <div style={{ marginTop: 12, marginBottom: 12 }}>
              <Input.TextArea
                value={replyContent}
                onChange={(e) => setReplyContent(e.target.value)}
                placeholder={`回复 ${renderUserName(root.user_uuid)}`}
                autoSize={{ minRows: 2, maxRows: 4 }}
              />
              <div style={{ textAlign: 'right', marginTop: 8 }}>
                <Space>
                  <Button onClick={() => { setReplyingTo(null); setReplyContent(''); }}>取消</Button>
                  <Button type="primary" loading={commentSubmitting} onClick={() => handleSubmitReply(root)}>发送</Button>
                </Space>
              </div>
            </div>
          )}

          {/* Replies */}
          <div style={{ marginTop: 12, paddingLeft: 12, borderLeft: '2px solid #f0f0f0' }}>
            {root.repliesLoaded && (root.replies || []).map((reply) => {
              const isReplying = replyingTo && replyingTo.rootId === (root.root_uuid || root.comment_uuid) && replyingTo.target.comment_uuid === reply.comment_uuid;
              const parentName = getParentName(reply.parent_uuid);
              return (
                <div key={reply.comment_uuid} style={{ marginBottom: 12, display: 'flex', gap: 12 }}>
                  {renderAvatar(reply.user_uuid)}
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <Text strong>{renderUserName(reply.user_uuid)}</Text>
                      <Text type="secondary">{formatPublishedTime(reply.created_at)}</Text>
                    </div>
                    <div style={{ marginTop: 4, whiteSpace: 'pre-wrap' }}>
                      {parentName ? <Text type="secondary" style={{ marginRight: 6 }}>回复 @{parentName}</Text> : null}
                      {reply.content}
                    </div>
                    <Space size="large" style={{ marginTop: 6 }}>
                      <span onClick={() => handleToggleLikeAny(reply, root)} style={{ cursor: 'pointer' }}>
                        <LikeOutlined style={{ color: reply.liked ? '#1677ff' : undefined, marginRight: 4 }} />
                        {reply.like_count ?? 0}
                      </span>
                      <span onClick={() => handleReplyClick(root, reply as CommentWithReplies)} style={{ cursor: 'pointer' }}>
                        <MessageOutlined style={{ marginRight: 4 }} />
                        回复
                      </span>
                    </Space>

                    {isReplying && (
                      <div style={{ marginTop: 8 }}>
                        <Input.TextArea
                          value={replyContent}
                          onChange={(e) => setReplyContent(e.target.value)}
                          placeholder={`回复 ${renderUserName(reply.user_uuid)}`}
                          autoSize={{ minRows: 2, maxRows: 4 }}
                        />
                        <div style={{ textAlign: 'right', marginTop: 6 }}>
                          <Space>
                            <Button onClick={() => { setReplyingTo(null); setReplyContent(''); }}>取消</Button>
                            <Button type="primary" loading={commentSubmitting} onClick={() => handleSubmitReply(reply as CommentWithReplies)}>发送</Button>
                          </Space>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}

            {root.replyTotal && !root.repliesLoaded ? (
              <Button
                type="link"
                size="small"
                loading={root.repliesLoading}
                onClick={() => fetchRepliesAny(root)}
              >
                展开回复 ({root.replyTotal})
              </Button>
            ) : null}

            {root.repliesLoaded && (root.replyTotal || 0) > REPLY_PAGE_SIZE ? (
              <Pagination
                size="small"
                simple
                pageSize={REPLY_PAGE_SIZE}
                current={root.replyPage || 1}
                total={root.replyTotal || 0}
                showSizeChanger={false}
                onChange={(page) => fetchRepliesAny(root, page)}
                style={{ marginTop: 4 }}
              />
            ) : null}
          </div>
        </div>
      </div>
    );
  };

  return (
    <Layout style={{ minHeight: '100vh', background: '#f7f8fa' }}>
      <Content style={{ padding: '24px 16px' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <Space size="middle" style={{ marginBottom: 16 }}>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/')}>返回首页</Button>
            <Title level={4} style={{ margin: 0 }}>{video?.title || '播放页面'}</Title>
          </Space>

          {loading ? (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 480 }}>
              <Spin size="large" />
            </div>
          ) : video && video.video_url ? (
            <>
              <div style={{
                position: 'relative',
                width: '100%',
                paddingTop: '56.25%',
                background: '#000',
                borderRadius: 12,
                overflow: 'hidden',
                boxShadow: '0 8px 24px rgba(0,0,0,0.15)'
              }}>
                <div style={{ position: 'absolute', inset: 0 }}>
                  <VideoPlayer src={video.video_url} autoPlay />
                </div>
              </div>

              {/* Uploader Info & Description */}
              <div style={{ marginTop: 24, background: '#fff', padding: 24, borderRadius: 12, boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 16, cursor: 'pointer' }} onClick={() => window.location.href = `/user/${video.user_uuid}`}>
                    <Avatar
                      size={48}
                      src={video.uploader_avatar_url}
                      icon={<UserOutlined />}
                    />
                    <div>
                    <Title level={5} style={{ margin: 0 }}>{video.uploader_account || '创作者'}</Title>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      发布于 {formatPublishedTime(video.published_at)}
                    </Text>
                  </div>
                  </div>

                  {!isOwner && (
                    following ? (
                      <Dropdown
                        trigger={['click']}
                        menu={{
                          items: [
                            { key: 'unfollow', label: '取消关注' },
                          ],
                          onClick: async ({ key }) => {
                            if (key === 'unfollow') {
                              await handleUnfollow();
                            }
                          },
                        }}
                      >
                        <Button type="default" icon={<MenuOutlined />} loading={followLoading}>
                          已关注
                        </Button>
                      </Dropdown>
                    ) : (
                      <Button
                        type="primary"
                        icon={<PlusOutlined />}
                        onClick={handleFollow}
                        loading={followLoading}
                      >
                        关注
                      </Button>
                    )
                  )}
                </div>

                <div style={{ paddingTop: 16, borderTop: '1px solid #f0f0f0' }}>
                  <Title level={5} style={{ fontSize: 16, marginBottom: 8 }}>{video.title}</Title>
                  <Paragraph style={{ color: '#666', whiteSpace: 'pre-wrap' }}>
                    {video.description || '暂无简介'}
                  </Paragraph>
                  <div style={{ marginTop: 12 }}>
                    <Space size={[8, 8]} wrap>
                      {video.tags?.map(tag => (
                        <Tag key={tag} color="blue">#{tag}</Tag>
                      ))}
                    </Space>
                  </div>
                </div>

                <div style={{ marginTop: 16, display: 'flex', gap: 12, alignItems: 'center' }}>
                  <Button
                    icon={<LikeOutlined />}
                    type={liked ? 'primary' : 'default'}
                    onClick={handleToggleLike}
                  >
                    点赞 {likeCount}
                  </Button>
                </div>
              </div>

              {/* Comments */}
              <div style={{ marginTop: 16, background: '#fff', padding: 24, borderRadius: 12, boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                  <Space size={8}>
                    <Title level={5} style={{ margin: 0 }}>
                      评论 <Text type="secondary">({commentTotal})</Text>
                    </Title>
                  </Space>
                  <Tabs
                    size="small"
                    activeKey={commentSort}
                    onChange={(key) => fetchComments(1, key as 'hot' | 'time')}
                    items={[
                      { key: 'hot', label: <span><FireOutlined /> 最热</span> },
                      { key: 'time', label: <span><ClockCircleOutlined /> 最新</span> },
                    ]}
                  />
                </div>

                <div style={{ marginBottom: 16 }}>
                  <Input.TextArea
                    value={commentInput}
                    onChange={(e) => setCommentInput(e.target.value)}
                    placeholder="发一条友善的评论吧~"
                    autoSize={{ minRows: 2, maxRows: 6 }}
                  />
                  <div style={{ textAlign: 'right', marginTop: 8 }}>
                    <Button type="primary" loading={commentSubmitting} onClick={handleSubmitComment}>发表评论</Button>
                  </div>
                </div>

                <List
                  loading={commentLoading}
                  dataSource={comments}
                  locale={{ emptyText: commentLoading ? '加载中...' : '还没有评论，抢沙发吧' }}
                  renderItem={(item) => (
                    <List.Item key={item.comment_uuid} style={{ alignItems: 'flex-start' }}>
                      {renderCommentItem(item as CommentWithReplies)}
                    </List.Item>
                  )}
                />
              </div>
            </>
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 480 }}>
              <Empty description="暂无可播放的视频" />
            </div>
          )}
        </div>
      </Content>
    </Layout>
  );
};

export default Watch;
