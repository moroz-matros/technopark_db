package main

import "github.com/moroz-matros/technopark_db/application/server"

func main() {
	s := server.NewServer()
	s.ListenAndServe()
}
