import axios, { AxiosInstance, AxiosResponse } from 'axios';
import {
  ApiResponse,
  UserRegisterRequest,
  UserRegisterResponse,
  UserLoginRequest,
  UserLoginResponse,
  UserInfoResponse,
  UploadVideoInitRequest,
  UploadVideoInfo,
  UploadVideoStoragePathRequest,
  PublishVideoRequest,
  VideoDetail,
  VideoListResponse,
  UploadVideoStatusResponse,
  PresignImageRequest,
  PresignImageResponse,
  UploadImageRequest,
  UploadImageResponse,
  PresignChunkUploadRequest,
  PresignChunkUploadResponse,
  CompleteChunkRequest,
  TagListResponse,
  UserProfile,
  UserBasicInfo,
  UserRelationStat,
  CommentItem,
  CommentListResponse,
  NotificationItem,
  NotificationListResponse,
} from '@/types/api';

class ApiService {
  private api: AxiosInstance;
  private refreshing: boolean = false;
  private pending: Array<(token: string) => void> = [];

  private assetBase(): string {
    const base = import.meta.env.VITE_ASSET_BASE || window.location.origin;
    return String(base).replace(/\/$/, '');
  }

  private normalizeAsset(url?: string): string | undefined {
    if (!url) return undefined;
    let u = String(url).trim();
    const base = this.assetBase();
    if (/^https?:\/\//i.test(u)) {
      if (u.startsWith(base + '/')) return u.slice(base.length);
      try {
        const abs = new URL(u);
        const b = new URL(base);
        if (abs.origin === b.origin) return abs.pathname + abs.search + abs.hash;
      } catch {}
      return u;
    }
    u = u.replace(/(\/storage\/image\/)(?:storage\/image\/)+/g, '$1');
    u = '/' + u.replace(/^\/+/, '');
    return u;
  }

