package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/messiah-of-ur/murker/mur"
	"github.com/messiah-of-ur/murker/murabi"
)

type PlayerAuth struct {
	Key string `json:"key" binding:"required"`
}

func gameGenerationHandler(runner mur.GameRunner, roomRegistry RoomRegistry, murabiClient *murabi.MurabiClient) func(*gin.Context) {
	return func(c *gin.Context) {
		var auth PlayerAuth
		if err := c.ShouldBindJSON(&auth); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		gameID, controls, interrupt, err := runner.AddGame()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Error adding game: %s", err.Error())})
		}

		game := runner[gameID]
		roomRegistry.addRoom(auth.Key, game, controls, interrupt, gameID, murabiClient)

		c.JSON(http.StatusOK, gin.H{"gameID": gameID})
	}
}
