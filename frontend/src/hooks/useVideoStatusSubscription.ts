import { VideoDetail } from '@/types/api';

export type VideoStatusListener = (video: VideoDetail) => void;

// No-op hook; SSE-based video status subscription has been removed.
export function useVideoStatusSubscription(_listener: VideoStatusListener, _enabled = true) {
  // intentionally empty
}

