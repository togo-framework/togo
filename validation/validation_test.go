package validation

import "testing"

func TestValidate(t *testing.T) {
	type In struct {
		Title string  `json:"title" validate:"required,min=3,max=10"`
		Email string  `json:"email" validate:"required,email"`
		Role  string  `json:"role" validate:"in=admin user"`
		Bio   *string `json:"bio" validate:"max=5"`
	}
	long := "toolongbio"
	errs := Validate(In{Title: "ab", Email: "nope", Role: "x", Bio: &long})
	for _, f := range []string{"title", "email", "role", "bio"} {
		if len(errs[f]) == 0 {
			t.Fatalf("expected error for %s; got %+v", f, errs)
		}
	}
	if ok := Validate(In{Title: "hello", Email: "a@b.co", Role: "admin"}); ok.Any() {
		t.Fatalf("valid input flagged: %+v", ok)
	}
}
