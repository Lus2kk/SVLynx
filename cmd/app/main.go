package main

import (
	"log/slog"
	"github.com/svlynx/messenger/internal/auth_handler"
	"github.com/svlynx/messenger/internal/config"
)



func main(){
    cfg := config.MustLoad()

    server := auth_handler.NewServer(cfg)
    defer server.Close()  

    if err := server.Run(); err != nil {
        slog.Warn(err.Error())
    }
}



