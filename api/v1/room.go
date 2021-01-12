package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/messiah-of-ur/murker/game"
)

const (
	NUM_PLAYERS = 2
	PLR_1       = 0
	PLR_2       = 1
)

type Room struct {
	mur                *game.Game
	controls           [NUM_PLAYERS]*game.GameController
	stateNotifications [NUM_PLAYERS]chan []byte
}

type MoveData struct {
	PawnID int `json:"pawnID"`
}

func NewRoom(mur *game.Game, controls [NUM_PLAYERS]*game.GameController) *Room {
	return &Room{
		mur:                mur,
		controls:           controls,
		stateNotifications: [NUM_PLAYERS]chan []byte{make(chan []byte), make(chan []byte)},
	}
}

func (r *Room) play(conn *websocket.Conn, plrID int) {
	defer func() {
		r.controls[plrID].Close()
		log.Printf("Player %d finished the game\n", plrID)
	}()

	go func() {
		for {
			gameState := <-r.stateNotifications[plrID]

			if err := conn.WriteMessage(websocket.TextMessage, gameState); err != nil {
				log.Printf("Error sending message over socket: %s", err.Error())
				return
			}
		}
	}()

	var pawnID int

	for {
		select {
		case <-r.controls[plrID].Turns:
			log.Printf("PL %d\n", plrID)

			_, body, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message over socket: %s", err.Error())
				return
			}

			var moveData MoveData
			if err := json.Unmarshal(body, &moveData); err != nil {
				log.Printf("Error unmarshaling move data: %s\n", err.Error())
				return
			}

			r.controls[plrID].Move(pawnID)

			fmt.Println(r.mur.PlrPawns)

			gameState, err := json.Marshal(r.mur)
			if err != nil {
				log.Printf("Error marshaling game state: %s\n", err.Error())
				return
			}

			r.stateNotifications[PLR_1] <- gameState
			r.stateNotifications[PLR_2] <- gameState

		case <-r.controls[plrID].End:
			return
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type RoomData struct {
	Key      string `json:"key" binding:"required"`
	GameID   string `json:"gameID" binding:"required"`
	PlayerID int    `json:"playerID" binding:"required"`
}

type RoomRegistry map[string]*Room

func (r RoomRegistry) roomHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

		if err != nil {
			log.Println(err)
			return
		}

		_, body, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message over socket: %s", err.Error())
			return
		}

		var data RoomData
		if err := json.Unmarshal(body, &data); err != nil {
			log.Printf("Error unmarshaling move data: %s\n", err.Error())
			return
		}

		room := r[data.GameID]

		go room.play(conn, data.PlayerID)
	}
}
