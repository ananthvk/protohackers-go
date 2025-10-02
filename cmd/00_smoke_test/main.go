package main

import (
	"context"
	"flag"

	"github.com/ananthvk/protohackers-go/internal/tcp"
)

func main() {
	server := tcp.NewServer(context.Background())
	server.AddToFlags()
	flag.Parse()
	server.LoadFromFlags()
	server.ListenAndServe()
}
