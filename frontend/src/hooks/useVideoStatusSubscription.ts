import { useEffect } from 'react';
import { VideoDetail } from '@/types/api';
import { useAuthStore } from '@/store/auth';
import { setVideoStatusAccessToken, subscribeVideoStatus } from '@/services/sse';

export type VideoStatusListener = (video: VideoDetail) => void;

export function useVideoStatusSubscription(listener: VideoStatusListener, enabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  useEffect(() => {
    if (!enabled) {
      setVideoStatusAccessToken(null);
      return;
    }

    const token = accessToken ?? localStorage.getItem('access_token');
    setVideoStatusAccessToken(token);

    const unsubscribe = subscribeVideoStatus(listener);
    return () => {
      unsubscribe();
    };
  }, [accessToken, enabled, listener]);
}
