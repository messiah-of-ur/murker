package mur

import (
	"fmt"
	"math/rand"
)

const (
	MaxPlayerPawns = 8
	NumDice        = 4
	MaxDiceScore   = 2
	EscapedField   = 15
	RosetteField   = 8
)

type Game struct {
	PlrPawns  [2][MaxPlayerPawns]int `json:"playerPawns"`
	Roll      int                    `json:"roll"`
	TurnPlr   int                    `json:"turn"`
	pawns     []<-chan int
	turns     []chan<- struct{}
	moveDone  chan<- struct{}
	end       chan<- struct{}
	interrupt <-chan struct{}
}

func NewGame(pawns []<-chan int, turns []chan<- struct{}, moveDone, end chan<- struct{}, interrupt <-chan struct{}) *Game {
	return &Game{
		PlrPawns:  [2][MaxPlayerPawns]int{},
		Roll:      0,
		TurnPlr:   0,
		pawns:     pawns,
		turns:     turns,
		moveDone:  moveDone,
		end:       end,
		interrupt: interrupt,
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
	var pawnID int

	for {
		fmt.Println(g.PlrPawns)
		if !rosette {
			g.RollDice()
		} else {
			rosette = false
		}

		g.turns[g.TurnPlr] <- struct{}{}

		select {
		case pawnID = <-g.pawns[g.TurnPlr]:
		case <-g.interrupt:
			return
		}

		newField := g.move(g.TurnPlr, pawnID)

		if newField == RosetteField {
			rosette = true
			continue
		}
		g.removePawns(OpositePlayer(g.TurnPlr), newField)

		if g.isGameFinished(g.TurnPlr) {
			break
		}

		g.TurnPlr = OpositePlayer(g.TurnPlr)

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

func OpositePlayer(plrID int) int { return 1 - plrID }

func (g *Game) isGameFinished(plrID int) bool {
	res := 0

	for i := 0; i < MaxPlayerPawns; i++ {
		if g.PlrPawns[plrID][i] >= EscapedField {
			res++
		}
	}

	return (res == MaxPlayerPawns)
}
