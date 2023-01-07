package game

import "github.com/renju24/backend/internal/pkg/apierror"

const BoardSize = 15

// Move structure.
type Move struct {
	x     int   // X-coordinate.
	y     int   // Y-coordinate.
	color Color // color.
}

type Color int

const (
	Nil   Color = 0
	Black Color = 1
	White Color = 2
)

func NewMove(x, y int, color Color) Move {
	return Move{
		x:     x,
		y:     y,
		color: color,
	}
}

// Game structure.
type Game struct {
	board    [BoardSize * BoardSize]Color // Board 15x15.
	lastMove Move                         // The last move.
}

func NewGame() *Game {
	return &Game{}
}

func (g *Game) ApplyMove(move Move) (winner Color, err error) {
	// If it's the first move, then user should be black and move should be in board's center.
	if g.lastMove.color == Nil {
		if move.color != Black {
			return Nil, apierror.ErrFirstMoveShouldBeBlack
		}
		if move.x != 7 || move.y != 7 {
			return Nil, apierror.ErrFirstMoveShouldBeInCenter
		}
	}
	// Check the coordinates are not outside the board.
	if move.x >= BoardSize || move.x < 0 || move.y >= BoardSize || move.y < 0 {
		return Nil, apierror.ErrCoordinatesOutside
	}
	// Check the field is not already taken.
	if g.board[move.x*BoardSize+move.y] != Nil {
		return Nil, apierror.ErrFieldAlreadyTaken
	}
	// Check the last move was made by another player.
	if g.lastMove.color == move.color {
		return Nil, apierror.ErrInvalidTurn
	}

	// Apply the move and change the board.
	g.board[move.x*BoardSize+move.y] = move.color
	g.lastMove = move

	// After a successful move, we should check if there is a winner.
	if g.hasWinner() {
		return g.lastMove.color, nil
	}

	return Nil, nil
}

func (game *Game) hasWinner() bool {
	var xcount, ycount, zcount int
	x, y := game.lastMove.x, game.lastMove.y
	x2, y2 := game.lastMove.x, game.lastMove.y
	if game.lastMove.color == Nil {
		return false
	}
	for i := 0; i < BoardSize; i++ {
		// "-"
		if xcount == 5 {
			return true
		}
		if game.board[x*BoardSize+i] == game.lastMove.color {
			xcount++
		} else {
			xcount = 0
		}
		// "|"
		if ycount == 5 {
			return true
		}
		if game.board[i*BoardSize+y] == game.lastMove.color {
			ycount++
		} else {
			ycount = 0
		}
	}
	// "/"
	for x2 > 0 && y2 < BoardSize {
		x2--
		y2++
	}
	for x2 < BoardSize && y2 > 0 {
		if zcount == 5 {
			return true
		}
		if game.board[x2*BoardSize+y2] == game.lastMove.color {
			zcount++
		} else {
			zcount = 0
		}
		x2++
		y2--
	}
	zcount = 0
	// "\"
	for x > 0 && y > 0 {
		x--
		y--
	}
	for x < BoardSize && y < BoardSize {
		if zcount == 5 {
			return true
		}
		if game.board[x*BoardSize+y] == game.lastMove.color {
			zcount++
		} else {
			zcount = 0
		}
		x++
		y++
	}
	return false
}
