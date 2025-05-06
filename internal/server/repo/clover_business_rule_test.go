package repo

import (
	"os"
	"testing"

	M "air-example/internal/model"
	CP "air-example/internal/typeclass/predicate"

	A "github.com/IBM/fp-go/array"
	E "github.com/IBM/fp-go/either"
	F "github.com/IBM/fp-go/function"
	O "github.com/IBM/fp-go/option"
	"github.com/stretchr/testify/assert"
)

func TestBusinessRule(t *testing.T) {
	idE := M.PrefixIdOf(CP.NonEmptyStr)("EBX")(CP.PositiveInt)(64)()
	nameE := M.NameOf(CP.NonEmptyStr)("X-Tool-Rule")
	descriptionE := M.DescriptionOf(CP.NonEmptyStr)("Business rule for the tool X")

	sIdE := M.PrefixIdOf(CP.NonEmptyStr)("S")(CP.PositiveInt)(76312)()
	sNameE := M.NameOf(CP.NonEmptyStr)("X-Tool-Rule-Schema")
	sDescriptionE := M.DescriptionOf(CP.NonEmptyStr)("Schema for the tool X business rule")
	typeE := M.TypeOf(CP.NonEmptyStr)("Static")

	configNameE := M.NameOf(CP.NonEmptyStr)("SensorA")
	configDescriptionE := M.DescriptionOf(CP.NonEmptyStr)("Sensor A config")
	intConfigE := F.Pipe3(
		E.Of[error](M.IntConfigOf),
		E.Ap[func(M.Description) func(bool) M.IntConfig](configNameE),
		E.Ap[func(bool) M.IntConfig](configDescriptionE),
		E.Ap[M.IntConfig](E.Of[error](true)),
	)

	configsE := E.SequenceArray(A.Of(E.Either[error, M.Config](intConfigE)))
	schemaE := F.Pipe5(
		E.Of[error](M.SchemaOf),
		E.Ap[func(M.Name) func(M.Description) func(M.Type) func([]M.Config) M.Schema](sIdE),
		E.Ap[func(M.Description) func(M.Type) func([]M.Config) M.Schema](sNameE),
		E.Ap[func(M.Type) func([]M.Config) M.Schema](sDescriptionE),
		E.Ap[func([]M.Config) M.Schema](typeE),
		E.Ap[M.Schema](configsE),
	)

	businessRuleE := F.Pipe4(
		E.Of[error](M.BusinessRuleOf),
		E.Ap[func(M.Name) func(M.Description) func(M.Schema) M.BusinessRule](idE),
		E.Ap[func(M.Description) func(M.Schema) M.BusinessRule](nameE),
		E.Ap[func(M.Schema) M.BusinessRule](descriptionE),
		E.Ap[M.BusinessRule](schemaE),
	)

	defer os.RemoveAll("test.db")

	repo := CloverBusinessRuleRepo{
		dbPath:     "test.db",
		collection: "business_rule",
	}

	rule, err := E.Unwrap(
		E.Chain(
			func(rule M.BusinessRule) E.Either[error, M.BusinessRule] {
				return repo.SaveBusinessRule(rule)
			})(businessRuleE),
	)

	assert.Nil(t, err)
	assert.Equal(t, rule.Name.Value, "X-Tool-Rule")
	assert.Equal(t, 1, len(rule.Schema.Configs))

	_, err2 := E.Unwrap(
		E.Chain(
			func(rule M.BusinessRule) E.Either[error, M.BusinessRule] {
				return repo.SaveBusinessRule(rule)
			})(businessRuleE),
	)

	assert.NotNil(t, err2)
	assert.ErrorAs(t, err2, &BusinessRuleAlreadyExists{"EBX-64"})

	rule2, err3 := E.Unwrap(repo.FindBusinessRuleByID(M.ID{Value: "EBX-64"}))

	assert.Nil(t, err3)
	assert.Equal(t, rule2.Name.Value, "X-Tool-Rule")
	assert.Equal(t, 1, len(rule2.Schema.Configs))
}

func TestConfig2BytesTransformation(t *testing.T) {
	configs := A.From(
		M.Config(M.IntConfig{
			Name:        M.Name{Value: "SensorA"},
			Description: M.Description{Value: "Sensor A config"},
			Default:     O.Some(3),
			Required:    false,
		}),
		M.Config(M.BooleanConfig{
			Name:        M.Name{Value: "SensorB"},
			Description: M.Description{Value: "Sensor B config"},
			Default:     O.None[bool](),
			Required:    true,
		}),
	)

	r, err := E.Unwrap(Configs2Bytes(configs))

	assert.Nil(t, err)
	assert.Equal(t, len(r), 2)
	assert.Equal(t, r["int-SensorA"].([]byte), []byte(`{"Name":{"Value":"SensorA"},"Description":{"Value":"Sensor A config"},"Default":3,"Required":false}`))
	assert.Equal(t, r["bool-SensorB"].([]byte), []byte(`{"Name":{"Value":"SensorB"},"Description":{"Value":"Sensor B config"},"Default":null,"Required":true}`))

	r2, err2 := E.Unwrap(Bytes2Configs(r))

	assert.Nil(t, err2)
	assert.Equal(t, len(r2), 2)

	assert.Equal(t, r2[0].(M.IntConfig).Name.Value, "SensorA")
	assert.Equal(t, r2[0].(M.IntConfig).Description.Value, "Sensor A config")
	assert.Equal(t, r2[0].(M.IntConfig).Default, O.Some(3))
	assert.Equal(t, r2[0].(M.IntConfig).Required, false)
	assert.Equal(t, r2[1].(M.BooleanConfig).Name.Value, "SensorB")
	assert.Equal(t, r2[1].(M.BooleanConfig).Description.Value, "Sensor B config")
	assert.Equal(t, r2[1].(M.BooleanConfig).Default, O.None[bool]())
	assert.Equal(t, r2[1].(M.BooleanConfig).Required, true)
}
