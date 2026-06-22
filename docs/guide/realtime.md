# Realtime & events

`App.Emit(ctx, event, payload)` fires kernel **hooks** (in-process listeners /
actions) and **broadcasts** over realtime to frontend subscribers.

- **SSE** (default, `togo-framework/realtime`) — `/events`; `useEvents()` hook on the frontend.
- **WebSocket** (`togo-framework/realtime-ws`) — public/private **channels**, signed
  **tickets** for private channels, channel-scoped broadcasting. Emit `"channel:event"`.

```go
a.Emit(ctx, "orders:created", order)   // → subscribers of channel "orders"
a.Emit(ctx, "post.created", post)      // → everyone (no channel)
```
