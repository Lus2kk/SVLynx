package main

import (
	"log/slog"
	"github.com/svlynx/messenger/internal/config"
	"github.com/svlynx/messenger/internal/service"
)

func main() {
	cfg := config.MustLoad()
	server := service.NewServer(cfg)
    defer server.Close()  

    if err := server.Run(); err != nil {
        slog.Warn(err.Error())
    }
	

}



