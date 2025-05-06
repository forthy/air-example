package repo

import (
	M "air-example/internal/model"
	"fmt"
	"os"
	"strings"

	c "github.com/ostafen/clover/v2"
	D "github.com/ostafen/clover/v2/document"
	Q "github.com/ostafen/clover/v2/query"

	A "github.com/IBM/fp-go/array"
	E "github.com/IBM/fp-go/either"
	F "github.com/IBM/fp-go/function"
	I "github.com/IBM/fp-go/io"
	J "github.com/IBM/fp-go/json"
	R "github.com/IBM/fp-go/record"
	T "github.com/IBM/fp-go/tuple"
)

// DBPath :: string -> Collection :: string -> (DB -> Either error a) -> Either error a
func withDB[T any](dbPath string) func(string) func(func(*c.DB) E.Either[error, T]) E.Either[error, T] {
	return func(collection string) func(func(*c.DB) E.Either[error, T]) E.Either[error, T] {
		return E.WithResource[error, *c.DB, T](
			func() E.Either[error, *c.DB] {
				if _, err := os.Stat(dbPath); os.IsNotExist(err) {
					if err := os.Mkdir(dbPath, 0755); err != nil {
						return E.Left[*c.DB](err)
					}
				}

				return F.Pipe1(
					E.Eitherize1(c.Open)(dbPath),
					E.Chain(func(db *c.DB) E.Either[error, *c.DB] {
						return F.Pipe1(
							E.Eitherize1(db.HasCollection)(collection),
							E.Chain(func(exist bool) E.Either[error, *c.DB] {
								if exist {
									return E.Right[error](db)
								}
								return E.Map[error](
									func(_ string) *c.DB {
										return db
									},
								)(E.FromError(db.CreateCollection)(collection))
							}),
						)
					}),
				)
			},
			func(db *c.DB) E.Either[error, any] {
				if err := db.Close(); err != nil {
					return E.Left[any](err)
				} else {
					return E.Right[error](F.ToAny("DB is closed"))
				}
			},
		)
	}
}

type CloverBusinessRuleRepo struct {
	dbPath     string
	collection string
}

func (repo CloverBusinessRuleRepo) SaveBusinessRule(b M.BusinessRule) E.Either[error, M.BusinessRule] {
	return withDB[M.BusinessRule](repo.dbPath)(repo.collection)(
		func(db *c.DB) E.Either[error, M.BusinessRule] {
			return F.Pipe1(
				E.Eitherize1(db.FindFirst)(Q.NewQuery(repo.collection).Where(Q.Field("id").Eq(b.ID.Value))),
				E.Chain(func(d *D.Document) E.Either[error, M.BusinessRule] {
					if d == nil {
						rule := b
						configs := b.Schema.Configs

						rule.Schema.Configs = A.Empty[M.Config]()
						return F.Pipe1(
							J.Marshal(rule.Schema),
							E.Chain(func(v []byte) E.Either[error, M.BusinessRule] {
								bDoc := D.NewDocument()
								bDoc.Set("id", rule.ID.Value)
								bDoc.Set("name", rule.Name.Value)
								bDoc.Set("description", rule.Description.Value)
								bDoc.Set("created_at", I.Now())
								bDoc.Set("schema", string(v))

								return F.Pipe1(
									E.Map[error](func(m map[string]any) *D.Document {
										bDoc.Set("configs", m)

										return bDoc
									})(Configs2Bytes(configs)),
									E.Chain(func(doc *D.Document) E.Either[error, M.BusinessRule] {
										return E.Map[error](func(_ string) M.BusinessRule {
											return b
										})(E.Eitherize2(db.InsertOne)(repo.collection, bDoc))
									}),
								)
							}),
						)
					}

					return E.Left[M.BusinessRule](error(BusinessRuleAlreadyExists{ID: b.ID.Value}))
				}),
			)
		},
	)
}

