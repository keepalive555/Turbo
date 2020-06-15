package main

import (
	"os"
	"strconv"
	_ "turbo/client"
	"turbo/server"
)

func main() {
	p := os.Getenv("TURBO_PORT")
	port, err := strconv.ParseInt(p, 10, 64)
	if err != nil {
		port = 0
	}
	config := &server.TcpConfig{
		Host: os.Getenv("TURBO_HOST"),
		Port: int(port),
	}
	if err := config.Check(); err != nil {
		panic(err)
	}
	srv, err := server.NewServer(server.ProtoTCP, config)
	if err != nil {
		panic(err)
	}
	srv.Run()
}
