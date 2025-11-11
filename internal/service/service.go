package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/vinayb91/news-aggregator/internal/cache"
	"github.com/vinayb91/news-aggregator/internal/fetcher"
)

type Article = fetcher.Article

type Fetcher interface {
	Fetch(ctx context.Context) ([]Article, error)
}

// ADD THIS INTERFACE
type CacheInterface interface {
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

type ArticleService struct {
	cache   CacheInterface // CHANGE THIS LINE - was *cache.RedisCache
	fetcher Fetcher
}

// CHANGE THIS FUNCTION SIGNATURE
func NewArticleService(c CacheInterface, f Fetcher) *ArticleService {
	return &ArticleService{cache: c, fetcher: f}
}

// ADD THIS HELPER (optional, for existing code compatibility)
func NewArticleServiceWithRedis(c *cache.RedisCache, f Fetcher) *ArticleService {
	return &ArticleService{cache: c, fetcher: f}
}

// Rest of your ListArticles function stays the same
func (s *ArticleService) ListArticles(ctx context.Context, page, perPage int) ([]Article, int, error) {
	key := fmt.Sprintf("articles:page:%d:%d", page, perPage)
	var arts []Article
	ok, err := s.cache.Get(ctx, key, &arts)
	if err != nil {
		return nil, 0, err
	}
	if ok {
		return arts, len(arts), nil
	}

	all, err := s.fetcher.Fetch(ctx)
	if err != nil {
		return nil, 0, err
	}

	dedup := map[string]Article{}
	for _, a := range all {
		k := a.URL
		if k == "" {
			h := sha1.Sum([]byte(a.Title + a.Source))
			k = hex.EncodeToString(h[:])
		}
		if _, exists := dedup[k]; !exists {
			dedup[k] = a
		}
	}
	var unique []Article
	for _, v := range dedup {
		unique = append(unique, v)
	}
	total := len(unique)
	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	pageItems := unique[start:end]

	_ = s.cache.Set(ctx, key, pageItems, 60*time.Second)
	return pageItems, total, nil
}
