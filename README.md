# News Aggregator (Go)

A concurrent content aggregator in Go that fetches Hacker News stories, caches them in Redis, and serves a paginated REST API.

## Run locally

```bash
docker compose up --build
```

Visit http://localhost:8080/v1/articles
