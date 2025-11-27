import React, { useEffect, useRef, useState, useCallback } from 'react';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
  SoundOutlined,
  MutedOutlined,
} from '@ant-design/icons';
import { Slider } from 'antd';

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

const formatTime = (seconds: number) => {
  if (!seconds || isNaN(seconds)) return '00:00';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h > 0) {
    return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  }
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
};

const VideoPlayer: React.FC<VideoPlayerProps> = ({ src, className, autoPlay }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hlsRef = useRef<any>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [buffered, setBuffered] = useState<TimeRanges | null>(null);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [showControls, setShowControls] = useState(false);
  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout>>();

  // Initialize HLS and Video
  useEffect(() => {
    const videoEl = videoRef.current;
    if (!videoEl || !src) return;

    const isHLS = src.toLowerCase().endsWith('.m3u8');

    const setup = async () => {
      if (autoPlay) {
        try {
          videoEl.muted = false;
          videoEl.volume = 1;
          (videoEl as any).autoplay = true;
        } catch {}
      }
      if (!isHLS) {
        videoEl.src = src;
      } else if (videoEl.canPlayType('application/vnd.apple.mpegurl')) {
        videoEl.src = src;
      } else {
        await loadHlsIfNeeded();
        if (window.Hls && window.Hls.isSupported?.()) {
          const hlsInstance = new window.Hls({
            maxBufferLength: 60,
            maxMaxBufferLength: 600,
            startLevel: -1,
            enableWorker: true,
          });
          hlsRef.current = hlsInstance;
          hlsInstance.loadSource(src);
          hlsInstance.attachMedia(videoEl);

          if (autoPlay) {
            hlsInstance.on(window.Hls.Events.MANIFEST_PARSED, () => {
              videoEl.play().catch(console.warn);
            });
          }
        } else {
          videoEl.src = src;
        }
      }

      if (autoPlay && !isHLS) {
        videoEl.play().catch(console.warn);
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

  // Event Listeners
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const onPlay = () => setIsPlaying(true);
    const onPause = () => setIsPlaying(false);
    const onTimeUpdate = () => {
      setCurrentTime(video.currentTime);
      setBuffered(video.buffered);
    };
    const onDurationChange = () => setDuration(video.duration);
    const onProgress = () => setBuffered(video.buffered);
    const onVolumeChange = () => {
      setVolume(video.volume);
      setIsMuted(video.muted);
    };

    video.addEventListener('play', onPlay);
    video.addEventListener('pause', onPause);
    video.addEventListener('timeupdate', onTimeUpdate);
    video.addEventListener('durationchange', onDurationChange);
    video.addEventListener('progress', onProgress);
    video.addEventListener('volumechange', onVolumeChange);

    return () => {
      video.removeEventListener('play', onPlay);
      video.removeEventListener('pause', onPause);
      video.removeEventListener('timeupdate', onTimeUpdate);
      video.removeEventListener('durationchange', onDurationChange);
      video.removeEventListener('progress', onProgress);
      video.removeEventListener('volumechange', onVolumeChange);
    };
  }, []);

  const togglePlay = useCallback(() => {
    const video = videoRef.current;
    if (!video) return;
    if (video.paused) {
      video.play();
    } else {
      video.pause();
    }
  }, []);

  const handleSeek = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    const video = videoRef.current;
    if (!video) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const percent = (e.clientX - rect.left) / rect.width;
    video.currentTime = percent * video.duration;
  }, []);

  const toggleFullscreen = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    if (!document.fullscreenElement) {
      container.requestFullscreen().then(() => setIsFullscreen(true));
    } else {
      document.exitFullscreen().then(() => setIsFullscreen(false));
    }
  }, []);

  useEffect(() => {
    const handleFsChange = () => {
      const fs = !!document.fullscreenElement;
      setIsFullscreen(fs);
    };
    document.addEventListener('fullscreenchange', handleFsChange);
    return () => {
      document.removeEventListener('fullscreenchange', handleFsChange);
    };
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const video = videoRef.current;
      if (!video) return;
      switch (e.code) {
        case 'Space':
        case 'KeyK':
          e.preventDefault();
          togglePlay();
          break;
        case 'KeyF':
          e.preventDefault();
          toggleFullscreen();
          break;
        case 'KeyM':
          e.preventDefault();
          video.muted = !video.muted;
          break;
        case 'ArrowRight':
          e.preventDefault();
          video.currentTime = Math.min(video.currentTime + 5, video.duration || video.currentTime + 5);
          break;
        case 'ArrowLeft':
          e.preventDefault();
          video.currentTime = Math.max(video.currentTime - 5, 0);
          break;
        case 'ArrowUp':
          e.preventDefault();
          video.volume = Math.min((video.volume ?? 1) + 0.1, 1);
          video.muted = video.volume === 0;
          break;
        case 'ArrowDown':
          e.preventDefault();
          video.volume = Math.max((video.volume ?? 1) - 0.1, 0);
          video.muted = video.volume === 0;
          break;
        default:
          break;
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [togglePlay, toggleFullscreen]);

  const handleMouseMove = useCallback(() => {
    setShowControls(true);
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current);
    }
    controlsTimeoutRef.current = setTimeout(() => {
      if (isPlaying) {
        setShowControls(false);
      }
    }, 3000);
  }, [isPlaying]);

  const renderBufferedRanges = () => {
    if (!buffered || !duration) return null;
    const ranges = [];
    for (let i = 0; i < buffered.length; i++) {
      const start = buffered.start(i);
      const end = buffered.end(i);
      const left = (start / duration) * 100;
      const width = ((end - start) / duration) * 100;
      ranges.push(
        <div
          key={i}
          style={{
            position: 'absolute',
            left: `${left}%`,
            width: `${width}%`,
            height: '100%',
            backgroundColor: 'rgba(255, 255, 255, 0.4)',
            borderRadius: 2,
          }}
        />
      );
    }
    return ranges;
  };

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ position: 'relative', width: '100%', backgroundColor: '#000', overflow: 'hidden' }}
      onMouseMove={handleMouseMove}
      onMouseLeave={() => setShowControls(false)}
    >
      <video
        ref={videoRef}
        style={{ width: '100%', height: '100%', display: 'block' }}
        playsInline
        preload="metadata"
        onDoubleClick={toggleFullscreen}
        onClick={togglePlay}
      />

      {/* Controls Overlay */}
      <div
        style={{
          position: 'absolute',
          bottom: 0,
          left: 0,
          right: 0,
          background: 'linear-gradient(to top, rgba(0,0,0,0.7), transparent)',
          padding: '20px 12px 12px',
          opacity: showControls || !isPlaying ? 1 : 0,
          transition: 'opacity 0.3s',
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
        }}
      >
        {/* Progress Bar */}
        <div
          style={{
            position: 'relative',
            width: '100%',
            height: 4,
            backgroundColor: 'rgba(255, 255, 255, 0.2)',
            cursor: 'pointer',
            borderRadius: 2,
            marginBottom: 8,
          }}
          onClick={handleSeek}
        >
          {/* Buffered Ranges */}
          {renderBufferedRanges()}

          {/* Play Progress */}
          <div
            style={{
              position: 'absolute',
              left: 0,
              width: `${(currentTime / duration) * 100}%`,
              height: '100%',
              backgroundColor: '#1890ff',
              borderRadius: 2,
            }}
          >
            {/* Handle */}
            <div
              style={{
                position: 'absolute',
                right: -4,
                top: -3,
                width: 10,
                height: 10,
                borderRadius: '50%',
                backgroundColor: '#fff',
                boxShadow: '0 0 4px rgba(0,0,0,0.3)',
                transform: 'scale(0)',
                transition: 'transform 0.1s',
              }}
              className="progress-handle"
            />
          </div>
        </div>

        {/* Control Buttons */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', color: '#fff' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <div onClick={togglePlay} style={{ cursor: 'pointer', fontSize: 24 }}>
              {isPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            </div>
            <div style={{ fontSize: 12, fontFamily: 'monospace' }}>
              {formatTime(currentTime)} / {formatTime(duration)}
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, width: 100 }}>
              <div onClick={() => {
                if (videoRef.current) {
                  videoRef.current.muted = !isMuted;
                }
              }} style={{ cursor: 'pointer', fontSize: 20 }}>
                {isMuted || volume === 0 ? <MutedOutlined /> : <SoundOutlined />}
              </div>
              <Slider
                min={0}
                max={1}
                step={0.1}
                value={isMuted ? 0 : volume}
                onChange={(v) => {
                  if (videoRef.current) {
                    videoRef.current.volume = v;
                    videoRef.current.muted = v === 0;
                  }
                }}
                style={{ width: 60, margin: 0 }}
                tooltip={{ open: false }}
              />
            </div>
            <div onClick={toggleFullscreen} style={{ cursor: 'pointer', fontSize: 20 }}>
              {isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
            </div>
          </div>
        </div>
      </div>

      {/* Add hover effect for progress handle */}
      <style>{`
        .progress-handle { transform: scale(0); }
        div:hover > .progress-handle,
        div:hover .progress-handle { transform: scale(1) !important; }
      `}</style>
    </div>
  );
};

export default VideoPlayer;
