# Cache, queue, storage, workers

All are plugins registered via providers and reached through the App container.

- **Cache** (`togo-framework/cache`) — `CACHE_DRIVER` = memory | file | database | redis.
- **Queue** (`togo-framework/queue`) — in-process job dispatch; `a.Queue.Dispatch(ctx, name, payload)`.
- **Storage** (`togo-framework/storage`) — filesystem blobs; `a.Storage.Put/Get/Delete`.
- **Workers** (`togo-framework/worker`) — supervised multi-threaded pools:

```go
worker.Register("emails", 4, func(ctx context.Context) error {
    // pull + process a job; restarted with backoff on error/panic
    return nil
})
```
