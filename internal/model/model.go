package model

// data BusinessRule = ID Name Description Schema
// data Schema = ID Name Description Type [Config]
// data Config = EnumConfig | StringConfig | IntConfig | BooleanConfig
// data StringConfig = Name Description Option Default :: String Required :: Bool
// data IntConfig = Name Description Option Default :: int Required :: Bool
// data BooleanConfig = Name Description Option Default :: Boolean Required :: Bool
// data EnumConfig = Name Description Enum :: [String] Default :: String Required :: Bool
// data Type = String
// data Name = String
// data Description = String
// data ID = String

// businessRuleOf :: ID -> Name -> Description -> Schema -> BusinessRule
// schemaOf :: ID -> Name -> Description -> Type -> [Config] -> Schema
// descriptionOf :: Predicate -> String -> Either error Description
// nameOf :: Predicate -> String -> Either error Name
// typeOf :: Predicate -> String -> Either error Type
// idOf :: () -> ID

import (
	CP "air-example/internal/typeclass/predicate"

	"fmt"

	E "github.com/IBM/fp-go/either"
	F "github.com/IBM/fp-go/function"
	O "github.com/IBM/fp-go/option"
)

type ID struct {
	Value string
}

type Name struct {
	Value string
}

type InvalidName struct {
	Value string
}

func (e InvalidName) Error() string {
	return fmt.Sprintf(`Invalid name:"%s"`, e.Value)
}

func NameOf(predicate CP.Predicate[string]) func(string) E.Either[error, Name] {
	return func(v string) E.Either[error, Name] {
		return E.Map[error](
			func(v string) Name {
				return Name{Value: v}
			},
		)(E.FromPredicate(predicate, func(v string) error {
			return InvalidName{Value: v}
		})(v))
	}
}

type Description struct {
	Value string
}

type InvalidDescription struct {
	Value string
}

func (e InvalidDescription) Error() string {
	return fmt.Sprintf(`Invalid description:"%s"`, e.Value)
}

func DescriptionOf(predicate CP.Predicate[string]) func(string) E.Either[error, Description] {
	return func(v string) E.Either[error, Description] {
		return E.Map[error](
			func(v string) Description {
				return Description{Value: v}
			},
		)(E.FromPredicate(predicate, func(v string) error {
			return InvalidDescription{Value: v}
		})(v))
	}
}

type Type struct {
	Value string
}

type InvalidType struct {
	Value string
}

func (e InvalidType) Error() string {
	return fmt.Sprintf(`Invalid type:"%s"`, e.Value)
}

func TypeOf(predicate CP.Predicate[string]) func(string) E.Either[error, Type] {
	return func(v string) E.Either[error, Type] {
		return E.Map[error](
			func(v string) Type {
				return Type{Value: v}
			},
		)(E.FromPredicate(predicate, func(v string) error {
			return InvalidType{Value: v}
		})(v))
	}
}

type Config interface {
	ConfigTag() string
}

type StringConfig struct {
	Name        Name
	Description Description
	Default     O.Option[string]
	Required    bool
}

func (c StringConfig) ConfigTag() string {
	return "StringConfig"
}

var StringConfigOf = F.Curry3(
	func(name Name, description Description, required bool) StringConfig {
		return StringConfig{
			Name:        name,
			Description: description,
			Default:     O.None[string](),
			Required:    required,
		}
	},
)

var StringConfigWithDefaultOf = F.Curry4(
	func(name Name, description Description, defaultValue string, required bool) StringConfig {
		return StringConfig{
			Name:        name,
			Description: description,
			Default:     O.Some(defaultValue),
			Required:    required,
		}
	},
)

type IntConfig struct {
	Name        Name
	Description Description
	Default     O.Option[int]
	Required    bool
}

func (c IntConfig) ConfigTag() string {
	return "IntConfig"
}

var IntConfigOf = F.Curry3(
	func(name Name, description Description, required bool) IntConfig {
		return IntConfig{
			Name:        name,
			Description: description,
			Default:     O.None[int](),
			Required:    required,
		}
	},
)

