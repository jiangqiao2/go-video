package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Session represents an active SSE connection tracked in Redis.
type Session struct {
	UserID       string    `json:"user_id"`
	ConnectionID string    `json:"connection_id"`
	InstanceID   string    `json:"instance_id"`
	LastSeen     time.Time `json:"last_seen"`
	Metadata     string    `json:"metadata,omitempty"`
}

// Reference is a lightweight pointer to a stored session.
type Reference struct {
	UserID       string
	ConnectionID string
}

// RedisSessionStore persists SSE session metadata in Redis for multi-node coordination.
type RedisSessionStore struct {
	client redis.Cmdable
	ttl    time.Duration
}

// NewRedisSessionStore builds a session store backed by Redis.
func NewRedisSessionStore(client redis.Cmdable, ttl time.Duration) *RedisSessionStore {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &RedisSessionStore{client: client, ttl: ttl}
}

// Register records a new SSE session for the given user and instance.
func (s *RedisSessionStore) Register(ctx context.Context, sess Session) error {
	if sess.UserID == "" || sess.ConnectionID == "" {
		return fmt.Errorf("invalid session: missing identifiers")
	}
	if sess.InstanceID == "" {
		return fmt.Errorf("invalid session: missing instance id")
	}

	sess.LastSeen = time.Now().UTC()

	payload, err := json.Marshal(&sess)
	if err != nil {
		return err
	}

	userKey := userHashKey(sess.UserID)
	instanceKey := instanceSetKey(sess.InstanceID)
	ref := encodeReference(sess.UserID, sess.ConnectionID)

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, userKey, sess.ConnectionID, payload)
	pipe.Expire(ctx, userKey, s.ttl)
	pipe.SAdd(ctx, instanceKey, ref)
	pipe.Expire(ctx, instanceKey, s.ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// Heartbeat refreshes the last seen timestamp and TTL for an existing session.
func (s *RedisSessionStore) Heartbeat(ctx context.Context, userID, connectionID string) error {
	userKey := userHashKey(userID)
	raw, err := s.client.HGet(ctx, userKey, connectionID).Bytes()
	if err != nil {
		return err
	}

	var sess Session
	if err := json.Unmarshal(raw, &sess); err != nil {
		return err
	}

	sess.LastSeen = time.Now().UTC()
	payload, err := json.Marshal(&sess)
	if err != nil {
		return err
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, userKey, connectionID, payload)
	pipe.Expire(ctx, userKey, s.ttl)
	pipe.Expire(ctx, instanceSetKey(sess.InstanceID), s.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

// Remove removes a session from both the user hash and the instance set.
func (s *RedisSessionStore) Remove(ctx context.Context, sess Session) error {
	if sess.UserID == "" || sess.ConnectionID == "" {
		return fmt.Errorf("invalid session: missing identifiers")
	}

	pipe := s.client.TxPipeline()
	pipe.HDel(ctx, userHashKey(sess.UserID), sess.ConnectionID)
	if sess.InstanceID != "" {
		pipe.SRem(ctx, instanceSetKey(sess.InstanceID), encodeReference(sess.UserID, sess.ConnectionID))
	}
	_, err := pipe.Exec(ctx)
	return err
}

// ActiveSessions returns all sessions currently tracked for a user.
func (s *RedisSessionStore) ActiveSessions(ctx context.Context, userID string) ([]Session, error) {
	raw, err := s.client.HGetAll(ctx, userHashKey(userID)).Result()
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0, len(raw))
	for _, v := range raw {
		var sess Session
		if err := json.Unmarshal([]byte(v), &sess); err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// ReferencesByInstance lists all session references bound to a gateway instance.
func (s *RedisSessionStore) ReferencesByInstance(ctx context.Context, instanceID string) ([]Reference, error) {
	members, err := s.client.SMembers(ctx, instanceSetKey(instanceID)).Result()
	if err != nil {
		return nil, err
	}

	refs := make([]Reference, 0, len(members))
	for _, member := range members {
		ref, ok := decodeReference(member)
		if ok {
			refs = append(refs, ref)
		}
	}
	return refs, nil
}

func userHashKey(userID string) string {
	return fmt.Sprintf("sse:conn:%s", userID)
}

func instanceSetKey(instanceID string) string {
	return fmt.Sprintf("sse:group:%s", instanceID)
}

func encodeReference(userID, connectionID string) string {
	return userID + "|" + connectionID
}

func decodeReference(val string) (Reference, bool) {
	parts := strings.SplitN(val, "|", 2)
	if len(parts) != 2 {
		return Reference{}, false
	}
	return Reference{UserID: parts[0], ConnectionID: parts[1]}, true
}
