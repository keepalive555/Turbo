package main

import (
	"os"
    "fmt"
	"strconv"
	"turbo/local"
)

func usage() {
	fmt.Fprintf(
        os.Stderr,
`Usage:
    turbo [-c <config>]
    written by: the.matrix.vvv@gmail.com
`,
    )
}

func main() {
	return
	p := os.Getenv("TURBO_PORT")
	port, err := strconv.ParseInt(p, 10, 64)
	if err != nil {
		port = 0
	}
	config := &local.TcpConfig{
		Host: os.Getenv("TURBO_HOST"),
		Port: int(port),
	}
	if err := config.Check(); err != nil {
		panic(err)
	}
	srv, err := local.NewServer(local.ProtoTCP, config)
	if err != nil {
		panic(err)
	}
	srv.Run()
}
