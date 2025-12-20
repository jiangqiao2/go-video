import { useEffect } from 'react';
import { useAuthStore } from '@/store/auth';
import { useNotificationStore } from '@/store/notifications';
import { setNotificationStreamAuth, subscribeNotificationEvents } from '@/services/notificationStream';

export function useNotificationStream(enabled = true) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const user = useAuthStore((s) => s.user);
  const fetchNotifications = useNotificationStore((s) => s.fetchNotifications);

  useEffect(() => {
    if (!enabled || !user) {
      setNotificationStreamAuth(null, null);
      return;
    }

    const token = accessToken ?? localStorage.getItem('access_token');
    const userUuid = user?.user_uuid ?? localStorage.getItem('user_uuid');
    setNotificationStreamAuth(token, userUuid);

    const unsubscribe = subscribeNotificationEvents(() => {
      fetchNotifications().catch(() => undefined);
    });

    return () => {
      unsubscribe();
    };
  }, [enabled, accessToken, user, fetchNotifications]);
}

