package main

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	apiV1 "github.com/messiah-of-ur/murker/api/v1"

	"github.com/messiah-of-ur/murker/game"
)

func main() {
	rand.Seed(time.Now().Unix())

	router := gin.Default()
	runner := game.NewGameRunner()

	apiV1.RegisterHandlers(router, runner)

	router.Run()
}
