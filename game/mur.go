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
	plrPawns [2][PawnsPerPlayer]int
	roll     int
	turnPlr  int
	key      string
	pawns    []<-chan int
	turns    []chan<- struct{}
	end      chan<- struct{}
}

func NewGame(key string, pawns []<-chan int, turns []chan<- struct{}, end chan<- struct{}) *Game {
	return &Game{
		plrPawns: [2][PawnsPerPlayer]int{},
		roll:     0,
		turnPlr:  0,
		key:      key,
		pawns:    pawns,
		turns:    turns,
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
		fmt.Println(g.plrPawns)
		if !rosette {
			g.rollDice()
		} else {
			rosette = false
		}

		g.turns[g.turnPlr] <- struct{}{}
		pawnID := <-g.pawns[g.turnPlr]
		newField := g.move(g.turnPlr, pawnID)

		if newField == RosetteField {
			rosette = true
			continue
		}
		g.removePawns(opositePlayer(g.turnPlr), newField)

		if g.isGameFinished(g.turnPlr) {
			break
		}

		g.turnPlr = opositePlayer(g.turnPlr)
	}
}

func (g *Game) rollDice() {
	res := 0

	for i := 0; i < NumDice; i++ {
		res += rand.Intn(MaxDiceScore)
	}

	g.roll = res
}

func (g *Game) move(plrID int, pawnID int) int {
	pos := g.plrPawns[plrID][pawnID]
	pos = abs(pos)

	pos += g.roll
	pos = changeSideOfField(plrID, pos)

	g.plrPawns[plrID][pawnID] = pos
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

	for i, v := range g.plrPawns[plrID] {
		if v == fieldId {
			g.plrPawns[plrID][i] = 0
			break
		}
	}
}

func opositePlayer(plrID int) int { return 1 - plrID }

func (g *Game) isGameFinished(plrID int) bool {
	res := 0

	for i := 0; i < PawnsPerPlayer; i++ {
		if g.plrPawns[plrID][i] >= EscapedField {
			res++
		}
	}

	return (res == PawnsPerPlayer)
}
