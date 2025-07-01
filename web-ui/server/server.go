package server

import (
	"fmt"
	"net/http"
	"serina/engine"
)

type Server struct {
	mux   *http.ServeMux
	chess *engine.Chess
}

func NewServer() *Server {
	return &Server{
		mux:   http.NewServeMux(),
		chess: engine.NewChess(),
	}
}

func (server *Server) Config() {
	//Setup file server
	fileSrv := http.Dir("./web-ui/client")
	server.mux.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./web-ui/client/assets"))))
	server.mux.Handle("/", http.FileServer(fileSrv))

	//Setup API endpoint
	server.mux.HandleFunc("/fen/", server.HandleFEN)
	server.mux.HandleFunc("/flip/", server.HandleFlipBoard)
	server.mux.HandleFunc("/move", server.HandleMove)
	server.mux.HandleFunc("/perft", server.HandlePerft)
	server.mux.HandleFunc("/search", server.HandleSearch)
}

func (server *Server) Start() {
	server.Config()

	fmt.Println("Server start at http://localhost:8080 ...")
	err := http.ListenAndServe(":8080", server.mux)
	if err != nil {
		fmt.Println("Error starting server")
		fmt.Println(err)
		return
	}
}