var IntConfigWithDefaultOf = F.Curry4(
	func(name Name, description Description, defaultValue int, required bool) IntConfig {
		return IntConfig{
			Name:        name,
			Description: description,
			Default:     O.Some(defaultValue),
			Required:    required,
		}
	},
)

type BooleanConfig struct {
	Name        Name
	Description Description
	Default     O.Option[bool]
	Required    bool
}

var BooleanConfigOf = F.Curry3(
	func(name Name, description Description, required bool) BooleanConfig {
		return BooleanConfig{
			Name:        name,
			Description: description,
			Default:     O.None[bool](),
			Required:    required,
		}
	},
)

var BooleanConfigWithDefaultOf = F.Curry4(
	func(name Name, description Description, defaultValue bool, required bool) BooleanConfig {
		return BooleanConfig{
			Name:        name,
			Description: description,
			Default:     O.Some(defaultValue),
			Required:    required,
		}
	},
)

func (c BooleanConfig) ConfigTag() string {
	return "BooleanConfig"
}

type EnumConfig struct {
	Name        Name
	Description Description
	Enum        []string
	Default     O.Option[string]
	Required    bool
}

func (c EnumConfig) ConfigTag() string {
	return "EnumConfig"
}

var EnumConfigOf = F.Curry4(
	func(name Name, description Description, enum []string, required bool) EnumConfig {
		return EnumConfig{
			Name:        name,
			Description: description,
			Enum:        enum,
			Default:     O.None[string](),
			Required:    required,
		}
	},
)

var EnumConfigWithDefaultOf = F.Curry5(
	func(name Name, description Description, enum []string, defaultValue string, required bool) EnumConfig {
		return EnumConfig{
			Name:        name,
			Description: description,
			Enum:        enum,
			Default:     O.Some(defaultValue),
			Required:    required,
		}
	},
)

type Schema struct {
	ID          ID
	Name        Name
	Description Description
	Type        Type
	Configs     []Config
}

var SchemaOf = F.Curry5(
	func(id ID, name Name, description Description, t Type, configs []Config) Schema {
		return Schema{
			ID:          id,
			Name:        name,
			Description: description,
			Type:        t,
			Configs:     configs,
		}
	},
)

type BusinessRule struct {
	ID          ID
	Name        Name
	Description Description
	Schema      Schema
}

var BusinessRuleOf = F.Curry4(
	func(id ID, name Name, description Description, schema Schema) BusinessRule {
		return BusinessRule{
			ID:          id,
			Name:        name,
			Description: description,
			Schema:      schema,
		}
	},
)

type IdOf = func() E.Either[error, ID]

type InvalidPrefix struct {
	Value string
}

func (e InvalidPrefix) Error() string {
	return fmt.Sprintf(`Invalid prefix:"%s"`, e.Value)
}

type InvalidSerial struct {
	Value int
}

func (e InvalidSerial) Error() string {
	return fmt.Sprintf(`Invalid serial:"%d"`, e.Value)
}

// Predicate -> string -> Predicate -> int -> (() -> Either error ID)
func PrefixIdOf(pPredicate CP.Predicate[string]) func(string) func(CP.Predicate[int]) func(int) IdOf {
	return func(prefix string) func(CP.Predicate[int]) func(int) IdOf {
		return func(p CP.Predicate[int]) func(int) IdOf {
			return func(serial int) IdOf {
				prefixE := E.FromPredicate(
					pPredicate,
					func(v string) error {
						return InvalidPrefix{Value: v}
					},
				)(prefix)
				serialE := E.FromPredicate(
					p,
					func(v int) error {
						return InvalidSerial{Value: v}
					},
				)(serial)

				return func() E.Either[error, ID] {
					return F.Pipe2(
						E.Of[error](
							func(p string) func(int) ID {
								return func(s int) ID {
									return ID{Value: fmt.Sprintf("%s-%d", p, s)}
								}
							},
						),
						E.Ap[func(int) ID](prefixE),
						E.Ap[ID](serialE),
					)
				}
			}
		}
	}
}
