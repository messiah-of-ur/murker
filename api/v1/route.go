package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/messiah-of-ur/murker/mur"
)

func RegisterHandlers(router *gin.Engine, runner mur.GameRunner, registry RoomRegistry) {
	router.GET("/state", stateHandler(runner))
	router.POST("/game", gameGenerationHandler(runner, registry))
	router.GET("/join/:game_id", registry.roomHandler())
}
