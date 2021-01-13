package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	apiV1 "github.com/messiah-of-ur/murker/api/v1"
	"github.com/messiah-of-ur/murker/mur"
)

const MurkerPort = "MURKER_PORT"
const DefaultPort = "8080"

func main() {
	rand.Seed(time.Now().Unix())

	router := gin.Default()
	router.Use(cors.Default())

	runner := mur.NewGameRunner()
	registry := apiV1.RoomRegistry{}

	apiV1.RegisterHandlers(router, runner, registry)

	port := getPort()
	router.Run(port)
}

func getPort() string {
	port := os.Getenv(MurkerPort)
	if port == "" {
		port = DefaultPort
	}

	return ":" + port
}
