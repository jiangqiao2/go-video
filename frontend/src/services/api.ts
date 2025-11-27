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
} from '@/types/api';

class ApiService {
  private api: AxiosInstance;
  private refreshing: boolean = false;
  private pending: Array<(token: string) => void> = [];

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

        const userUuid = localStorage.getItem('user_uuid');
        if (userUuid) {
          config.headers['X-User-UUID'] = userUuid;
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
  async saveUserInfo(data: { avatar_url?: string }): Promise<UserInfoResponse> {
    const response = await this.api.post<ApiResponse<UserInfoResponse>>('/user/v1/inner/users/save', data);
    return response.data.data!;
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
    const response = await this.api.get<ApiResponse<VideoListResponse>>('/upload/v1/inner/videos', {
      params,
    });
    return response.data.data!;
  }

  async listPublicVideos(params: { page?: number; size?: number; status?: string }): Promise<VideoListResponse> {
    const response = await this.api.get<ApiResponse<VideoListResponse>>('/upload/v1/open/videos', {
      params,
    });
    return response.data.data!;
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
    return response.data.data!;
  }

  // 获取标签列表
  async listTags(): Promise<TagListResponse> {
    const response = await this.api.get<ApiResponse<TagListResponse>>('/upload/v1/open/tags');
    return response.data.data!;
  }

  private clearAuth() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_uuid');
  }
}

export const apiService = new ApiService();
export default apiService;
