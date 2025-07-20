package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	TC "github.com/testcontainers/testcontainers-go"
	TCM "github.com/testcontainers/testcontainers-go/modules/mariadb"

	A "github.com/IBM/fp-go/array"
	E "github.com/IBM/fp-go/either"
	F "github.com/IBM/fp-go/function"
	I "github.com/IBM/fp-go/iterator/stateless"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func variadicRun(ctx context.Context, image string, options []TC.ContainerCustomizer) (*TCM.MariaDBContainer, error) {
	return TCM.Run(ctx, image, options...)
}

func variadicTerminate(container *TCM.MariaDBContainer) error {
	return TC.TerminateContainer(*container)
}

func withContainer[T any](ctx context.Context) func(string) func([]TC.ContainerCustomizer) func(func(*TCM.MariaDBContainer) E.Either[error, T]) E.Either[error, T] {
	return func(image string) func([]TC.ContainerCustomizer) func(func(*TCM.MariaDBContainer) E.Either[error, T]) E.Either[error, T] {
		return func(options []TC.ContainerCustomizer) func(func(*TCM.MariaDBContainer) E.Either[error, T]) E.Either[error, T] {
			return E.WithResource[error, *TCM.MariaDBContainer, T](
				func() E.Either[error, *TCM.MariaDBContainer] {
					return E.Eitherize3(variadicRun)(ctx, image, options)
				},
				func(container *TCM.MariaDBContainer) E.Either[error, any] {
					return E.Either[error, any](E.FromError(variadicTerminate)(container))
				},
			)
		}
	}
}

/*
connString :: string -> DriverName :: string -> (*sql.DB -> Either error T) -> Either error T
*/
func withMariaDB[T any](connString string) func(string) func(func(*sql.DB) E.Either[error, T]) E.Either[error, T] {
	return func(driverName string) func(func(*sql.DB) E.Either[error, T]) E.Either[error, T] {
		return E.WithResource[error, *sql.DB, T](
			func() E.Either[error, *sql.DB] {
				return E.Eitherize2(sql.Open)(driverName, connString)
			},
			func(db *sql.DB) E.Either[error, any] {
				return E.Either[error, any](E.FromError(func(msg string) error {
					// DEBUG
					log.Println(msg)

					return db.Close()
				})(fmt.Sprintf("Closing DB, DB Name:[%s]", driverName)))
			},
		)
	}
}

/*
*sql.DB -> query :: string -> args :: []any -> (*sql.Rows -> Either error T) -> Either error T
 */
func withRows[T any](db *sql.DB) func(string) func([]any) func(func(*sql.Rows) E.Either[error, T]) E.Either[error, T] {
	return func(query string) func([]any) func(func(*sql.Rows) E.Either[error, T]) E.Either[error, T] {
		return func(args []any) func(func(*sql.Rows) E.Either[error, T]) E.Either[error, T] {
			return E.WithResource[error, *sql.Rows, T](
				func() E.Either[error, *sql.Rows] {
					return E.Eitherize2(func(q string, as []any) (*sql.Rows, error) {
						return db.Query(q, as...)
					})(query, args)
				},
				func(rows *sql.Rows) E.Either[error, any] {
					return E.Either[error, any](E.FromError(func(msg string) error {
						// DEBUG
						log.Println(msg)

						return rows.Close()
					})(fmt.Sprintf("Closing Rows, Query:[%s]", query)))
				},
			)
		}
	}
}

/*
rows2Users converts sql.Rows to a slice of User wrapped in Either.

*sql.Rows -> Either error []User
*/
func rows2Users(rows *sql.Rows) E.Either[error, []User] {
	users := make([]E.Either[error, User], 0)

	for rows.Next() {
		var id int
		var name, email string

		var userE E.Either[error, User]

		if err := rows.Scan(&id, &name, &email); err == nil {
			userE = E.Right[error](User{ID: id, Name: name, Email: email})
		} else {
			userE = E.Left[User](err)
		}
		users = append(users, userE)
	}

	return E.SequenceArray(users)
}

type NoUserFound struct{}

func (e NoUserFound) Error() string {
	return "No user found"
}

/*
*sql.Rows -> Either error Iterator[User]
 */
