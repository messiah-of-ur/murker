package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	apiV1 "github.com/messiah-of-ur/murker/api/v1"
	"github.com/messiah-of-ur/murker/mur"
	"github.com/messiah-of-ur/murker/murabi"
)

const MurkerPort = "MURKER_PORT"
const MurabiAddr = "MURABI_ADDR"
const DefaultPort = "8080"

func main() {
	rand.Seed(time.Now().Unix())

	router := gin.Default()
	router.Use(cors.Default())

	runner := mur.NewGameRunner()
	registry := apiV1.RoomRegistry{}

	murabiAddr := getMurabiAddr()
	murabiClient := murabi.NewMurabiClient(murabiAddr)

	apiV1.RegisterHandlers(router, runner, registry, murabiClient)

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

func getMurabiAddr() string {
	addr := os.Getenv(MurabiAddr)
	if addr == "" {
		log.Fatalf("Expected a murabi address as a config param")
	}

	return addr
}
