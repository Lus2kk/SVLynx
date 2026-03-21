package main

import (
	"log"
	"github.com/svlynx/messenger/internal/config"

	"github.com/svlynx/messenger/internal/auth_handler"
)

func main(){
	cfg := config.MustLoad()
	server := auth_handler.NewServer(cfg)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}

}