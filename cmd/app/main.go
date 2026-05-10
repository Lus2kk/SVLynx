package main

import (
	"log/slog"

	"github.com/svlynx/messenger/internal/config"
	"github.com/svlynx/messenger/internal/server"
)

func main() {
	cfg := config.MustLoad()
	server := server.NewServer(cfg)
    defer server.Close()  

    if err := server.Run(); err != nil {
        slog.Warn(err.Error())
    }
	

}



