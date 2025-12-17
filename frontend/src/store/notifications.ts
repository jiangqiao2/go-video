import { create } from 'zustand';
import { NotificationItem, NotificationCategory } from '@/types/api';
import apiService from '@/services/api';

export interface NotificationState {
  items: NotificationItem[];
  unreadCount: number;
  loading: boolean;
  activeCategory: NotificationCategory | 'all';

  // actions
  fetchNotifications: () => Promise<void>;
  markAsRead: (id: number) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  setActiveCategory: (category: NotificationCategory | 'all') => void;
}

// 将后端的 type 字段归类为前端 Tab 所需的类型
export const mapNotificationCategory = (type: string | undefined | null): NotificationCategory => {
  const t = String(type || '').toLowerCase();
  if (!t) return 'other';
  if (t.includes('reply')) return 'reply';
  if (t.includes('comment')) return 'comment';
  if (t.includes('mention') || t.includes('@')) return 'mention';
  if (t.includes('like') || t.includes('thumb') || t.includes('favorite')) return 'like';
  if (t.includes('system')) return 'system';
  return 'other';
};

export const useNotificationStore = create<NotificationState>((set, get) => ({
  items: [],
  unreadCount: 0,
  loading: false,
  activeCategory: 'reply',

  async fetchNotifications() {
    set({ loading: true });
    try {
      const res = await apiService.listNotifications({ page: 1, page_size: 50 });
      // 按时间倒序
      const sorted = [...res.notifications].sort((a, b) => {
        const ta = Number(new Date(a.created_at as any));
        const tb = Number(new Date(b.created_at as any));
        return tb - ta;
      });
      set({
        items: sorted,
        unreadCount: res.unread_count,
      });
    } catch (e) {
      console.error('fetchNotifications failed', e);
    } finally {
      set({ loading: false });
    }
  },

  async markAsRead(id: number) {
    const { items, unreadCount } = get();
    const target = items.find((x) => x.id === id);
    if (!target || target.is_read) return;
    try {
      await apiService.markNotificationsRead([id]);
      const next = items.map((n) => (n.id === id ? { ...n, is_read: true } : n));
      set({
        items: next,
        unreadCount: unreadCount > 0 ? unreadCount - 1 : 0,
      });
    } catch (e) {
      console.error('markAsRead failed', e);
    }
  },

  async markAllAsRead() {
    const { items, unreadCount } = get();
    const unreadIds = items.filter((n) => !n.is_read).map((n) => n.id);
    if (unreadIds.length === 0) return;
    try {
      await apiService.markNotificationsRead(unreadIds);
      const next = items.map((n) => (n.is_read ? n : { ...n, is_read: true }));
      set({
        items: next,
        unreadCount: 0,
      });
    } catch (e) {
      console.error('markAllAsRead failed', e);
      // 如果失败，就退化为本地减少未读计数，防止 UI 卡死
      if (unreadCount > 0) {
        set({ unreadCount: 0 });
      }
    }
  },

  setActiveCategory(category) {
    set({ activeCategory: category });
  },
}));
