package main

import (
	"bank/internal/api/misc"
	"bank/internal/config"
	"bank/internal/database"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(db) // temporary

	router := gin.Default()
	router.GET(misc.EndpointWellKnownBank, misc.PingV1dot0)

	err = router.Run(":27462")
	if err != nil {
		log.Fatal(err)
	}
}
