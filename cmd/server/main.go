package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/vinayb91/news-aggregator/internal/cache"
	"github.com/vinayb91/news-aggregator/internal/fetcher"
	"github.com/vinayb91/news-aggregator/internal/handlers"
	"github.com/vinayb91/news-aggregator/internal/service"
	"github.com/vinayb91/news-aggregator/internal/worker"
)

func main() {
	_ = godotenv.Load()

	redisAddr := getenv("REDIS_ADDR", "localhost:6379")
	listenAddr := getenv("LISTEN_ADDR", ":8080")

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	cacheClient := cache.NewRedisCache(rdb)

	hn := fetcher.NewHackerNewsFetcher()
	svc := service.NewArticleService(cacheClient, hn)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // your frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))
	handlers.RegisterRoutes(r, svc)

	w := worker.NewWorker(svc, 5*time.Minute)
	go w.Start(context.Background())

	srv := &http.Server{Addr: listenAddr, Handler: r}

	go func() {
		log.Printf("starting server at %s", listenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
