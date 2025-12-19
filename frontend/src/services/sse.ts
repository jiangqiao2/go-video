import { VideoDetail } from '@/types/api';

export type VideoStatusListener = (video: VideoDetail) => void;

// SSE-based video status updates are currently disabled.
// Keep a no-op API so existing callers do not break.
export function setVideoStatusAccessToken(_token: string | null | undefined) {
  // no-op
}

export function subscribeVideoStatus(_listener: VideoStatusListener) {
  // return a no-op unsubscribe function
  return () => {};
}

