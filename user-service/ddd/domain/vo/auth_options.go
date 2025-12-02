package vo

import "time"

type AuthOptions struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}
