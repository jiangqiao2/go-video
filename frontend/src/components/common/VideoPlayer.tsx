import React, { useEffect, useRef } from 'react';

declare global {
  interface Window {
    Hls?: any;
    __hlsLoaderPromise?: Promise<void>;
  }
}

export interface VideoPlayerProps {
  src?: string;
  className?: string;
  autoPlay?: boolean;
}

const HLS_SCRIPT_SRC = 'https://cdn.jsdelivr.net/npm/hls.js@1.4.12/dist/hls.min.js';

const loadHlsIfNeeded = async () => {
  if (typeof window === 'undefined') return;
  if (window.Hls) {
    return;
  }
  if (!window.__hlsLoaderPromise) {
    window.__hlsLoaderPromise = new Promise<void>((resolve, reject) => {
      const script = document.createElement('script');
      script.src = HLS_SCRIPT_SRC;
      script.async = true;
      script.onload = () => resolve();
      script.onerror = () => reject(new Error('failed to load hls.js'));
      document.body.appendChild(script);
    });
  }
  try {
    await window.__hlsLoaderPromise;
  } catch (err) {
    console.warn('[VideoPlayer] load hls.js failed', err);
  }
};

const VideoPlayer: React.FC<VideoPlayerProps> = ({ src, className, autoPlay }) => {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hlsRef = useRef<any>(null);

  useEffect(() => {
    const videoEl = videoRef.current;
    if (!videoEl || !src) {
      return;
    }

    const isHLS = src.toLowerCase().endsWith('.m3u8');

    const setup = async () => {
      if (!videoEl) return;
      if (!isHLS) {
        videoEl.src = src;
        if (autoPlay) {
          try {
            await videoEl.play();
          } catch (err) {
            console.warn('[VideoPlayer] autoplay failed', err);
          }
        }
        return;
      }

      if (videoEl.canPlayType('application/vnd.apple.mpegurl')) {
        videoEl.src = src;
        if (autoPlay) {
          try {
            await videoEl.play();
          } catch (err) {
            console.warn('[VideoPlayer] autoplay failed', err);
          }
        }
        return;
      }

      await loadHlsIfNeeded();
      if (!window.Hls || !window.Hls.isSupported?.()) {
        console.warn('[VideoPlayer] hls.js not supported');
        videoEl.src = src;
        return;
      }

      const hlsInstance = new window.Hls();
      hlsRef.current = hlsInstance;
      hlsInstance.loadSource(src);
      hlsInstance.attachMedia(videoEl);
      if (autoPlay) {
        hlsInstance.on(window.Hls.Events.MANIFEST_PARSED, async () => {
          try {
            await videoEl.play();
          } catch (err) {
            console.warn('[VideoPlayer] autoplay failed', err);
          }
        });
      }
    };

    setup();

    return () => {
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
    };
  }, [src, autoPlay]);

  return (
    <video
      ref={videoRef}
      className={className}
      controls
      style={{ width: '100%' }}
      playsInline
    />
  );
};

export default VideoPlayer;
