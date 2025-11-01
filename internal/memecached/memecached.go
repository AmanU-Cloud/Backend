package memecached

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	client *memcache.Client
	ttl    time.Duration
)

func Client() *memcache.Client {
	return client
}

func Init(addr string, t time.Duration) {
	client = memcache.New(addr)
	ttl = t
}

func DefaultTTL() time.Duration {
	return ttl
}
