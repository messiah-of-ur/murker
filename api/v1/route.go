package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/messiah-of-ur/murker/mur"
)

func RegisterHandlers(router *gin.Engine, runner mur.GameRunner, registry RoomRegistry) {
	router.GET("/v1/state", stateHandler(runner))
	router.POST("/v1/game", gameGenerationHandler(runner, registry))
	router.GET("/v1/join/:game_id", registry.roomHandler())
}
