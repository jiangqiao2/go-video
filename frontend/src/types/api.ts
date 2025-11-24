// 通用API响应结构
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
  timestamp?: number;
}

// 用户相关类型
export interface UserRegisterRequest {
  account: string;
  password: string;
}

export interface UserRegisterResponse {
  user_uuid: string;
  account: string;
}

export interface UserLoginRequest {
  account: string;
  password: string;
}

export interface UserLoginResponse {
  user_uuid: string;
  account: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface UserInfoResponse {
  user_uuid: string;
  account: string;
  avatar_url?: string;
}

// 上传相关类型
export interface UploadVideoInitRequest {
  file_name: string;
  file_size: number;
  total_chunks: number;
  user_uuid: string;
  file_hash: string;
}

export interface UploadChunkInfo {
  chunk_uuid: string;
  chunk_index: number;
  status?: string;
}

export interface UploadVideoInfo {
  upload_video_uuid: string;
  chunk_size?: number;
  total_chunks?: number;
  upload_chunks: UploadChunkInfo[];
  status: string; // 上传视频的状态：Init, Uploading, Merging, Success, Failed
}

export interface UploadChunkRequest {
  chunk_uuid: string;
  user_uuid: string;
  upload_video_uuid: string;
  chunk_size: number;
  chunk_index: number;
  chunk_data: ArrayBuffer;
  chunk_hash: string;
}

export interface MergeChunkRequest {
  upload_video_uuid: string;
  user_uuid: string;
}

export interface UploadVideoStoragePathRequest {
  user_uuid: string;
  chunk_uuid: string;
}

export interface PublishVideoRequest {
  upload_video_uuid: string;
  title: string;
  description?: string;
  tags?: string[];
  cover_url?: string;
}

export interface VideoDetail {
  video_uuid: string;
  upload_video_uuid: string;
  user_uuid: string;
  title: string;
  description?: string;
  tags: string[];
  cover_url?: string;
  status: string;
  published_at?: string;
  transcode_task_uuid?: string;
  video_url?: string;
  error_message?: string;
}

export interface VideoListResponse {
  videos: VideoDetail[];
  total: number;
  page: number;
  size: number;
  total_pages: number;
}

// 错误类型
export interface ApiError {
  code: number;
  message: string;
  details?: string;
}

export interface UploadVideoStatusResponse {
  upload_video_uuid: string;
  status: string;
}

// 图片直传相关类型
export interface PresignImageRequest {
  file_name: string;
  category?: string;
  expires_seconds?: number;
}

export interface PresignImageResponse {
  bucket: string;
  key: string;
  put_url: string;
}

export interface UploadImageRequest {
  file: File;
  category?: string;
}

export interface UploadImageResponse {
  bucket: string;
  key: string;
  url: string;
}

// 保存用户信息（当前仅支持 avatar_url）
export interface UserSaveRequest {
  avatar_url?: string;
}

// 标签相关类型
export interface TagDto {
  tag_uuid: string;
  name: string;
  code: string;
  description: string;
}

export interface TagListResponse {
  list: TagDto[];
}
