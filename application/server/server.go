package server

import (
	"context"
	"github.com/jackc/pgx/pgxpool"
	"github.com/labstack/echo"
	"github.com/moroz-matros/technopark_db/application/app/delivery/http"
	"github.com/moroz-matros/technopark_db/application/app/repository"
	"github.com/moroz-matros/technopark_db/application/app/usecase"
	"github.com/moroz-matros/technopark_db/pkg/constants"
	"log"
)

type Server struct {
	e       *echo.Echo
}

func NewServer() *Server {
	var server Server

	e := echo.New()


	pool, err := pgxpool.Connect(context.Background(), constants.DBConnect)
	if err != nil {
		log.Fatal(err)
	}
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	rep := repository.NewDatabase(pool)

	uc := usecase.NewForum(rep)

	http.CreateForumHandler(e, uc)

	return &server
}

func (s Server) ListenAndServe() {
	s.e.Logger.Fatal(s.e.Start(":5000"))
}