package model

import (
	"testing"

	CP "air-example/internal/typeclass/predicate"

	E "github.com/IBM/fp-go/either"
	P "github.com/IBM/fp-go/predicate"
)

func TestPrefixID(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		prefix   string
		serial   int
		expected E.Either[error, ID]
	}{
		{"Empty prefix, valid serial", "", 24, E.Left[ID, error](InvalidPrefix{""})},
		{"Non-empty prefix, invalid serial", "EBX", -1, E.Left[ID, error](InvalidSerial{-1})},
		{"Empty prefix, invalid serial", "", -75, E.Left[ID, error](InvalidPrefix{""})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PrefixIdOf(CP.NonEmptyStr)(test.prefix)(P.Or(CP.PositiveInt)(CP.ZeroInt))(test.serial)()
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}
