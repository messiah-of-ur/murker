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
	mur       *game.Game
	controls  [NUM_PLAYERS]*game.GameController
	sendState [NUM_PLAYERS]chan []byte
	joined    [NUM_PLAYERS]chan struct{}
}

type PlayerAction struct {
	PawnID *int `json:"pawnID"`
}

func NewRoom(mur *game.Game, controls [NUM_PLAYERS]*game.GameController) *Room {
	return &Room{
		mur:       mur,
		controls:  controls,
		sendState: [NUM_PLAYERS]chan []byte{make(chan []byte), make(chan []byte)},
		joined:    [NUM_PLAYERS]chan struct{}{make(chan struct{}), make(chan struct{})},
	}
}

func (r *Room) play(ctx *gin.Context, conn *websocket.Conn, plrID int) {
	defer func() {
		r.controls[plrID].Close()
		close(r.sendState[plrID])
		conn.Close()
		log.Printf("Player %d finished the game\n", plrID)
	}()

	go r.runMessageSender(ctx, conn, plrID)

	r.waitForOpponentToJoin(plrID)

	for {
		select {
		case <-r.controls[plrID].Turns:
			if err := r.sendGameState(); err != nil {
				r.interruptWithError(ctx, plrID, err)
				return
			}

			var action PlayerAction
			if err := readSocketData(conn, &action); err != nil {
				r.interruptWithError(ctx, plrID, fmt.Errorf("Error reading player %d action: %s", plrID, err.Error()))
				return
			}

			if action.PawnID == nil {
				r.interruptWithError(ctx, plrID, fmt.Errorf("Missing required field pawnID for player %d", plrID))
				return
			}

			r.controls[plrID].Move(*action.PawnID)

		case <-r.controls[plrID].End:
			return
		case <-r.controls[plrID].Interrupt:
			return
		}
	}
}

func (r *Room) waitForOpponentToJoin(plrID int) {
	close(r.joined[plrID])
	<-r.joined[1-plrID]
}

func (r *Room) runMessageSender(ctx *gin.Context, conn *websocket.Conn, plrID int) {
	for {
		select {
		case gameState := <-r.sendState[plrID]:
			if err := conn.WriteMessage(websocket.TextMessage, gameState); err != nil {
				r.interruptWithError(ctx, plrID, fmt.Errorf("Error sending message over socket: %s", err.Error()))
				return
			}
		case <-r.controls[plrID].End:
			return
		case <-r.controls[plrID].Interrupt:
			return
		}
	}
}

func (r *Room) sendGameState() error {
	gameState, err := json.Marshal(r.mur)
	if err != nil {
		return fmt.Errorf("Error marshaling game state: %s\n", err.Error())
	}

	r.sendState[PLR_1] <- gameState
	r.sendState[PLR_2] <- gameState

	return nil
}

func (r *Room) interruptWithError(ctx *gin.Context, plrID int, err error) {
	fmt.Println("Closing int channel for ", plrID)
	log.Println(err.Error())
	r.interruptOpponent(plrID)
}

func (r *Room) interruptOpponent(plrID int) {
	opp := game.OpositePlayer(plrID)
	close(r.controls[opp].Interrupt)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type RoomCredentials struct {
	Key      string `json:"key" binding:"required"`
	PlayerID int    `json:"playerID" binding:"required"`
}

type RoomRegistry map[string]*Room

func (r RoomRegistry) roomHandler() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Error upgrading websocket: %s", err.Error())})
			return
		}

		var credentials RoomCredentials
		if err := readSocketData(conn, &credentials); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("Error reading client credentials: %s", err.Error())})
			return
		}

		gameID := ctx.Param("game_id")
		room := r[gameID]

		go room.play(ctx, conn, credentials.PlayerID)
	}
}

func readSocketData(conn *websocket.Conn, dataModel interface{}) error {
	msgType, body, err := conn.ReadMessage()

	if msgType != websocket.TextMessage {
		return fmt.Errorf("Unsupported socket message type: %d", msgType)
	}

	if err != nil {
		return fmt.Errorf("Error reading message over socket: %s", err.Error())
	}

	log.Print("Parse")
	if err := json.Unmarshal(body, dataModel); err != nil {
		log.Print("Parse error")
		return fmt.Errorf("Error unmarshaling move data: %s\n", err.Error())
	}

	return nil
}
