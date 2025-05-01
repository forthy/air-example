package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	E "github.com/IBM/fp-go/either"
	F "github.com/IBM/fp-go/function"
)

type Server struct {
	port int
}

func NewServer() (*http.Server, error) {
	return E.Unwrap(
		F.Pipe2(
			E.Eitherize1(strconv.Atoi)(os.Getenv("PORT")),
			E.Map[error](func(port int) *Server {
				return &Server{
					port: port,
				}
			}),
			E.Map[error](func(s *Server) *http.Server {
				return &http.Server{
					Addr:         fmt.Sprintf(":%d", s.port),
					Handler:      s.RegisterRoutes(),
					IdleTimeout:  time.Minute,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 30 * time.Second,
				}
			}),
		),
	)
}
