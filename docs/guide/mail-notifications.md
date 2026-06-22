# Mail & notifications

**Mail** (`togo-framework/mail`): a `Mailer` with SMTP (default) + log drivers;
Resend via `togo-framework/mail-resend` (`MAIL_DRIVER=resend`). SES works through SMTP.

**Notifications** (`togo-framework/notifications`): Laravel-style channels —
`mail`, `broadcast` (realtime), `database`, and push providers (OneSignal via
`togo-framework/notifications-onesignal`; FCM/Pusher follow the same pattern).

```go
type OrderShipped struct{ ID string }
func (OrderShipped) Via(n notifications.Notifiable) []string { return []string{"mail", "database"} }
func (o OrderShipped) ToMail(n notifications.Notifiable) mail.Message { ... }
```
