package main

import (
	"github.com/FirdavsMF/gosql/pkg/security"
	"github.com/FirdavsMF/gosql/pkg/customers"
	"github.com/FirdavsMF/gosql/cmd/app"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/dig"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	dsn := "postgres://app:pass@localhost:5432/db"

	if err := execute(host, port, dsn); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(host, port, dsn string) (err error) {
	deps := []interface{}{
		app.NewServer,
		mux.NewRouter,
		func() (*pgxpool.Pool, error) {
			connCtx,err:=context.WithTimeout(context.Background(),time.Second*5)
			if err != nil {
				log.Print(err)
			}
			return pgxpool.Connect(connCtx,dsn)
			// connCtx, _ := context.WithTimeout(context.Background(), time.Second * 5)
			// return pgxpool.Connect(connCtx, dsn)
		},
		customers.NewService,
		security.NewService,
		func(server *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}

	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			return err
		}
	}

	err = container.Invoke(func(server *app.Server) { 
		server.Init() 
	})
	if err != nil {
		return err
	}

	return container.Invoke(func(s *http.Server) error { 
		return s.ListenAndServe() 
	})
}