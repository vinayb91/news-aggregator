package worker

import (
	"context"
	"log"
	"time"

	"github.com/vinayb91/news-aggregator/internal/service"
)

type Worker struct {
	svc      *service.ArticleService
	interval time.Duration
	quit     chan struct{}
}

func NewWorker(svc *service.ArticleService, interval time.Duration) *Worker {
	return &Worker{svc: svc, interval: interval, quit: make(chan struct{})}
}

func (w *Worker) Start(ctx context.Context) {
	w.refresh(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("worker: context cancelled, stopping")
			return
		case <-w.quit:
			log.Println("worker: received stop signal, stopping")
			return
		case <-ticker.C:
			w.refresh(ctx)
		}
	}
}

func (w *Worker) Stop() {
	select {
	case <-w.quit:
	default:
		close(w.quit)
	}
}

func (w *Worker) refresh(parentCtx context.Context) {
	log.Println("worker: starting refresh")
	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
	defer cancel()
	_, _, err := w.svc.ListArticles(ctx, 1, 50)
	if err != nil {
		log.Printf("worker: refresh error: %v", err)
		return
	}
	log.Println("worker: refresh completed")
}
