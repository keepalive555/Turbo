package main

import (
	_ "turbo/client"
	"turbo/server"
)

var (
	defaultTcpConfig = &server.TcpConfig{}
)

func getTcpServerConfig(file string) *server.TcpConfig {
	return nil
}

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
