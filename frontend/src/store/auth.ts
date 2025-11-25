import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { UserLoginResponse, UserInfoResponse } from '@/types/api';
import apiService from '@/services/api';

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
          
          const toAvatar = (u?: string) =>
            u && u.startsWith('http') ? u : (u ? `/storage/image/${u}` : undefined);

          set({
            isAuthenticated: true,
            accessToken: response.access_token,
            refreshToken: response.refresh_token,
            user: {
              user_uuid: response.user_uuid,
              account: response.account,
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
        // 清除本地存储
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
          const toAvatar = (u?: string) =>
            u && u.startsWith('http') ? u : (u ? `/storage/image/${u}` : undefined);
          const normalized = {
            ...userInfo,
            avatar_url: toAvatar(userInfo.avatar_url),
          };
          set({ user: normalized });
        } catch (error) {
          console.error('Failed to refresh user info:', error);
          // 如果获取用户信息失败，可能是token过期，执行登出
          get().logout();
          throw error;
        }
      },

      setAuth: (authData: UserLoginResponse) => {
        const toAvatar = (u?: string) =>
          u && u.startsWith('http') ? u : (u ? `/storage/image/${u}` : undefined);

        set({
          isAuthenticated: true,
          accessToken: authData.access_token,
          refreshToken: authData.refresh_token,
          user: {
            user_uuid: authData.user_uuid,
            account: authData.account,
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
