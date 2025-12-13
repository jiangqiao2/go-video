import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { UserLoginResponse, UserInfoResponse } from '@/types/api';
import apiService from '@/services/api';

const toAvatar = (u?: string) => {
  if (!u) return undefined;
  if (u.startsWith('http')) return u;
  const base = import.meta.env.VITE_ASSET_BASE || window.location.origin;
  return `${base}/${u.replace(/^\/+/, '')}`;
};

interface AuthState {
  isAuthenticated: boolean;
  user: UserInfoResponse | null;
  accessToken: string | null;
  refreshToken: string | null;

  // Actions
  login: (credentials: { account: string; password: string }) => Promise<void>;
  register: (userData: { account: string; password: string }) => Promise<void>;
  logout: () => void;
  refreshUserInfo: () => Promise<void>;
  setAuth: (authData: UserLoginResponse) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      isAuthenticated: false,
      user: null,
      accessToken: null,
      refreshToken: null,

      login: async (credentials) => {
        try {
          const response = await apiService.login(credentials);

          set({
            isAuthenticated: true,
            accessToken: response.access_token,
            refreshToken: response.refresh_token,
            user: {
              user_uuid: response.user_uuid,
              account: response.account,
              nickname: response.nickname,
              avatar_url: toAvatar(response.avatar_url),
            },
          });
        } catch (error) {
          console.error('Login failed:', error);
          throw error;
        }
      },

      register: async (userData) => {
        try {
          await apiService.register(userData);
          // 注册成功后自动登录
          await get().login(userData);
        } catch (error) {
          console.error('Registration failed:', error);
          throw error;
        }
      },

      logout: () => {
        // 优先使用内存中的 refresh_token，没有则回退到 localStorage
        const state = get();
        const refreshToken = state.refreshToken || localStorage.getItem('refresh_token');

        if (refreshToken) {
          // 异步调用后端退出接口，删除对应的刷新令牌
          void apiService.logout(refreshToken).catch((err) => {
            console.error('Logout api failed:', err);
          });
        }

        // 清除本地存储和内存中的登录状态
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user_uuid');

        set({
          isAuthenticated: false,
          user: null,
          accessToken: null,
          refreshToken: null,
        });
      },

      refreshUserInfo: async () => {
        try {
          const userInfo = await apiService.getUserInfo();
          const normalized = {
            ...userInfo,
            avatar_url: toAvatar(userInfo.avatar_url),
          };
          set({ user: normalized });
        } catch (error) {
          console.error('Failed to refresh user info:', error);
          const status = (error as any)?.response?.status;
          if (status === 401) {
            get().logout();
          }
          throw error;
        }
      },

      setAuth: (authData: UserLoginResponse) => {
        set({
          isAuthenticated: true,
          accessToken: authData.access_token,
          refreshToken: authData.refresh_token,
          user: {
            user_uuid: authData.user_uuid,
            account: authData.account,
            nickname: authData.nickname,
            avatar_url: toAvatar(authData.avatar_url),
          },
        });
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    }
  )
);
