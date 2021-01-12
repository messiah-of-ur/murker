package game

import (
	"fmt"

	"github.com/google/uuid"
)

type GameController struct {
	End   <-chan struct{}
	Turns <-chan struct{}
	pawns chan<- int
}

func newGameControls(end <-chan struct{}, turns <-chan struct{}, pawns chan<- int) *GameController {
	return &GameController{
		End:   end,
		Turns: turns,
		pawns: pawns,
	}
}

func (r *GameController) Move(pawnID int) {
	r.pawns <- pawnID
}

func (r *GameController) Close() {
	close(r.pawns)
}

type GameRunner map[string]*Game

func NewGameRunner() GameRunner {
	return GameRunner{}
}

func (r GameRunner) AddGame(key string) (gameID string, controls [2]*GameController, err error) {
	if gameID, err = uniqGameID(); err != nil {
		return "", controls, err
	}

	pawns := []chan int{make(chan int), make(chan int)}
	turns := []chan struct{}{make(chan struct{}), make(chan struct{})}
	end := make(chan struct{})

	controls = [2]*GameController{newGameControls(end, turns[0], pawns[0]), newGameControls(end, turns[1], pawns[1])}

	r[gameID] = NewGame(key, []<-chan int{pawns[0], pawns[1]}, []chan<- struct{}{turns[0], turns[1]}, end)

	go r.runGame(r[gameID], gameID)

	return gameID, controls, nil
}

func (r GameRunner) runGame(game *Game, gameID string) {
	defer func() {
		r.removeGame(gameID)
	}()
	game.Run()
}

func (r GameRunner) removeGame(gameID string) {
	delete(r, gameID)
}

func uniqGameID() (string, error) {
	uniqID, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("Error generating uuid: %s", err.Error())
	}

	return uniqID.String(), nil
}
