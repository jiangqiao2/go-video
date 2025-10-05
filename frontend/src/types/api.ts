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
  message: string;
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
  message: string;
}

export interface UserInfoResponse {
  user_uuid: string;
  account: string;
  message: string;
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
  chunk_size: number;
  status: string;
  storage_path: string;
}

export interface UploadVideoInfo {
  upload_video_uuid: string;
  file_name: string;
  file_size: number;
  total_chunks: number;
  uploaded_chunks: number;
  status: string;
  storage_path: string;
  chunk_storage_path: string;
  chunks: UploadChunkInfo[];
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

// 错误类型
export interface ApiError {
  code: number;
  message: string;
  details?: string;
}