# Testing

- **`togo test`** runs Go tests.
- **`togo make:test <Model>`** generates a feature test (PHPUnit-style) that boots
  the real stack via `internal/server.Boot()` and hits the API with the
  `togo-framework/testing` harness (`Do`, `Status`, `JSON`, `Contains`).
- **`togo make:e2e <Model>`** generates a Playwright spec (Dusk-style) under `web/e2e/`.

```go
togotest.Do(t, router, "GET", "/api/posts", nil).Status(t, 200)
```
