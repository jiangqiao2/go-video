import { NotificationListResponse } from '@/types/api';
import { EventSourcePolyfill } from 'event-source-polyfill';

export type NotificationEventListener = () => void;

class NotificationStream {
  private eventSource: EventSource | null = null;
  private listeners: Set<NotificationEventListener> = new Set();
  private accessToken: string | null = null;
  private userUuid: string | null = null;

  setAuth(accessToken: string | null | undefined, userUuid: string | null | undefined) {
    const token = accessToken || null;
    const uuid = userUuid || null;
    if (this.accessToken === token && this.userUuid === uuid) {
      return;
    }
    this.accessToken = token;
    this.userUuid = uuid;
    this.restart();
  }

  subscribe(listener: NotificationEventListener) {
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

    if (typeof window === 'undefined') {
      // SSR or非常旧的环境，直接跳过 SSE。
      console.warn('[SSE] 当前环境不支持 EventSource，无法订阅通知变更');
      return;
    }

    const url = `/api/notification/v1/inner/notifications/stream`;

    try {
      const es = new EventSourcePolyfill(url, {
        headers: {
          Authorization: `Bearer ${this.accessToken}`,
        },
        withCredentials: false,
      });
      const handler = this.handleEvent as EventListener;
      es.addEventListener('notification.created', handler);
      es.addEventListener('notification.updated', handler);
      es.onerror = (event) => {
        // rely on EventSource built-in reconnect; just log for debugging.
        console.warn('[SSE] 通知流连接异常，等待自动重连', event);
      };
      this.eventSource = es as unknown as EventSource;
    } catch (error) {
      console.error('[SSE] 建立通知流失败', error);
    }
  }

  private disconnect() {
    if (this.eventSource) {
      const handler = this.handleEvent as EventListener;
      this.eventSource.removeEventListener('notification.created', handler);
      this.eventSource.removeEventListener('notification.updated', handler);
      this.eventSource.close();
      this.eventSource = null;
    }
  }

  private handleEvent = (event: MessageEvent<NotificationListResponse>) => {
    // 打印一行日志，帮助线上排查 SSE 是否收到事件以及事件类型。
    try {
      // 某些浏览器/Polyfill 下 event.data 可能已经是对象，也可能是字符串。
      const raw = event.data as any;
      let parsed: unknown = raw;
      if (typeof raw === 'string') {
        try {
          parsed = JSON.parse(raw);
        } catch {
          parsed = raw;
        }
      }
      // 这里的 event.type 对应 addEventListener 时注册的事件名。
      console.log('[SSE] notification event received', event.type || 'message', parsed);
    } catch {
      // 日志解析失败不影响后续逻辑。
    }
    // 具体数据前端不直接使用，只作为“有变更”的信号，触发刷新接口。
    this.listeners.forEach((listener) => listener());
  };
}

const notificationStream = new NotificationStream();

export function setNotificationStreamAuth(accessToken: string | null | undefined, userUuid: string | null | undefined) {
  notificationStream.setAuth(accessToken, userUuid);
}

export function subscribeNotificationEvents(listener: NotificationEventListener) {
  return notificationStream.subscribe(listener);
}