  toAssetUrl(url?: string): string | undefined {
    const u = this.normalizeAsset(url);
    if (!u) return undefined;
    if (/^https?:\/\//i.test(u)) return u;
    return `${this.assetBase()}${u}`;
  }

  constructor() {
    const API_BASE = import.meta.env.VITE_API_BASE || '/api';

    this.api = axios.create({
      baseURL: API_BASE,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器
    this.api.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('access_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.api.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        return response;
      },
      async (error) => {
        const status = error.response?.status;
        const original = error.config as any;
        if (status === 401 && !original.__isRetryRequest) {
          const refreshToken = localStorage.getItem('refresh_token');
          if (!refreshToken) {
            this.clearAuth();
            window.location.href = '/login';
            return Promise.reject(error);
          }
          if (this.refreshing) {
            return new Promise((resolve, reject) => {
              this.pending.push((token: string) => {
                original.__isRetryRequest = true;
                original.headers = original.headers || {};
                original.headers.Authorization = `Bearer ${token}`;
                this.api.request(original).then(resolve).catch(reject);
              });
            });
          }
          this.refreshing = true;
          try {
            const tokenData = await this.refreshToken(refreshToken);
            localStorage.setItem('access_token', tokenData.access_token);
            localStorage.setItem('refresh_token', tokenData.refresh_token);
            const newToken = tokenData.access_token;
            this.pending.forEach(cb => cb(newToken));
            this.pending = [];
            original.__isRetryRequest = true;
            original.headers = original.headers || {};
            original.headers.Authorization = `Bearer ${newToken}`;
            return this.api.request(original);
          } catch (e) {
            this.clearAuth();
            window.location.href = '/login';
            return Promise.reject(e);
          } finally {
            this.refreshing = false;
          }
        }
        return Promise.reject(error);
      }
    );
  }

  // 用户注册
  async register(data: UserRegisterRequest): Promise<UserRegisterResponse> {
    const response = await this.api.post<ApiResponse<UserRegisterResponse>>('/user/v1/open/users/register', data);
    return response.data.data!;
  }

  // 用户登录
  async login(data: UserLoginRequest): Promise<UserLoginResponse> {
    const response = await this.api.post<ApiResponse<UserLoginResponse>>('/user/v1/open/users/login', data);
    const result = response.data.data!;

    // 保存认证信息到本地存储
    localStorage.setItem('access_token', result.access_token);
    localStorage.setItem('refresh_token', result.refresh_token);
    localStorage.setItem('user_uuid', result.user_uuid);

    return result;
  }

  // 用户退出登录（删除刷新令牌）
  async logout(refreshToken: string): Promise<void> {
    await this.api.post('/user/v1/open/users/logout', {
      refresh_token: refreshToken,
    });
  }

  async refreshToken(refreshToken: string): Promise<{ access_token: string; refresh_token: string; expires_in: number }> {
    const response = await this.api.post<ApiResponse<{ access_token: string; refresh_token: string; expires_in: number }>>('/user/v1/open/users/refresh', { refresh_token: refreshToken });
    return response.data.data!;
  }

  // 获取用户信息
  async getUserInfo(): Promise<UserInfoResponse> {
    const response = await this.api.get<ApiResponse<UserInfoResponse>>('/user/v1/inner/users/me');
    return response.data.data!;
  }

  // 保存用户信息（部分字段）
  async saveUserInfo(data: { avatar_url?: string; account?: string; nickname?: string }): Promise<UserInfoResponse> {
    const response = await this.api.post<ApiResponse<UserInfoResponse>>('/user/v1/inner/users/save', data);
    return response.data.data!;
  }

  async changePassword(data: { old_password: string; new_password: string }): Promise<void> {
    await this.api.post('/user/v1/inner/users/password', data);
  }

  // 初始化视频上传
  async initVideoUpload(data: UploadVideoInitRequest): Promise<UploadVideoInfo> {
    const response = await this.api.post<ApiResponse<UploadVideoInfo>>('/upload/v1/inner/init', data);
    return response.data.data!;
  }

  // 上传视频分片
  async presignChunkUpload(data: PresignChunkUploadRequest): Promise<PresignChunkUploadResponse> {
    const response = await this.api.post<ApiResponse<PresignChunkUploadResponse>>('/upload/v1/inner/chunk/presign', data);
    return response.data.data!;
  }

  async completeChunkUpload(data: CompleteChunkRequest): Promise<void> {
    await this.api.post('/upload/v1/inner/chunk/complete', data);
  }

  async getUploadStatus(params: { upload_video_uuid: string; user_uuid: string }): Promise<UploadVideoStatusResponse> {
    const response = await this.api.get<ApiResponse<UploadVideoStatusResponse>>('/upload/v1/inner/status', {
      params,
    });
    return response.data.data!;
  }

  // 获取存储路径
  async getStoragePath(params: UploadVideoStoragePathRequest): Promise<string> {
    const response = await this.api.get<ApiResponse<{ storage_path: string }>>('/upload/v1/inner/chunk', {
      params,
    });
    return response.data.data!.storage_path;
  }

  // 发布视频
  async publishVideo(data: PublishVideoRequest): Promise<VideoDetail> {
    const response = await this.api.post<ApiResponse<VideoDetail>>('/upload/v1/inner/videos', data);
    return response.data.data!;
  }

  // 获取用户视频列表
  async listUserVideos(params: { page?: number; size?: number; status?: string }): Promise<VideoListResponse> {
    const response = await this.api.get<ApiResponse<any>>('/upload/v1/inner/videos', {
      params,
    });
    const data = response.data.data || {};
    const list: Array<any> = Array.isArray(data.videos || data.list) ? (data.videos || data.list) : [];
    const videos: VideoDetail[] = list.map((item) => ({
      video_uuid: item.video_uuid || item.VideoUUID || '',
      upload_video_uuid: item.upload_video_uuid || item.UploadVideo || '',
      user_uuid: item.user_uuid || item.UserUUID || '',
      title: item.title || item.Title || '',
      description: item.description || item.Description || '',
      tags: item.tags || [],
      cover_url: this.toAssetUrl(item.cover_url || item.CoverURL || ''),
      status: item.status || item.Status || '',
      created_at: item.created_at || item.CreatedAt,
      published_at: item.published_at ? String(item.published_at) : (item.PublishedAt ? String(item.PublishedAt) : undefined),
      transcode_task_uuid: item.transcode_task_uuid || item.TranscodeTaskUUID,
      duration_seconds: item.duration_seconds || item.DurationSec || item.DurationSeconds,
      like_count: item.like_count ?? item.LikeCount,
      play_count: item.play_count ?? item.PlayCount,
      comment_count: item.comment_count ?? item.CommentCount,
    }));
    const total = data.total ?? videos.length;
    const pageNum = data.page ?? params.page ?? 1;
    const size = data.size ?? params.size ?? 20;
    const totalPages = size > 0 ? Math.max(1, Math.ceil(total / size)) : 1;
    return {
      videos,
      total,
      page: pageNum,
      size,
      total_pages: data.total_pages ?? totalPages,
    };
  }

  async listPublicVideos(params: { page?: number; size?: number; status?: string }): Promise<VideoListResponse> {
    const response = await this.api.get<ApiResponse<any>>('/video/v1/open/list', {
      params,
    });
    const data = response.data.data || {};
    const list: Array<any> = Array.isArray(data.list) ? data.list : [];
    const videos: VideoDetail[] = list.map((item) => ({
      video_uuid: item.video_uuid || item.VideoUUID || '',
      upload_video_uuid: item.upload_video_uuid || item.UploadVideo || '',
      user_uuid: item.user_uuid || item.UserUUID || '',
      title: item.title || item.Title || '',
      description: item.description || item.Description || '',
      tags: [],
      cover_url: this.toAssetUrl(item.cover_url || item.CoverURL || ''),
      status: item.status || item.Status || '',
      created_at: item.created_at || item.CreatedAt,
      published_at: item.published_at ? String(item.published_at) : (item.PublishedAt ? String(item.PublishedAt) : undefined),
      video_url: this.toAssetUrl(item.video_url || item.VideoURL || ''),
      liked: item.liked ?? item.Liked,
      like_count: item.like_count ?? item.LikeCount,
      play_count: item.play_count ?? item.PlayCount,
      comment_count: item.comment_count ?? item.CommentCount,
      duration_seconds: item.duration_seconds || item.DurationSec || item.DurationSeconds,
    }));
    const total = typeof data.total === 'number' ? data.total : videos.length;
    const page = typeof data.page === 'number' ? data.page : (params.page ?? 1);
    const size = typeof data.size === 'number' ? data.size : (params.size ?? 20);
    const total_pages = size > 0 ? Math.max(1, Math.ceil(total / size)) : 1;
    return { videos, total, page, size, total_pages };
  }

  async getVideo(videoUuid: string): Promise<VideoDetail> {
    const user_uuid = localStorage.getItem('user_uuid') || undefined;
    const response = await this.api.get<ApiResponse<any>>(`/video/v1/open/get/${videoUuid}`, {
      params: user_uuid ? { user_uuid } : undefined,
    });
    const item = response.data.data || {};
    const v: VideoDetail = {
      video_uuid: item.video_uuid || item.VideoUUID || '',
      upload_video_uuid: item.upload_video_uuid || item.UploadVideo || '',
      user_uuid: item.user_uuid || item.UserUUID || '',
      title: item.title || item.Title || '',
      description: item.description || item.Description || '',
      tags: [],
      cover_url: this.toAssetUrl(item.cover_url || item.CoverURL || ''),
      status: item.status || item.Status || '',
      created_at: item.created_at || item.CreatedAt,
      published_at: item.published_at ? String(item.published_at) : (item.PublishedAt ? String(item.PublishedAt) : undefined),
      video_url: this.toAssetUrl(item.video_url || item.VideoURL || ''),
      liked: item.liked ?? item.Liked,
      like_count: item.like_count ?? item.LikeCount,
      play_count: item.play_count ?? item.PlayCount,
      comment_count: item.comment_count ?? item.CommentCount,
      duration_seconds: item.duration_seconds || item.DurationSec || item.DurationSeconds,
    };
    return v;
  }

  async toggleLike(video_uuid: string): Promise<{ like_count?: number; liked?: boolean }> {
    const response = await this.api.post<ApiResponse<any>>('/video/v1/inner/like', { video_uuid });
    const data = response.data.data || {};
    return { like_count: data.like_count ?? data.LikeCount, liked: data.liked ?? data.Liked };
  }

  async listComments(videoUuid: string, params?: { page?: number; size?: number; sort_by?: 'hot' | 'time'; parent_uuid?: string; root_uuid?: string }): Promise<CommentListResponse> {
    const response = await this.api.get<ApiResponse<any>>(`/video/v1/open/comments/${videoUuid}`, {
      params,
    });
    const data = response.data.data || {};
    const list: Array<any> = Array.isArray(data.list) ? data.list : [];
    const items: CommentItem[] = list.map((item) => ({
      comment_uuid: item.comment_uuid || item.CommentUUID || '',
      root_uuid: item.root_uuid || item.RootUUID || item.comment_uuid || item.CommentUUID,
      video_uuid: item.video_uuid || item.VideoUUID || videoUuid,
      user_uuid: item.user_uuid || item.UserUUID || '',
      content: item.content || item.Content || '',
      parent_uuid: item.parent_uuid || item.ParentUUID || undefined,
      parent_type: item.parent_type || item.ParentType,
      depth: item.depth ?? item.Depth,
      path: item.path || item.Path,
      like_count: item.like_count ?? item.LikeCount,
      reply_count: item.reply_count ?? item.ReplyCount,
      liked: item.liked ?? item.Liked,
      created_at: item.created_at || item.CreatedAt,
    }));
    return {
      list: items,
      page: typeof data.page === 'number' ? data.page : params?.page || 1,
      size: typeof data.size === 'number' ? data.size : params?.size || 20,
      total: typeof data.total === 'number' ? data.total : items.length,
    };
  }

  async addComment(data: { video_uuid: string; content: string; parent_uuid?: string }): Promise<CommentItem> {
    const response = await this.api.post<ApiResponse<any>>('/video/v1/inner/comment', data);
    const item = response.data.data || {};
    return {
      comment_uuid: item.comment_uuid || item.CommentUUID || '',
      root_uuid: item.root_uuid || item.RootUUID || item.comment_uuid || item.CommentUUID,
      video_uuid: item.video_uuid || item.VideoUUID || data.video_uuid,
      user_uuid: item.user_uuid || item.UserUUID || '',
      content: item.content || item.Content || '',
      parent_uuid: item.parent_uuid || item.ParentUUID || undefined,
      like_count: item.like_count ?? item.LikeCount ?? 0,
      reply_count: item.reply_count ?? item.ReplyCount ?? 0,
      liked: item.liked ?? item.Liked ?? false,
      created_at: item.created_at || item.CreatedAt || Date.now(),
    };
  }

  async toggleCommentLike(comment_uuid: string): Promise<{ liked?: boolean; like_count?: number }> {
    const response = await this.api.post<ApiResponse<any>>('/video/v1/inner/comment/like', { comment_uuid });
    const data = response.data.data || {};
    return {
      liked: data.liked ?? data.Liked,
      like_count: data.like_count ?? data.LikeCount,
    };
  }

  async attachUploaderBasicInfo(videos: VideoDetail[]): Promise<VideoDetail[]> {
    const uuids = Array.from(new Set(videos.map(v => v.user_uuid).filter(Boolean)));
    const map = new Map<string, { display?: string; avatar_url?: string }>();
    for (const uuid of uuids) {
      try {
        const info = await this.getUserBasicInfo(uuid);
        const display = info.nickname || '';
        map.set(uuid, { display, avatar_url: info.avatar_url });
      } catch {}
    }
    return videos.map(v => {
      const u = map.get(v.user_uuid);
      if (!u) return v;
      return { ...v, uploader_account: u.display || '创作者', uploader_avatar_url: u.avatar_url } as VideoDetail;
    });
  }

  // 健康检查
  async healthCheck(): Promise<any> {
    const response = await this.api.get('/health');
    return response.data;
  }

  // 图片直传：获取PUT预签名
  async presignImage(data: PresignImageRequest): Promise<PresignImageResponse> {
    const payload = {
      file_name: data.file_name,
      category: data.category ?? 'avatar',
      expires_seconds: data.expires_seconds ?? 900,
    };
    const response = await this.api.post<ApiResponse<PresignImageResponse>>('/upload/v1/open/image/presign', payload);
    return response.data.data!;
  }

  // 上传图片并返回完整地址
  async uploadImage(data: UploadImageRequest): Promise<UploadImageResponse> {
    const form = new FormData();
    form.append('file', data.file);
    if (data.category) {
      form.append('category', data.category);
    }
    const response = await this.api.post<ApiResponse<UploadImageResponse>>('/upload/v1/inner/image', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    const res = response.data.data!;
    return { ...res, url: this.toAssetUrl(res.url) || '' };
  }

  // 获取标签列表
  async listTags(): Promise<TagListResponse> {
    const response = await this.api.get<ApiResponse<TagListResponse>>('/upload/v1/open/tags');
    return response.data.data!;
  }

  // 获取用户基本信息（公开）
  async getUserBasicInfo(userUuid: string): Promise<UserBasicInfo> {
    const response = await this.api.get<ApiResponse<UserBasicInfo>>(`/user/v1/open/users/${userUuid}`);
    const basic = response.data.data!;
    return {
      ...basic,
      avatar_url: this.toAssetUrl(basic.avatar_url),
      cover_url: this.toAssetUrl(basic.cover_url),
    };
  }

  // 获取用户关系统计（粉丝数、关注数、关注状态）
  async getUserRelation(userUuid: string): Promise<UserRelationStat> {
    const response = await this.api.get<ApiResponse<UserRelationStat>>(`/user/v1/open/users/${userUuid}/relation`);
    return response.data.data!;
  }

  // 组合获取完整用户Profile（前端便捷方法）
  async getUserProfile(userUuid: string): Promise<UserProfile> {
    const [basicInfo, relationStat] = await Promise.all([
      this.getUserBasicInfo(userUuid),
      this.getUserRelation(userUuid),
    ]);
    return {
      ...basicInfo,
      follower_count: relationStat?.follower_count ?? 0,
      following_count: relationStat?.following_count ?? 0,
      is_followed: !!relationStat?.is_followed,
    };
  }

  // 设置关注状态（统一接口）：follow=true 关注，false 取消关注
  async setFollow(targetUserUuid: string, follow: boolean): Promise<void> {
    await this.api.post('/user/v1/inner/relation/follow/toggle', {
      target_user_uuid: targetUserUuid,
      action: follow ? 'follow' : 'unfollow',
    });
  }

  // 兼容旧调用：关注用户
  async followUser(targetUserUuid: string): Promise<void> {
    return this.setFollow(targetUserUuid, true);
  }

  // 兼容旧调用：取消关注
  async unfollowUser(targetUserUuid: string): Promise<void> {
    return this.setFollow(targetUserUuid, false);
  }

  // 查询关注状态（需要认证）
  async getFollowStatus(targetUserUuid: string): Promise<boolean> {
    const response = await this.api.get<ApiResponse<{ following: boolean }>>('/user/v1/inner/relation/status', {
      params: { target_uuid: targetUserUuid, target_user_uuid: targetUserUuid },
    });
    return !!response.data.data?.following;
  }

  // 获取指定用户的视频列表
  async listVideosByUser(userUuid: string, params: { page?: number; size?: number }): Promise<VideoListResponse> {
    const response = await this.api.get<ApiResponse<any>>(`/video/v1/open/list`, { params: { ...params, user_uuid: userUuid } });
    const data = response.data.data || {};
    const list: Array<any> = Array.isArray(data.list) ? data.list : [];
    const videos: VideoDetail[] = list.map((item) => ({
      video_uuid: item.video_uuid || item.VideoUUID || '',
      upload_video_uuid: item.upload_video_uuid || item.UploadVideo || '',
      user_uuid: item.user_uuid || item.UserUUID || '',
      title: item.title || item.Title || '',
      description: item.description || item.Description || '',
      tags: [],
      cover_url: this.toAssetUrl(item.cover_url || item.CoverURL || ''),
      status: item.status || item.Status || '',
      published_at: item.published_at ? String(item.published_at) : (item.PublishedAt ? String(item.PublishedAt) : undefined),
      video_url: this.toAssetUrl(item.video_url || item.VideoURL || ''),
    }));
    const total = typeof data.total === 'number' ? data.total : videos.length;
    const page = typeof data.page === 'number' ? data.page : (params.page ?? 1);
    const size = typeof data.size === 'number' ? data.size : (params.size ?? 20);
    const total_pages = size > 0 ? Math.max(1, Math.ceil(total / size)) : 1;
    return { videos, total, page, size, total_pages };
  }

  // 获取当前登录用户的通知列表（站内信）
  async listNotifications(params?: { page?: number; page_size?: number }): Promise<NotificationListResponse> {
    const response = await this.api.get<ApiResponse<any>>('/notification/v1/inner/notifications', {
      params,
    });
    const data = response.data.data || {};
    const list: Array<any> = Array.isArray(data.notifications) ? data.notifications : [];
    const notifications: NotificationItem[] = list.map((n) => ({
      id: Number(n.id ?? n.ID ?? 0),
      type: n.type || n.Type || '',
      title: n.title || n.Title || '',
      content: n.content || n.Content || '',
      extra_json: n.extra_json || n.ExtraJSON || '',
      is_read: Boolean(n.is_read ?? n.IsRead),
      created_at: n.created_at || n.CreatedAt,
      read_at: n.read_at || n.ReadAt,
    }));
    return {
      notifications,
      unread_count: Number(data.unread_count ?? data.UnreadCount ?? 0),
    };
  }

  // 将指定通知标记为已读
  async markNotificationsRead(ids: number[]): Promise<void> {
    if (!ids || ids.length === 0) return;
    await this.api.post('/notification/v1/inner/notifications/read', { ids });
  }

  private clearAuth() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_uuid');
  }
}

export const apiService = new ApiService();
export default apiService;
