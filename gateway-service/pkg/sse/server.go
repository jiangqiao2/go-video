package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"gateway-service/pkg/auth"
	"gateway-service/pkg/sse/bus"
	"gateway-service/pkg/sse/store"
)

const (
	defaultSessionTTL     = 10 * time.Minute
	defaultHeartbeat      = 45 * time.Second
	defaultStreamBlock    = 25 * time.Second
	defaultStreamBatch    = int64(20)
	streamKeyTemplate     = "sse:event:%s"
	topicTemplate         = "user:%s"
	lastEventIDQueryKey   = "last_event_id"
	accessTokenQueryKey   = "access_token"
	authorizationHeader   = "Authorization"
	lastEventIDHeader     = "Last-Event-ID"
	defaultKeepAliveFrame = ": ping\n\n"
)

// Server manages SSE client connections and forwards Redis stream events.
type Server struct {
	client    redis.Cmdable
	store     *store.RedisSessionStore
	logger    *logrus.Entry
	instance  string
	heartbeat time.Duration
	block     time.Duration
	count     int64
}

// NewServer builds an SSE server bound to a Redis backend.
func NewServer(client redis.Cmdable, logger *logrus.Logger) *Server {
	if client == nil {
		return nil
	}
	if logger == nil {
		logger = logrus.New()
	}
	instanceID := strings.ReplaceAll(uuid.NewString(), "-", "")

	return &Server{
		client:    client,
		store:     store.NewRedisSessionStore(client, defaultSessionTTL),
		logger:    logrus.NewEntry(logger).WithField("component", "sse"),
		instance:  instanceID,
		heartbeat: defaultHeartbeat,
		block:     defaultStreamBlock,
		count:     defaultStreamBatch,
	}
}

// Handler upgrades the HTTP connection to an SSE stream after validating the access token.
func (s *Server) Handler(jwt *auth.JWTUtil) gin.HandlerFunc {
	return func(c *gin.Context) {
		if s == nil || s.client == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "SSE 服务未就绪",
			})
			return
		}
		token := extractAccessToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "缺少访问令牌",
			})
			return
		}

		userUUID, userID, err := jwt.ValidateAccessTokenWithUUID(token)
		if err != nil || userUUID == "" {
			s.logger.WithError(err).Warn("invalid sse token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "访问令牌无效",
			})
			return
		}

		connectionID := uuid.NewString()
		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		session := store.Session{
			UserID:       userUUID,
			ConnectionID: connectionID,
			InstanceID:   s.instance,
		}
		if err := s.store.Register(ctx, session); err != nil {
			s.logger.WithError(err).Error("register sse session failed")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "注册 SSE 会话失败",
			})
			return
		}
		defer func() {
			// 使用后台上下文确保注销成功
			if err := s.store.Remove(context.Background(), session); err != nil {
				s.logger.WithError(err).Warn("cleanup sse session failed")
			}
		}()

		setupStreamHeaders(c.Writer)
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "服务器不支持流式响应",
			})
			return
		}

		lastID := strings.TrimSpace(c.Query(lastEventIDQueryKey))
		if lastID == "" {
			lastID = strings.TrimSpace(c.GetHeader(lastEventIDHeader))
		}
		if lastID == "" {
			lastID = "$"
		}

		s.logger.WithFields(logrus.Fields{
			"user_uuid":     userUUID,
			"user_id":       userID,
			"connection_id": connectionID,
			"last_id":       lastID,
		}).Info("SSE connection established")

		events := make(chan bus.Message, 32)
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			s.consume(ctx, userUUID, lastID, events)
		}()

		go func() {
			defer wg.Done()
			s.keepAlive(ctx, userUUID, connectionID)
		}()

		defer func() {
			cancel()
			wg.Wait()
		}()

		keepAliveTicker := time.NewTicker(s.block)
		defer keepAliveTicker.Stop()

		// 立即发送一次 ping 保持连接
		if _, err := io.WriteString(c.Writer, defaultKeepAliveFrame); err == nil {
			flusher.Flush()
		}

		c.Stream(func(w io.Writer) bool {
			select {
			case msg, ok := <-events:
				if !ok {
					return false
				}
				if len(msg.Payload) == 0 {
					msg.Payload = json.RawMessage(`{}`)
				}
				c.SSEvent(msg.Event, msg.Payload)
				flusher.Flush()
				return true
			case <-keepAliveTicker.C:
				if _, err := w.Write([]byte(defaultKeepAliveFrame)); err == nil {
					flusher.Flush()
					return true
				}
				return false
			case <-ctx.Done():
				return false
			}
		})

		s.logger.WithFields(logrus.Fields{
			"user_uuid":     userUUID,
			"connection_id": connectionID,
		}).Info("SSE connection closed")
	}
}

func (s *Server) consume(ctx context.Context, userUUID, lastID string, out chan<- bus.Message) {
	defer close(out)

	streamKey := fmt.Sprintf(streamKeyTemplate, fmt.Sprintf(topicTemplate, userUUID))
	cursor := lastID
	if cursor == "" {
		cursor = "$"
	}

	for {
		if ctx.Err() != nil {
			return
		}

		res, err := s.client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamKey, cursor},
			Block:   s.block,
			Count:   s.count,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			s.logger.WithError(err).Warn("读取 SSE 消息失败")
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, stream := range res {
			for _, raw := range stream.Messages {
				body, ok := raw.Values["body"].(string)
				if !ok {
					continue
				}
				var message bus.Message
				if err := json.Unmarshal([]byte(body), &message); err != nil {
					s.logger.WithError(err).Warn("解码 SSE 消息失败")
					continue
				}
				message.ID = raw.ID
				cursor = raw.ID

				select {
				case out <- message:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (s *Server) keepAlive(ctx context.Context, userUUID, connectionID string) {
	ticker := time.NewTicker(s.heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.store.Heartbeat(context.Background(), userUUID, connectionID); err != nil {
				s.logger.WithError(err).Warn("刷新 SSE 心跳失败")
			}
		case <-ctx.Done():
			return
		}
	}
}

func extractAccessToken(c *gin.Context) string {
	if token := strings.TrimSpace(c.Query(accessTokenQueryKey)); token != "" {
		return token
	}
	authHeader := strings.TrimSpace(c.GetHeader(authorizationHeader))
	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return ""
	}
	return strings.TrimSpace(authHeader[7:])
}

func setupStreamHeaders(w http.ResponseWriter) {
	headers := w.Header()
	headers.Set("Content-Type", "text/event-stream; charset=utf-8")
	headers.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")
}
