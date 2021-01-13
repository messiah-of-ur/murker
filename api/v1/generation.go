package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/messiah-of-ur/murker/game"
)

type PlayerAuth struct {
	Key string `json:"key" binding:"required"`
}

func gameGenerationHandler(runner game.GameRunner, roomRegistry RoomRegistry) func(*gin.Context) {
	return func(c *gin.Context) {
		var auth PlayerAuth
		if err := c.ShouldBindJSON(&auth); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		gameID, controls, err := runner.AddGame()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Error adding game: %s", err.Error())})
		}

		mur := runner[gameID]

		room := NewRoom(auth.Key, mur, controls)
		roomRegistry[gameID] = room

		c.JSON(http.StatusOK, gin.H{"gameID": gameID})
	}
}
