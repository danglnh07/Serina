package engine

import (
	"math/bits"
)

var (
	//Material value of each piece type (centipawn)
	pawnMat   = 100
	knightMat = 320
	bishopMat = 330
	rookMat   = 500
	queenMat  = 900
	kingMat   = 20000

	//Piece-square value (https://www.chessprogramming.org/Simplified_Evaluation_Function)
	//This is for white perspective
	pawnPS = []int{
		0, 0, 0, 0, 0, 0, 0, 0,
		50, 50, 50, 50, 50, 50, 50, 50,
		10, 10, 20, 30, 30, 20, 10, 10,
		5, 5, 10, 25, 25, 10, 5, 5,
		0, 0, 0, 20, 20, 0, 0, 0,
		5, -5, -10, 0, 0, -10, -5, 5,
		5, 10, 10, -20, -20, 10, 10, 5,
		0, 0, 0, 0, 0, 0, 0, 0,
	}

	knightPS = []int{
		-50, -40, -30, -30, -30, -30, -40, -50,
		-40, -20, 0, 0, 0, 0, -20, -40,
		-30, 0, 10, 15, 15, 10, 0, -30,
		-30, 5, 15, 20, 20, 15, 5, -30,
		-30, 0, 15, 20, 20, 15, 0, -30,
		-30, 5, 10, 15, 15, 10, 5, -30,
		-40, -20, 0, 5, 5, 0, -20, -40,
		-50, -40, -30, -30, -30, -30, -40, -50,
	}

	bishopPS = []int{
		-20, -10, -10, -10, -10, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 5, 5, 10, 10, 5, 5, -10,
		-10, 0, 10, 10, 10, 10, 0, -10,
		-10, 10, 10, 10, 10, 10, 10, -10,
		-10, 5, 0, 0, 0, 0, 5, -10,
		-20, -10, -10, -10, -10, -10, -10, -20,
	}

	rookPS = []int{
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 10, 10, 10, 10, 10, 5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		0, 0, 0, 5, 5, 0, 0, 0,
	}

	queenPS = []int{
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-5, 0, 5, 5, 5, 5, 0, -5,
		0, 0, 5, 5, 5, 5, 0, -5,
		-10, 5, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20,
	}

	kingPS = []int{
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-20, -30, -30, -40, -40, -30, -30, -20,
		-10, -20, -20, -20, -20, -20, -20, -10,
		20, 20, 0, 0, 0, 0, 20, 20,
		20, 30, 10, 0, 0, 10, 30, 20,
	}
)

func EvaluatePS(bitboard uint64, piece_square []int, side Color) int {
	var (
		index, res int
	)

	for bitboard != 0 {
		index = 63 - bits.TrailingZeros64(bitboard) //bitboard and normal array has reverse index
		if side == BLACK {
			//If this is BLACK, get the mirror value of the PST (flipping vertically)
			res += piece_square[8*(7-index/8)+index%8]
		} else {
			res += piece_square[index]
		}
		ClearBit(bits.TrailingZeros64(bitboard), &bitboard)
	}

	return res
}

func (chess *Chess) Evaluate() int {
	//Calculate material value
	material := 0

	material += (bits.OnesCount64(chess.WhitePawns) - bits.OnesCount64(chess.BlackPawns)) * pawnMat
	material += (bits.OnesCount64(chess.WhiteRooks) - bits.OnesCount64(chess.BlackRooks)) * rookMat
	material += (bits.OnesCount64(chess.WhiteKnights) - bits.OnesCount64(chess.BlackKnights)) * knightMat
	material += (bits.OnesCount64(chess.WhiteBishops) - bits.OnesCount64(chess.BlackBishops)) * bishopMat
	material += (bits.OnesCount64(chess.WhiteQueens) - bits.OnesCount64(chess.BlackQueens)) * queenMat
	material += (bits.OnesCount64(chess.WhiteKing) - bits.OnesCount64(chess.BlackKing)) * kingMat

	//Calculate piece square table
	var (
		temp         uint64
		piece_square int
	)

	temp = chess.WhitePawns
	piece_square += EvaluatePS(temp, pawnPS, WHITE)
	temp = chess.BlackPawns
	piece_square += EvaluatePS(temp, pawnPS, BLACK)

	temp = chess.WhiteRooks
	piece_square += EvaluatePS(temp, rookPS, WHITE)
	temp = chess.BlackRooks
	piece_square += EvaluatePS(temp, rookPS, BLACK)

	temp = chess.WhiteKnights
	piece_square += EvaluatePS(temp, knightPS, WHITE)
	temp = chess.BlackKnights
	piece_square += EvaluatePS(temp, knightPS, BLACK)

	temp = chess.WhiteBishops
	piece_square += EvaluatePS(temp, bishopPS, WHITE)
	temp = chess.BlackBishops
	piece_square += EvaluatePS(temp, bishopPS, BLACK)

	temp = chess.WhiteQueens
	piece_square += EvaluatePS(temp, queenPS, WHITE)
	temp = chess.BlackQueens
	piece_square += EvaluatePS(temp, queenPS, BLACK)

	temp = chess.WhiteKing
	piece_square += EvaluatePS(temp, kingPS, WHITE)
	temp = chess.BlackKing
	piece_square += EvaluatePS(temp, kingPS, BLACK)

	return material + piece_square
	// return material
	// return piece_square
}
