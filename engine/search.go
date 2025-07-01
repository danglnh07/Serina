package engine

import (
	"math"
)

func (chess *Chess) Search(depth, alpha, beta int) (int, Move) {
	if depth == 0 {
		return chess.Evaluate(), Move{} // Return evaluation and no move at leaf nodes
	}

	moves := chess.MoveGeneration()

	// Check for game end (checkmate or stalemate)
	if len(moves) == 0 {
		if chess.IsBlackKingChecked() || chess.IsWhiteKingChecked() {
			// Checkmate: Large negative score (loss for side to move)
			return -math.MaxInt32, Move{}
		}
		return 0, Move{} // Stalemate
	}

	// Perform minimax with alpha-beta pruning (fail-soft)
	bestScore := -math.MaxInt32
	var bestMove Move
	for _, move := range moves {
		clone := chess.Clone()
		clone.MakeMove(move)
		// Recursive search with negated alpha/beta
		eval, _ := clone.Search(depth-1, -beta, -alpha)
		eval = -eval // Negate for negamax
		if eval > bestScore {
			bestScore = eval
			bestMove = move
			if eval > alpha {
				alpha = eval // Update alpha only when a new best move is found
			}
		}
		if eval >= beta {
			return bestScore, bestMove // Fail-soft beta cutoff
		}
	}

	return bestScore, bestMove
}
