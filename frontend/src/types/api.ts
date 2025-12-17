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
  account?: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
  avatar_url?: string;
  nickname?: string;
}

export interface TokenRefreshResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface UserInfoResponse {
  user_uuid: string;
  account?: string;
  nickname?: string;
  avatar_url?: string;
}

// 用户基本信息（公开）
export interface UserBasicInfo {
  user_uuid: string;
  account?: string;
  nickname?: string;
  avatar_url?: string;
  description?: string;
  cover_url?: string;
  created_at?: string;
}

// 用户关系统计
export interface UserRelationStat {
  user_uuid: string;
  follower_count: number;   // 粉丝数
  following_count: number;  // 关注数
  is_followed: boolean;     // 当前用户是否已关注此用户
}

// 完整的用户Profile（前端组合使用）
export interface UserProfile extends UserBasicInfo {
  follower_count: number;
  following_count: number;
  is_followed: boolean;
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
  storage_path?: string;
  put_url?: string;
  expires_seconds?: number;
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

export interface PresignChunkUploadRequest {
  chunk_uuid: string;
  upload_video_uuid: string;
  chunk_index: number;
  chunk_size: number;
  content_type: string;
}

export interface PresignChunkUploadResponse {
  upload_video_uuid: string;
  chunk_uuid: string;
  chunk_index: number;
  bucket: string;
  key: string;
  put_url: string;
  expires_seconds: number;
}

export interface CompleteChunkRequest {
  chunk_uuid: string;
  upload_video_uuid: string;
  chunk_index: number;
  chunk_size: number;
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
  created_at?: string | number;
  published_at?: string;
  transcode_task_uuid?: string;
  video_url?: string;
  duration_seconds?: number;
  like_count?: number;
  liked?: boolean;
  play_count?: number;
  comment_count?: number;
  uploader_account?: string;
  uploader_avatar_url?: string;
}

export interface CommentItem {
  comment_uuid: string;
  root_uuid?: string;
  video_uuid: string;
  user_uuid: string;
  content: string;
  parent_uuid?: string;
  parent_type?: string;
  depth?: number;
  path?: string;
  like_count?: number;
  reply_count?: number;
  liked?: boolean;
  created_at: number | string;
}

export interface CommentListResponse {
  list: CommentItem[];
  page: number;
  size: number;
  total: number;
}

// 通知相关类型（来自 notification-service）
export type NotificationCategory = 'reply' | 'mention' | 'comment' | 'system' | 'like' | 'other';

export interface NotificationItem {
  id: number;
  type: string;           // 原始类型字符串（后端自由定义，如 reply_comment、at_mention 等）
  title: string;
  content: string;
  extra_json?: string;
  is_read: boolean;
  created_at: string | number;
  read_at?: string | number;
}

export interface NotificationListResponse {
  notifications: NotificationItem[];
  unread_count: number;
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
