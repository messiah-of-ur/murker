package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/messiah-of-ur/murker/game"
)

func RegisterHandlers(router *gin.Engine, runner game.GameRunner) {
	router.GET("/state", stateHandler(runner))
	router.POST("/game", gameGenerationHandler(runner))
}
