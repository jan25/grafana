package tokenprovider

import (
	"context"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type AccessToken struct {
	Token     string
	ExpiresOn time.Time
}

type TokenCredential interface {
	GetCacheKey() string
	Init() error
	GetAccessToken(ctx context.Context, scopes []string) (*AccessToken, error)
}

type ConcurrentTokenCache interface {
	GetAccessToken(ctx context.Context, credential TokenCredential, scopes []string) (string, error)
}

func NewConcurrentTokenCache() ConcurrentTokenCache {
	return &tokenCacheImpl{}
}

type tokenCacheImpl struct {
	cache sync.Map // of *credentialCacheEntry
}
type credentialCacheEntry struct {
	credential TokenCredential

	credInit  uint32
	credMutex sync.Mutex
	cache     sync.Map // of *scopesCacheEntry
}

type scopesCacheEntry struct {
	credential TokenCredential
	scopes     []string

	cond        *sync.Cond
	refreshing  bool
	accessToken *AccessToken
}

func (c *tokenCacheImpl) GetAccessToken(ctx context.Context, credential TokenCredential, scopes []string) (string, error) {
	return c.getEntryFor(credential).getAccessToken(ctx, scopes)
}

func (c *tokenCacheImpl) getEntryFor(credential TokenCredential) *credentialCacheEntry {
	var entry interface{}
	var ok bool

	key := credential.GetCacheKey()

	if entry, ok = c.cache.Load(key); !ok {
		entry, _ = c.cache.LoadOrStore(key, &credentialCacheEntry{
			credential: credential,
		})
	}

	return entry.(*credentialCacheEntry)
}

func (c *credentialCacheEntry) getAccessToken(ctx context.Context, scopes []string) (string, error) {
	err := c.ensureInitialized()
	if err != nil {
		return "", err
	}

	return c.getEntryFor(scopes).getAccessToken(ctx)
}

func (c *credentialCacheEntry) ensureInitialized() error {
	if atomic.LoadUint32(&c.credInit) == 0 {
		c.credMutex.Lock()
		defer c.credMutex.Unlock()

		if c.credInit == 0 {
			// Initialize credential
			err := c.credential.Init()
			if err != nil {
				return err
			}

			atomic.StoreUint32(&c.credInit, 1)
		}
	}

	return nil
}

func (c *credentialCacheEntry) getEntryFor(scopes []string) *scopesCacheEntry {
	var entry interface{}
	var ok bool

	key := getKeyForScopes(scopes)

	if entry, ok = c.cache.Load(key); !ok {
		entry, _ = c.cache.LoadOrStore(key, &scopesCacheEntry{
			credential: c.credential,
			scopes:     scopes,
			cond:       sync.NewCond(&sync.Mutex{}),
		})
	}

	return entry.(*scopesCacheEntry)
}

func (c *scopesCacheEntry) getAccessToken(ctx context.Context) (string, error) {
	var accessToken *AccessToken
	var err error
	shouldRefresh := false

	c.cond.L.Lock()
	for {
		if c.accessToken != nil && c.accessToken.ExpiresOn.After(time.Now().Add(2*time.Minute)) {
			// Use the cached token since it's available and not expired yet
			accessToken = c.accessToken
			break
		}

		if !c.refreshing {
			// Start refreshing the token
			c.refreshing = true
			shouldRefresh = true
			break
		}

		// Wait for the token to be refreshed
		c.cond.Wait()
	}
	c.cond.L.Unlock()

	if shouldRefresh {
		accessToken, err = c.refreshAccessToken(ctx)
		if err != nil {
			return "", err
		}
	}

	return accessToken.Token, nil
}

func (c *scopesCacheEntry) refreshAccessToken(ctx context.Context) (*AccessToken, error) {
	var accessToken *AccessToken

	// Safeguarding from panic caused by credential implementation
	defer func() {
		c.cond.L.Lock()

		c.refreshing = false

		if accessToken != nil {
			c.accessToken = accessToken
		}

		c.cond.Broadcast()
		c.cond.L.Unlock()
	}()

	token, err := c.credential.GetAccessToken(ctx, c.scopes)
	if err != nil {
		return nil, err
	}
	accessToken = token
	return accessToken, nil
}

func getKeyForScopes(scopes []string) string {
	if len(scopes) > 1 {
		arr := make([]string, len(scopes))
		copy(arr, scopes)
		sort.Strings(arr)
		scopes = arr
	}

	return strings.Join(scopes, " ")
}
