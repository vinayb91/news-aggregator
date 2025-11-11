package fetcher

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sort"
    "time"
)

type Article struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Summary     string    `json:"summary,omitempty"`
    URL         string    `json:"url,omitempty"`
    Source      string    `json:"source"`
    PublishedAt time.Time `json:"published_at"`
}

type HackerNewsFetcher struct{}

func NewHackerNewsFetcher() *HackerNewsFetcher { return &HackerNewsFetcher{} }

func (h *HackerNewsFetcher) Fetch(ctx context.Context) ([]Article, error) {
    client := &http.Client{Timeout: 10 * time.Second}
    idsResp, err := client.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
    if err != nil { return nil, err }
    defer idsResp.Body.Close()

    var ids []int
    if err := json.NewDecoder(idsResp.Body).Decode(&ids); err != nil { return nil, err }
    if len(ids) > 30 { ids = ids[:30] }

    ch := make(chan Article, len(ids))
    for _, id := range ids {
        go func(id int) {
            url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
            resp, err := client.Get(url)
            if err != nil { return }
            defer resp.Body.Close()
            var it struct {
                ID int `json:"id"`; Title string `json:"title"`; Time int64 `json:"time"`; URL string `json:"url"`; Text string `json:"text"`
            }
            if err := json.NewDecoder(resp.Body).Decode(&it); err != nil { return }
            ch <- Article{
                ID: fmt.Sprintf("hn-%d", it.ID),
                Title: it.Title,
                Summary: it.Text,
                URL: it.URL,
                Source: "hackernews",
                PublishedAt: time.Unix(it.Time, 0),
            }
        }(id)
    }
    var arts []Article
    for i := 0; i < cap(ch); i++ { arts = append(arts, <-ch) }
    sort.Slice(arts, func(i, j int) bool { return arts[i].PublishedAt.After(arts[j].PublishedAt) })
    return arts, nil
}