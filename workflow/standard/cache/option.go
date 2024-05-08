package cache

import "time"

type ErrInvalidTTL struct {
	error
}

func (e ErrInvalidTTL) Error() string {
	return "invalid TTL in cache"
}

type ErrInvalidExpiration struct{}

func (e ErrInvalidExpiration) Error() string {
	return "invalid expiration in cache"
}

type ItemOption func(*Item) error

func WithTTL(duration time.Duration) ItemOption {
	return func(i *Item) error {
		if duration <= 0 {
			return ErrInvalidTTL{}
		}
		expiredAt := time.Now().Add(duration)
		i.expiredAt = &expiredAt
		return nil
	}
}

func WithExpiredAt(expiredAt time.Time) ItemOption {
	return func(i *Item) error {
		if expiredAt.Before(time.Now()) {
			return ErrInvalidExpiration{}
		}
		i.expiredAt = &expiredAt
		return nil
	}
}
