package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/messiah-of-ur/murker/game"
)

func RegisterHandlers(router *gin.Engine, runner game.GameRunner, registry RoomRegistry) {
	router.GET("/state", stateHandler(runner))
	router.POST("/game", gameGenerationHandler(runner, registry))
	router.GET("/room", registry.roomHandler())
}
