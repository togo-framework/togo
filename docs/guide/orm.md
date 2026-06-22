# ORM

A driver-agnostic, Eloquent-style query builder (`togo-framework/orm`). The
dialect (placeholders, ILIKE) is chosen from `DB_DRIVER` at runtime, so the same
code runs on SQLite, Postgres, or MySQL.

```go
posts, err := models.Posts(app).
    Where("title", "ILIKE", "%go%").
    Order("created_at DESC").
    Limit(20).
    Get(ctx)

one, err := models.Posts(app).Find(ctx, id)
created, err := models.Posts(app).Create(ctx, map[string]any{"title": "Hi"})
err = models.Posts(app).Where("id", "=", id).Update(ctx, map[string]any{"title": "Edited"})
err = models.Posts(app).Where("id", "=", id).Delete(ctx)
```

Column, operator, and ORDER BY inputs are validated against an allowlist; values
are always parameterized (SQL-injection safe). sqlc is kept for typed model structs.
