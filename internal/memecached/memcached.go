package memecached

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	client    *memcache.Client
	ttl       time.Duration
	keyPrefix string
)

func Client() *memcache.Client {
	return client
}

func Init(servers []string, t time.Duration, prefix string) {
	client = memcache.New(servers...)
	ttl = t
	keyPrefix = prefix
}

func DefaultTTL() time.Duration {
	return ttl
}

func GetKeyPrefix() string {
	return keyPrefix
}
