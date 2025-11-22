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
  UploadChunkRequest,
  MergeChunkRequest,
  UploadVideoStoragePathRequest,
  PublishVideoRequest,
  VideoDetail,
  VideoListResponse,
  UploadVideoStatusResponse,
} from '@/types/api';
import { arrayBufferToBase64 } from '@/utils/crypto';

class ApiService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: '/api',
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
      (error) => {
        if (error.response?.status === 401) {
          // 清除本地存储的认证信息
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('user_uuid');
          // 可以在这里跳转到登录页面
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // 用户注册
  async register(data: UserRegisterRequest): Promise<UserRegisterResponse> {
    const response = await this.api.post<ApiResponse<UserRegisterResponse>>('/v1/open/users/register', data);
    return response.data.data!;
  }

  // 用户登录
  async login(data: UserLoginRequest): Promise<UserLoginResponse> {
    const response = await this.api.post<ApiResponse<UserLoginResponse>>('/v1/open/users/login', data);
    const result = response.data.data!;
    
    // 保存认证信息到本地存储
    localStorage.setItem('access_token', result.access_token);
    localStorage.setItem('refresh_token', result.refresh_token);
    localStorage.setItem('user_uuid', result.user_uuid);
    
    return result;
  }

  // 获取用户信息
  async getUserInfo(): Promise<UserInfoResponse> {
    const response = await this.api.get<ApiResponse<UserInfoResponse>>('/v1/inner/users/me');
    return response.data.data!;
  }

  // 初始化视频上传
  async initVideoUpload(data: UploadVideoInitRequest): Promise<UploadVideoInfo> {
    const response = await this.api.post<ApiResponse<UploadVideoInfo>>('/v1/inner/upload/init', data);
    return response.data.data!;
  }

  // 上传视频分片
  async uploadChunk(data: UploadChunkRequest, options?: { signal?: AbortSignal }): Promise<void> {
    const payload = {
      chunk_uuid: data.chunk_uuid,
      user_uuid: data.user_uuid,
      upload_video_uuid: data.upload_video_uuid,
      chunk_size: data.chunk_size,
      chunk_index: data.chunk_index,
      chunk_hash: data.chunk_hash,
      chunk_data: arrayBufferToBase64(data.chunk_data),
    };

    await this.api.post('/v1/inner/upload/chunk', payload, {
      signal: options?.signal,
    });
  }

  // 合并分片
  async mergeChunks(data: MergeChunkRequest): Promise<void> {
    await this.api.post('/v1/inner/upload/merge', data);
  }

  async getUploadStatus(params: { upload_video_uuid: string; user_uuid: string }): Promise<UploadVideoStatusResponse> {
    const response = await this.api.get<ApiResponse<UploadVideoStatusResponse>>('/v1/inner/upload/status', {
      params,
    });
    return response.data.data!;
  }

  // 获取存储路径
  async getStoragePath(params: UploadVideoStoragePathRequest): Promise<string> {
    const response = await this.api.get<ApiResponse<{ storage_path: string }>>('/v1/inner/upload/chunk', {
      params,
    });
    return response.data.data!.storage_path;
  }

  // 发布视频
  async publishVideo(data: PublishVideoRequest): Promise<VideoDetail> {
    const response = await this.api.post<ApiResponse<VideoDetail>>('/v1/inner/videos', data);
    return response.data.data!;
  }

  // 获取用户视频列表
  async listUserVideos(params: { page?: number; size?: number; status?: string }): Promise<VideoListResponse> {
    const response = await this.api.get<ApiResponse<VideoListResponse>>('/v1/inner/videos', {
      params,
    });
    return response.data.data!;
  }

  // 健康检查
  async healthCheck(): Promise<any> {
    const response = await this.api.get('/health');
    return response.data;
  }
}

export const apiService = new ApiService();
export default apiService;
