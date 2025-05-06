package predicate

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New(validator.WithRequiredStructEnabled())

	NonEmptyStr Predicate[string] = func(v string) bool {
		return strings.TrimSpace(v) != ""
	}
)
