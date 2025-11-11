package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vinayb91/news-aggregator/internal/service"
)

type ArticleLister interface {
	ListArticles(ctx context.Context, page, perPage int) ([]service.Article, int, error)
}

func RegisterRoutes(r *chi.Mux, svc ArticleLister) {
	r.Get("/v1/articles", listArticlesHandler(svc))
}

func listArticlesHandler(svc ArticleLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 {
			perPage = 20
		}

		arts, total, err := svc.ListArticles(context.Background(), page, perPage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := map[string]any{
			"page":     page,
			"per_page": perPage,
			"total":    total,
			"articles": arts,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
