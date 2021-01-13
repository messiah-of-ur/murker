package game

import (
	"fmt"

	"github.com/google/uuid"
)

type GameController struct {
	End       <-chan struct{}
	Turns     <-chan struct{}
	Interrupt chan struct{}
	moveDone  <-chan struct{}
	pawns     chan<- int
}

func newGameControls(end, turns, moveDone <-chan struct{}, pawns chan<- int, interrupt chan struct{}) *GameController {
	return &GameController{
		End:       end,
		Turns:     turns,
		Interrupt: interrupt,
		moveDone:  moveDone,
		pawns:     pawns,
	}
}

func (g *GameController) Move(pawnID int) {
	g.pawns <- pawnID
	<-g.moveDone
}

func (g *GameController) Close() {
	close(g.pawns)
}

type GameRunner map[string]*Game

func NewGameRunner() GameRunner {
	return GameRunner{}
}

func (r GameRunner) AddGame() (gameID string, controls [2]*GameController, err error) {
	if gameID, err = uniqGameID(); err != nil {
		return "", controls, err
	}

	pawns := []chan int{make(chan int), make(chan int)}
	turns := []chan struct{}{make(chan struct{}), make(chan struct{})}
	moveDone := make(chan struct{})
	end := make(chan struct{})
	interrupt := []chan struct{}{make(chan struct{}), make(chan struct{})}

	controls = [2]*GameController{
		newGameControls(end, turns[0], moveDone, pawns[0], interrupt[0]),
		newGameControls(end, turns[1], moveDone, pawns[1], interrupt[0]),
	}

	r[gameID] = NewGame(
		[]<-chan int{pawns[0], pawns[1]},
		[]chan<- struct{}{turns[0], turns[1]},
		moveDone,
		end,
		[]<-chan struct{}{interrupt[0], interrupt[1]},
	)

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
