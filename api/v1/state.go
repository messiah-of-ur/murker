package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/messiah-of-ur/murker/game"
)

func stateHandler(runner game.GameRunner) func(*gin.Context) {
	return func(c *gin.Context) {
		gamesCount := len(runner)

		c.JSON(http.StatusOK, gin.H{
			"gameCount": gamesCount,
		})
	}
}
