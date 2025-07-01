package engine

import (
	"math/bits"
)

var (
	//Material value of each piece type (centipawn)
	material = map[int]int{
		WHITE_PAWN:   100,
		WHITE_ROOK:   500,
		WHITE_KNIGHT: 320,
		WHITE_BISHOP: 330,
		WHITE_QUEEN:  900,
		WHITE_KING:   20000,
	}

	//Piece-square value (https://www.chessprogramming.org/Simplified_Evaluation_Function)
	piece_square_table = map[int][]int{
		WHITE_PAWN: {
			0, 0, 0, 0, 0, 0, 0, 0,
			50, 50, 50, 50, 50, 50, 50, 50,
			10, 10, 20, 30, 30, 20, 10, 10,
			5, 5, 10, 25, 25, 10, 5, 5,
			0, 0, 0, 20, 20, 0, 0, 0,
			5, -5, -10, 0, 0, -10, -5, 5,
			5, 10, 10, -20, -20, 10, 10, 5,
			0, 0, 0, 0, 0, 0, 0, 0,
		},

		WHITE_ROOK: {
			0, 0, 0, 0, 0, 0, 0, 0,
			5, 10, 10, 10, 10, 10, 10, 5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			0, 0, 0, 5, 5, 0, 0, 0,
		},

		WHITE_KNIGHT: {
			-50, -40, -30, -30, -30, -30, -40, -50,
			-40, -20, 0, 0, 0, 0, -20, -40,
			-30, 0, 10, 15, 15, 10, 0, -30,
			-30, 5, 15, 20, 20, 15, 5, -30,
			-30, 0, 15, 20, 20, 15, 0, -30,
			-30, 5, 10, 15, 15, 10, 5, -30,
			-40, -20, 0, 5, 5, 0, -20, -40,
			-50, -40, -30, -30, -30, -30, -40, -50,
		},

		WHITE_BISHOP: {
			-20, -10, -10, -10, -10, -10, -10, -20,
			-10, 0, 0, 0, 0, 0, 0, -10,
			-10, 0, 5, 10, 10, 5, 0, -10,
			-10, 5, 5, 10, 10, 5, 5, -10,
			-10, 0, 10, 10, 10, 10, 0, -10,
			-10, 10, 10, 10, 10, 10, 10, -10,
			-10, 5, 0, 0, 0, 0, 5, -10,
			-20, -10, -10, -10, -10, -10, -10, -20,
		},

		WHITE_QUEEN: {
			-20, -10, -10, -5, -5, -10, -10, -20,
			-10, 0, 0, 0, 0, 0, 0, -10,
			-10, 0, 5, 5, 5, 5, 0, -10,
			-5, 0, 5, 5, 5, 5, 0, -5,
			0, 0, 5, 5, 5, 5, 0, -5,
			-10, 5, 5, 5, 5, 5, 0, -10,
			-10, 0, 5, 0, 0, 0, 0, -10,
			-20, -10, -10, -5, -5, -10, -10, -20,
		},

		WHITE_KING: {
			-30, -40, -40, -50, -50, -40, -40, -30,
			-30, -40, -40, -50, -50, -40, -40, -30,
			-30, -40, -40, -50, -50, -40, -40, -30,
			-30, -40, -40, -50, -50, -40, -40, -30,
			-20, -30, -30, -40, -40, -30, -30, -20,
			-10, -20, -20, -20, -20, -20, -20, -10,
			20, 20, 0, 0, 0, 0, 20, 20,
			20, 30, 10, 0, 0, 10, 30, 20,
		},
	}
)

func EvaluatePS(bitboard uint64, piece_square []int, side int) int {
	var (
		index, res int
	)

	for bitboard != 0 {
		index = 63 - bits.TrailingZeros64(bitboard) //bitboard and normal array has reverse index
		if side == BLACK {
			res += piece_square[FlipIndexVertical(index)]
		} else {
			res += piece_square[index]
		}
		ClearBit(bits.TrailingZeros64(bitboard), &bitboard)
	}

	return res
}

func (chess *Chess) CalculateBonus() int {
	/*
	 * Refer to the rule state in chessprograming wiki: https://www.chessprogramming.org/Material
	 * All the bonus/penalty point can be tune further
	 * This is based on White perspective, for Black we'll flip the board and reuse White logic
	 */
	if chess.SideToMove == BLACK {
		chess.Flip()
		defer chess.Flip()
	}

	var bonus = 0

	//Bonus for pair bishop
	if bits.OnesCount64(chess.Boards[WHITE_BISHOP]) >= 2 {
		bonus += 66
	}

	//Penalty for knight pair
	if bits.OnesCount64(chess.Boards[WHITE_KNIGHT]) >= 2 {
		bonus -= 64
	}

	//Penalty for rook pair
	if bits.OnesCount64(chess.Boards[WHITE_ROOK]) >= 2 {
		bonus -= 100
	}

	//Bonus for pair queen (encourage promotion to queen)
	if bits.OnesCount64(chess.Boards[WHITE_QUEEN]) >= 2 {
		bonus -= 180
	}

	//Penalty for not having any pawn left (harder for checkmate in endgame)
	if bits.OnesCount64(chess.Boards[WHITE_PAWN]) == 0 {
		bonus -= 300
	}

	return bonus
}

func (chess *Chess) Evaluate() int {
	//Calculate material value
	mat := 0

	for i := WHITE_PAWN; i <= WHITE_KING; i++ {
		mat += (bits.OnesCount64(chess.Boards[i]) - bits.OnesCount64(chess.Boards[i+6])) * material[i]
	}

	//Calculate piece square table
	ps := 0
	for i := WHITE_PAWN; i <= WHITE_KING; i++ {
		ps += EvaluatePS(chess.Boards[i], piece_square_table[i], WHITE)
		ps -= EvaluatePS(chess.Boards[i+6], piece_square_table[i], BLACK)
	}

	return mat + ps + chess.CalculateBonus()
}
