package main

import (
	_ "turbo/client"
	"turbo/server"
)

func main() {
	config := &server.TcpConfig{
		Host: "0.0.0.0",
		Port: 8000,
	}
	srv, err := server.NewServer(server.ProtoTCP, config)
	if err != nil {
		panic(err)
	}
	srv.Run()
}
