# Authentication & authorization

`togo-framework/auth` is the security boundary (enterprise-hardened; scanned with
govulncheck + gosec).

- **Login/JWT** — HS256, expiry-required, issuer-pinned; bcrypt passwords.
- **Sessions** — `SESSION_DRIVER` = cookie (stateless) | database | file | redis. Revocable.
- **RBAC** — roles + permissions; `RequireRole` / `RequirePermission` middleware; multi-guard.
- **MFA** — OTP (email via events), TOTP 2FA (RFC 6238), PIN lock screen.
- **API tokens** — scoped abilities (Sanctum/Cloudflare-style): `POST /api/auth/tokens`.
- **Security** — CSRF (double-submit), CORS allowlist, per-IP rate limiting, HttpOnly+SameSite cookies.
- **Drivers** — Supabase/GoTrue, OAuth (Google/GitHub/Facebook), Firebase, WorkOS SSO.
- **UI** — `togo-framework/dashboard` injects login/register/reset/2fa/pin/lock/profile/dashboard (i18n).

```go
r.With(auth.Svc.RequirePermission("posts.write")).Post("/api/posts", create)
```