func rows2UsersWithIterator(rows *sql.Rows) E.Either[error, I.Iterator[User]] {
	resultE := E.Left[I.Iterator[User], error](NoUserFound{})
	users := make([]User, 0)

	for rows.Next() {
		var id int
		var name, email string

		if err := rows.Scan(&id, &name, &email); err == nil {
			users = append(users, User{ID: id, Name: name, Email: email})
		}
	}

	if len(users) > 0 {
		resultE = E.Right[error](I.FromArray(users))
	}

	// -------------------------- functional style -------------------------------------
	// resultE := E.Right[error](I.Empty[User]())
	// userIteratorMonoid := I.Monoid[User]()

	// for rows.Next() {
	// 	var id int
	// 	var name, email string

	// 	if err := rows.Scan(&id, &name, &email); err == nil {
	// 		resultE = E.Map[error](func(ui I.Iterator[User]) I.Iterator[User] {
	// 			return userIteratorMonoid.Concat(ui, I.Of(User{ID: id, Name: name, Email: email}))
	// 		})(resultE)
	// 	} else {
	// 		resultE = E.Chain(func(_ I.Iterator[User]) E.Either[error, I.Iterator[User]] {
	// 			return E.Left[I.Iterator[User], error](NoUserFound{})
	// 		})(resultE)
	// 	}
	// }

	return resultE
}

func TestDB(t *testing.T) {
	t.Run("Test rows to Either[error, []User]", func(t *testing.T) {
		ctx := context.Background()

		users, err := E.Unwrap(withContainer[[]User](ctx)("mariadb:11.0.3")([]TC.ContainerCustomizer{
			TCM.WithScripts(filepath.Join("testdata", "schema.sql")),
			TCM.WithDatabase("mydb"),
			TCM.WithUsername("root"),
			TCM.WithPassword(""),
		})(func(container *TCM.MariaDBContainer) E.Either[error, []User] {
			return F.Pipe1(
				E.Eitherize2(func(ctx context.Context, args []string) (string, error) {
					return container.ConnectionString(ctx, args...)
				})(ctx, A.Empty[string]()),
				E.Chain(func(connString string) E.Either[error, []User] {
					return withMariaDB[[]User](connString)("mysql")(func(db *sql.DB) E.Either[error, []User] {
						return withRows[[]User](db)("SELECT * FROM users")(A.Empty[any]())(rows2Users)
					})
				}),
			)
		}))

		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, "Alice", users[0].Name)
		assert.Equal(t, "alice@gmail.com", users[0].Email)
		assert.Equal(t, "Bob", users[1].Name)
		assert.Equal(t, "bob@gmail.com", users[1].Email)
	})

	t.Run("Test rows to Either[error, Iterator[User]]", func(t *testing.T) {
		ctx := context.Background()

		userIterator, err := E.Unwrap(withContainer[I.Iterator[User]](ctx)("mariadb:11.0.3")([]TC.ContainerCustomizer{
			TCM.WithScripts(filepath.Join("testdata", "schema.sql")),
			TCM.WithDatabase("mydb"),
			TCM.WithUsername("root"),
			TCM.WithPassword(""),
		})(func(container *TCM.MariaDBContainer) E.Either[error, I.Iterator[User]] {
			return F.Pipe1(
				E.Eitherize2(func(ctx context.Context, args []string) (string, error) {
					return container.ConnectionString(ctx, args...)
				})(ctx, A.Empty[string]()),
				E.Chain(func(connString string) E.Either[error, I.Iterator[User]] {
					return withMariaDB[I.Iterator[User]](connString)("mysql")(func(db *sql.DB) E.Either[error, I.Iterator[User]] {
						return withRows[I.Iterator[User]](db)("SELECT * FROM users")(A.Empty[any]())(rows2UsersWithIterator)
					})
				}),
			)
		}))

		assert.NoError(t, err)

		users := I.ToArray(userIterator)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, "Alice", users[0].Name)
		assert.Equal(t, "Bob", users[1].Name)
		assert.Equal(t, "alice@gmail.com", users[0].Email)
		assert.Equal(t, "bob@gmail.com", users[1].Email)
		assert.NotEqual(t, users[0].ID, users[1].ID) // IDs should be different
	})

	t.Run("Test finding no user", func(t *testing.T) {
		ctx := context.Background()

		_, err := E.Unwrap(withContainer[I.Iterator[User]](ctx)("mariadb:11.0.3")([]TC.ContainerCustomizer{
			TCM.WithScripts(filepath.Join("testdata", "no_data_schema.sql")),
			TCM.WithDatabase("mydb"),
			TCM.WithUsername("root"),
			TCM.WithPassword(""),
		})(func(container *TCM.MariaDBContainer) E.Either[error, I.Iterator[User]] {
			return F.Pipe1(
				E.Eitherize2(func(ctx context.Context, args []string) (string, error) {
					return container.ConnectionString(ctx, args...)
				})(ctx, A.Empty[string]()),
				E.Chain(func(connString string) E.Either[error, I.Iterator[User]] {
					return withMariaDB[I.Iterator[User]](connString)("mysql")(func(db *sql.DB) E.Either[error, I.Iterator[User]] {
						return withRows[I.Iterator[User]](db)("SELECT * FROM users")(A.Empty[any]())(rows2UsersWithIterator)
					})
				}),
			)
		}))

		assert.Error(t, err)
		assert.ErrorAs(t, err, &NoUserFound{}, "Expected NoUserFound error but got: %v", err)
	})
}
