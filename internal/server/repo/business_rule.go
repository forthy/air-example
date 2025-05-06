package repo

import (
	M "air-example/internal/model"

	"fmt"

	E "github.com/IBM/fp-go/either"
)

type BusinessRuleRepo interface {
	// BusinessRule -> Either error BusinessRule
	SaveBusinessRule(b M.BusinessRule) E.Either[error, M.BusinessRule]

	// ID -> Either error BusinessRule
	FindBusinessRuleByID(id M.ID) E.Either[error, M.BusinessRule]
}

type BusinessRuleAlreadyExists struct {
	ID string
}

func (e BusinessRuleAlreadyExists) Error() string {
	return fmt.Sprintf("Business rule ID: [%s] already exists: ", e.ID)
}

type BusinessRuleNotFound struct {
	ID string
}

func (e BusinessRuleNotFound) Error() string {
	return fmt.Sprintf("Business rule ID: [%s] not found: ", e.ID)
}
