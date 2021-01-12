package game

import (
	"fmt"
	"math/rand"
)

const (
	PawnsPerPlayer = 8
	NumDice        = 4
	MaxDiceScore   = 2
	EscapedField   = 15
	RosetteField   = 8
)

type Game struct {
	PlrPawns [2][PawnsPerPlayer]int `json:"playerPawns" binding:"required"`
	Roll     int                    `json:"Roll" binding:"required"`
	TurnPlr  int                    `json:"turn" binding:"required"`
	key      string
	pawns    []<-chan int
	turns    []chan<- struct{}
	moveDone chan<- struct{}
	end      chan<- struct{}
}

func NewGame(key string, pawns []<-chan int, turns []chan<- struct{}, moveDone, end chan<- struct{}) *Game {
	return &Game{
		PlrPawns: [2][PawnsPerPlayer]int{},
		Roll:     0,
		TurnPlr:  0,
		key:      key,
		pawns:    pawns,
		turns:    turns,
		moveDone: moveDone,
		end:      end,
	}
}

func (g *Game) Run() {
	defer func() {
		close(g.end)

		<-g.pawns[0]
		<-g.pawns[1]

		close(g.turns[0])
		close(g.turns[1])
	}()

	rosette := false
	for {
		fmt.Println(g.PlrPawns)
		if !rosette {
			g.RollDice()
		} else {
			rosette = false
		}

		g.turns[g.TurnPlr] <- struct{}{}
		pawnID := <-g.pawns[g.TurnPlr]
		newField := g.move(g.TurnPlr, pawnID)

		if newField == RosetteField {
			rosette = true
			continue
		}
		g.removePawns(opositePlayer(g.TurnPlr), newField)

		if g.isGameFinished(g.TurnPlr) {
			break
		}

		g.TurnPlr = opositePlayer(g.TurnPlr)

		g.moveDone <- struct{}{}
	}
}

func (g *Game) RollDice() {
	res := 0

	for i := 0; i < NumDice; i++ {
		res += rand.Intn(MaxDiceScore)
	}

	g.Roll = res
}

func (g *Game) move(plrID int, pawnID int) int {
	pos := g.PlrPawns[plrID][pawnID]
	pos = abs(pos)

	pos += g.Roll
	pos = changeSideOfField(plrID, pos)

	g.PlrPawns[plrID][pawnID] = pos
	return pos
}

func abs(input int) int {
	if input < 0 {
		return input * -1
	}
	return input
}

func changeSideOfField(plrID int, fieldID int) int {
	if plrID == 0 {
		return fieldID
	}

	if fieldID < 5 || fieldID == 13 || fieldID == 14 {
		return fieldID * -1
	}

	return fieldID
}

func (g *Game) removePawns(plrID int, fieldId int) {
	if fieldId >= EscapedField {
		return
	}

	for i, v := range g.PlrPawns[plrID] {
		if v == fieldId {
			g.PlrPawns[plrID][i] = 0
			break
		}
	}
}

func opositePlayer(plrID int) int { return 1 - plrID }

func (g *Game) isGameFinished(plrID int) bool {
	res := 0

	for i := 0; i < PawnsPerPlayer; i++ {
		if g.PlrPawns[plrID][i] >= EscapedField {
			res++
		}
	}

	return (res == PawnsPerPlayer)
}
