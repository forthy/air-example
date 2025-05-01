package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"air-example/internal/server"

	C "github.com/xlab/closer"

	E "github.com/IBM/fp-go/either"
	F "github.com/IBM/fp-go/function"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

// *http.Server -> ()
func gracefulShutdown(apiServer *http.Server) func() {
	return func() {
		r := E.Fold(
			func(err error) string {
				return fmt.Sprintf("Failed to shutdown server: %s", err)
			},
			func(_ context.Context) string {
				return "Server shutdown successfully"
			},
		)(E.FromError(apiServer.Shutdown)(context.Background()))

		log.Print(r)
	}
}

func main() {
	_, err := E.Unwrap(
		F.Pipe1(
			E.Eitherize0(server.NewServer)(),
			E.Chain(func(s *http.Server) E.Either[error, *http.Server] {
				C.Bind(gracefulShutdown(s))

				return E.FromError(func(s *http.Server) error {
					log.Println("Starting server on port", s.Addr)

					return s.ListenAndServe()
				})(s)
			}),
		),
	)

	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	C.Hold()
}
