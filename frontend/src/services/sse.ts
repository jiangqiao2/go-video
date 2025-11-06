import { VideoDetail } from '@/types/api';

type VideoStatusListener = (video: VideoDetail) => void;

class VideoStatusStream {
  private eventSource: EventSource | null = null;
  private listeners: Set<VideoStatusListener> = new Set();
  private accessToken: string | null = null;

  setAccessToken(token: string | null | undefined) {
    const normalized = token || null;
    if (this.accessToken === normalized) {
      return;
    }
    this.accessToken = normalized;
    this.restart();
  }

  subscribe(listener: VideoStatusListener) {
    this.listeners.add(listener);
    if (!this.eventSource && this.accessToken) {
      this.open();
    }
    return () => {
      this.listeners.delete(listener);
      if (this.listeners.size === 0) {
        this.disconnect();
      }
    };
  }

  private restart() {
    this.disconnect();
    if (this.listeners.size > 0 && this.accessToken) {
      this.open();
    }
  }

  private open() {
    if (!this.accessToken) {
      return;
    }
    if (typeof EventSource === 'undefined') {
      console.warn('[SSE] 当前环境不支持 EventSource，无法订阅视频状态');
      return;
    }
    const url = `/api/v1/stream?access_token=${encodeURIComponent(this.accessToken)}`;
    try {
      const source = new EventSource(url);
      source.addEventListener('video.status.changed', this.handleVideoStatus as EventListener);
      source.onerror = (event) => {
        // 交由 EventSource 内部自动重连策略处理，必要时可补充告警
        console.warn('[SSE] 视频状态流连接异常，等待自动重连', event);
      };
      this.eventSource = source;
    } catch (error) {
      console.error('[SSE] 建立视频状态流失败', error);
    }
  }

  private disconnect() {
    if (this.eventSource) {
      this.eventSource.removeEventListener('video.status.changed', this.handleVideoStatus as EventListener);
      this.eventSource.close();
      this.eventSource = null;
    }
  }

  private handleVideoStatus = (event: MessageEvent) => {
    try {
      const detail: VideoDetail = JSON.parse(event.data);
      this.listeners.forEach((listener) => listener(detail));
    } catch (error) {
      console.error('[SSE] 解析视频状态消息失败', error);
    }
  };
}

const videoStatusStream = new VideoStatusStream();

export function setVideoStatusAccessToken(token: string | null | undefined) {
  videoStatusStream.setAccessToken(token);
}

export function subscribeVideoStatus(listener: VideoStatusListener) {
  return videoStatusStream.subscribe(listener);
}
