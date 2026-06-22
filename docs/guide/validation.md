# Validation

Request validation (`togo-framework/validation`) uses `validate` struct tags.
Generated REST handlers run it on the request body and return **422** with field
errors before touching the database.

```go
type CreatePost struct {
    Title string `json:"title" validate:"required,min=3,max=120"`
    Email string `json:"email" validate:"required,email"`
    Role  string `json:"role"  validate:"in=admin editor viewer"`
}
```

Rules: `required, email, url, uuid, min, max, len, in, numeric` (nullable-aware).
