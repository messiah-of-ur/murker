package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/messiah-of-ur/murker/mur"
)

const (
	NUM_PLAYERS = 2
	PLR_1       = 0
	PLR_2       = 1
)

type Room struct {
	key         string
	mur         *mur.Game
	controls    [NUM_PLAYERS]*mur.GameController
	sendState   [NUM_PLAYERS]chan []byte
	opponent    [NUM_PLAYERS]chan struct{}
	joined      [NUM_PLAYERS]bool
	interrupt   chan struct{}
	interrupted bool
	mu          sync.Mutex
}

type PlayerAction struct {
	PawnID *int `json:"pawnID"`
}

func newRoom(key string, mur *mur.Game, controls [NUM_PLAYERS]*mur.GameController, interrupt chan struct{}) *Room {
	return &Room{
		key:         key,
		mur:         mur,
		controls:    controls,
		sendState:   [NUM_PLAYERS]chan []byte{make(chan []byte), make(chan []byte)},
		opponent:    [NUM_PLAYERS]chan struct{}{make(chan struct{}), make(chan struct{})},
		joined:      [NUM_PLAYERS]bool{false, false},
		interrupt:   interrupt,
		interrupted: false,
	}
}

func (r *Room) play(ctx *gin.Context, conn *websocket.Conn, plrID int) {
	defer func() {
		r.controls[plrID].Close()
		close(r.sendState[plrID])
		conn.Close()
		r.closeInterrupt()

		log.Printf("Player %d finished the game\n", plrID)
	}()

	go r.runMessageSender(ctx, conn, plrID)

	r.waitForOpponentToJoin(plrID)

	for {
		select {
		case <-r.controls[plrID].Turns:
			if err := r.sendGameState(); err != nil {
				r.interruptWithError(ctx, err)
				return
			}

			var action PlayerAction
			if err := readSocketData(conn, &action); err != nil {
				r.interruptWithError(ctx, fmt.Errorf("Error reading player %d action: %s", plrID, err.Error()))
				return
			}

			if action.PawnID == nil {
				r.interruptWithError(ctx, fmt.Errorf("Missing required field pawnID for player %d", plrID))
				return
			}

			r.controls[plrID].Move(*action.PawnID)

		case <-r.controls[plrID].End:
			return
		case <-r.interrupt:
			return
		}
	}
}

func (r *Room) isKeyAuthentic(key string) bool {
	return r.key == key
}

func (r *Room) waitForOpponentToJoin(plrID int) {
	close(r.opponent[plrID])
	r.joined[plrID] = true

	<-r.opponent[1-plrID]
}

func (r *Room) runMessageSender(ctx *gin.Context, conn *websocket.Conn, plrID int) {
	for {
		select {
		case gameState := <-r.sendState[plrID]:
			if err := conn.WriteMessage(websocket.TextMessage, gameState); err != nil {
				r.interruptWithError(ctx, fmt.Errorf("Error sending message over socket: %s", err.Error()))
				return
			}
		case <-r.controls[plrID].End:
			return
		case <-r.interrupt:
			log.Printf("Closed msg sender of plr%d", plrID)
			return
		}
	}
}

func (r *Room) sendGameState() error {
	gameState, err := json.Marshal(r.mur)
	if err != nil {
		return fmt.Errorf("Error marshaling game state: %s", err.Error())
	}

	r.sendState[PLR_1] <- gameState
	r.sendState[PLR_2] <- gameState

	return nil
}

func (r *Room) interruptWithError(ctx *gin.Context, err error) {
	log.Printf("Closing int channel: %s\n", err.Error())
	r.closeInterrupt()
}

func (r *Room) closeInterrupt() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.interrupted {
		close(r.interrupt)
		r.interrupted = true
	}
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

func (r RoomRegistry) addRoom(
	key string,
	gameInstance *mur.Game,
	controls [2]*mur.GameController,
	interrupt chan struct{},
	gameID string,
) {
	room := newRoom(key, gameInstance, controls, interrupt)
	r[gameID] = room

	go func() {
		<-interrupt
		delete(r, gameID)
	}()
}

func (r RoomRegistry) roomHandler() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)

		if err != nil {
			closeConnectionBeforeJoining(conn, fmt.Errorf("Error upgrading websocket: %s", err.Error()))
			return
		}

		var credentials RoomCredentials
		if err := readSocketData(conn, &credentials); err != nil {
			closeConnectionBeforeJoining(conn, fmt.Errorf("Error reading client credentials: %s", err.Error()))
			return
		}

		gameID := ctx.Param("game_id")

		room, ok := r[gameID]

		if !ok {
			closeConnectionBeforeJoining(conn, fmt.Errorf("Error trying to connect to non-existing game room %s", gameID))
			return
		}

		if room.joined[credentials.PlayerID] {
			closeConnectionBeforeJoining(conn, fmt.Errorf("Player %d tried to join the game %s again", credentials.PlayerID, gameID))
			return
		}

		if !room.isKeyAuthentic(credentials.Key) {
			closeConnectionBeforeJoining(conn, fmt.Errorf("Player %d tried to join the game %s with invalid key", credentials.PlayerID, gameID))
			return
		}

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

	if err := json.Unmarshal(body, dataModel); err != nil {
		return fmt.Errorf("Error unmarshaling move data: %s", err.Error())
	}

	return nil
}

func closeConnectionBeforeJoining(conn *websocket.Conn, err error) {
	log.Println(err)
	conn.Close()
}
