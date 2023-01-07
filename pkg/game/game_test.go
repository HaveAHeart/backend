package game

import (
	"testing"

	"github.com/renju24/backend/internal/pkg/apierror"
	"github.com/stretchr/testify/require"
)

func TestGame(t *testing.T) {
	testCases := []struct {
		moves          []Move
		expectedWinner Color
		expectedError  error
	}{
		// Case when first move is not black.
		{
			moves: []Move{
				NewMove(1, 5, White),
			},
			expectedWinner: Nil,
			expectedError:  apierror.ErrFirstMoveShouldBeBlack,
		},
		// Case when first move is not in center.
		{
			moves: []Move{
				NewMove(1, 2, Black),
			},
			expectedWinner: Nil,
			expectedError:  apierror.ErrFirstMoveShouldBeInCenter,
		},
		// Case when the X-coordinate is outside the board.
		{
			moves: []Move{
				NewMove(7, 7, Black),
				NewMove(-1, 5, White),
			},
			expectedWinner: Nil,
			expectedError:  apierror.ErrCoordinatesOutside,
		},
		// Case when the Y-coordinate is outside the board.
		{
			moves: []Move{
				NewMove(7, 7, Black),
				NewMove(5, 15, Black),
			},
			expectedWinner: Nil,
			expectedError:  apierror.ErrCoordinatesOutside,
		},
		// Case when field is already taken.
		{
			moves: []Move{
				NewMove(7, 7, Black),
				NewMove(7, 7, White),
			},
			expectedWinner: Nil,
			expectedError:  apierror.ErrFieldAlreadyTaken,
		},
		// Case when the last move was made by same player.
		{
			moves: []Move{
				NewMove(7, 7, Black),
				NewMove(0, 0, White),
				NewMove(1, 0, White),
			},
			expectedWinner: Nil,
			expectedError:  apierror.ErrInvalidTurn,
		},
		// Case when black should win.
		{
			moves: []Move{
				NewMove(7, 7, Black),
				NewMove(8, 8, White),
				NewMove(0, 0, Black),
				NewMove(1, 0, White),
				NewMove(0, 1, Black),
				NewMove(2, 0, White),
				NewMove(0, 2, Black),
				NewMove(3, 0, White),
				NewMove(0, 3, Black),
				NewMove(4, 0, White),
				NewMove(0, 4, Black),
			},
			expectedWinner: Black,
			expectedError:  nil,
		},
		// Case when white should win.
		{
			moves: []Move{
				NewMove(7, 7, Black),
				NewMove(8, 8, White),
				NewMove(0, 0, Black),
				NewMove(1, 0, White),
				NewMove(0, 1, Black),
				NewMove(2, 0, White),
				NewMove(0, 2, Black),
				NewMove(3, 0, White),
				NewMove(0, 3, Black),
				NewMove(4, 0, White),
				NewMove(0, 5, Black),
				NewMove(5, 0, White),
			},
			expectedWinner: White,
			expectedError:  nil,
		},
	}
	for _, testCase := range testCases {
		game := NewGame()
		for i, move := range testCase.moves {
			actualWinner, actualErr := game.ApplyMove(move)
			if i+1 == len(testCase.moves) {
				require.Equal(t, testCase.expectedWinner, actualWinner)
				require.ErrorIs(t, testCase.expectedError, actualErr)
			} else {
				require.Equal(t, Nil, actualWinner)
				require.NoError(t, actualErr)
			}
		}
	}
}