func (repo CloverBusinessRuleRepo) FindBusinessRuleByID(id M.ID) E.Either[error, M.BusinessRule] {
	return withDB[M.BusinessRule](repo.dbPath)(repo.collection)(
		func(db *c.DB) E.Either[error, M.BusinessRule] {
			return F.Pipe1(
				E.Eitherize1(db.FindFirst)(Q.NewQuery(repo.collection).Where(Q.Field("id").Eq(id.Value))),
				E.Chain(func(d *D.Document) E.Either[error, M.BusinessRule] {
					if d != nil {
						if schema, err := E.Unwrap(J.Unmarshal[M.Schema]([]byte(d.Get("schema").(string)))); err != nil {
							return E.Left[M.BusinessRule](err)
						} else {
							return E.Map[error](func(configs []M.Config) M.BusinessRule {
								schema.Configs = configs

								return M.BusinessRule{
									ID:          M.ID{Value: d.Get("id").(string)},
									Name:        M.Name{Value: d.Get("name").(string)},
									Description: M.Description{Value: d.Get("description").(string)},
									Schema:      schema,
								}
							})(Bytes2Configs(d.Get("configs").(map[string]any)))
						}
					}

					return E.Left[M.BusinessRule](error(BusinessRuleNotFound{ID: id.Value}))
				}),
			)
		},
	)
}

func Configs2Bytes(c []M.Config) E.Either[error, map[string]any] {
	// make a tuple2 from []Config
	// applying pattern matching in handling Config as a sum type
	// make the first element the "type-<config name>" as a key in the later record
	// JSON marshalling the second element
	// turn []Tuple2 to a record
	// JSON marshalling the record
	return E.Map[error](func(x []T.Tuple2[string, any]) map[string]any {
		return R.FromEntries(x)
	})(E.SequenceArray(A.Map(func(config M.Config) E.Either[error, T.Tuple2[string, any]] {
		switch cg := config.(type) {
		case M.IntConfig:
			return E.Map[error](func(b []byte) T.Tuple2[string, any] {
				return T.MakeTuple2(fmt.Sprintf("int-%s", cg.Name.Value), any(b))
			})(J.Marshal(cg))
		case M.BooleanConfig:
			return E.Map[error](func(b []byte) T.Tuple2[string, any] {
				return T.MakeTuple2(fmt.Sprintf("bool-%s", cg.Name.Value), any(b))
			})(J.Marshal(cg))
		case M.StringConfig:
			return E.Map[error](func(b []byte) T.Tuple2[string, any] {
				return T.MakeTuple2(fmt.Sprintf("string-%s", cg.Name.Value), any(b))
			})(J.Marshal(cg))
		case M.EnumConfig:
			return E.Map[error](func(b []byte) T.Tuple2[string, any] {
				return T.MakeTuple2(fmt.Sprintf("enum-%s", cg.Name.Value), any(b))
			})(J.Marshal(cg))
		default:
			return E.Left[T.Tuple2[string, any]](fmt.Errorf("unknown config type: %T", cg))
		}
	})(c)))
}

func Bytes2Configs(b map[string]any) E.Either[error, []M.Config] {
	// make a tuple2 from []Config
	// applying pattern matching in handling Config as a sum type
	// make the first element the "type-<config name>" as a key in the later record
	// JSON marshalling the second element
	// turn []Tuple2 to a record
	// JSON marshalling the record
	return E.Chain(func(x map[string]any) E.Either[error, []M.Config] {
		return E.SequenceArray(A.Map(func(k string) E.Either[error, M.Config] {
			switch prefix := strings.Split(k, "-")[0]; prefix {
			case "int":
				return E.Either[error, M.Config](J.Unmarshal[M.IntConfig](x[k].([]byte)))
			case "bool":
				return E.Either[error, M.Config](J.Unmarshal[M.BooleanConfig](x[k].([]byte)))
			case "string":
				return E.Either[error, M.Config](J.Unmarshal[M.StringConfig](x[k].([]byte)))
			case "enum":
				return E.Either[error, M.Config](J.Unmarshal[M.EnumConfig](x[k].([]byte)))
			default:
				return E.Left[M.Config](fmt.Errorf("unknown config type: %s", prefix))
			}
		})(R.Keys(x)))
	})(E.Right[error](b))
}
