# Search

`togo-framework/search` — `SEARCH_DRIVER` = **paradedb** (default; Postgres BM25,
ILIKE fallback so dev on SQLite works) | elasticsearch | opensearch
(`togo-framework/search-elasticsearch`).

```go
s, _ := search.FromKernel(k)
s.Index(ctx, "posts", post.ID, map[string]any{"title": post.Title})
hits, _ := s.Search(ctx, "posts", "golang", 20)
```
