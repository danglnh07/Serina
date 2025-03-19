package engine

import "math/bits"

// Calculate sliding attack in horizontal and vertical direction
func HAndVMoves(index int, chess *Chess) uint64 {
	var (
		r                              uint64 = 0x1 << index
		o                              uint64 = chess.GenerateAllWhites() | chess.GenerateAllBlacks()
		m_H, m_V, horizontal, vertical uint64
	)
	m_H = RANK_MASK[index/8]
	m_V = FILE_MASK[7-index%8]
	horizontal = (((o & m_H) - 2*r) ^ bits.Reverse64(bits.Reverse64(o&m_H)-2*bits.Reverse64(r))) & m_H
	vertical = (((o & m_V) - 2*r) ^ bits.Reverse64(bits.Reverse64(o&m_V)-2*bits.Reverse64(r))) & m_V
	return horizontal | vertical
}

// Calculate sliding attack in diagonal (y = x) and anti-diagonal (y = -x) direction
func DAndAntiDMoves(index int, chess *Chess) uint64 {
	var (
		r                                  uint64 = 0x1 << index
		o                                  uint64 = chess.GenerateAllWhites() | chess.GenerateAllBlacks()
		m_D, m_AD, diagonal, anti_diagonal uint64
	)
	m_D = DIAGONAL_MASK[14-(index/8+index%8)]
	m_AD = ANTI_DIAGONAL_MASK[7-(index/8-index%8)]
	diagonal = (((o & m_D) - 2*r) ^ bits.Reverse64(bits.Reverse64(o&m_D)-2*bits.Reverse64(r))) & m_D
	anti_diagonal = (((o & m_AD) - 2*r) ^ bits.Reverse64(bits.Reverse64(o&m_AD)-2*bits.Reverse64(r))) & m_AD
	return diagonal | anti_diagonal
}

/*
 * Pin moves: for pin pieces, only Rooks, Bishops, Queens and Pawns can move (Knight can never move in any cases)
 * For Sliding pieces (SP), the logic is pretty much the same, so we group them into a single method
 * Pawns can move differently in different directions, so we handle them separately compared to sliding pin pieces
 * Pawns will NEVER be able to move if the pin ray is HORIZONTAL
 */

// Get all Sliding pin piece moves for every direction
func CalculateSPPinMoves(min, max, pseudoAttackerIndex, pinPieceIndex int, direction Direction) []string {
	var (
		moves []string
		move  = FromIndexToAlgebraic(pinPieceIndex)
	)

	//We go through all squares between (exclusively) min and max
	for i := min + int(direction); i < max; i += int(direction) {
		//Ignore the square the pin piece is currently occupied
		if i == pinPieceIndex {
			continue
		}
		moves = append(moves, move+FromIndexToAlgebraic(i))
	}
	//Unlike pawn that need condition, SP will always be able to capture the pseudo attacker
	moves = append(moves, move+FromIndexToAlgebraic(pseudoAttackerIndex))

	return moves
}

// Get all white pawn pin piece moves for FILE, DIAGONAL and ANTI_DIAGONAL direction
func CalculateWhitePawnPinMoves(pseudoAttackerIndex, pinPieceIndex int, direction Direction, empty uint64) []string {
	var (
		moves     []string
		move      = FromIndexToAlgebraic(pinPieceIndex)
		pawnMoves uint64
		attacker  uint64 = 0x1 << pseudoAttackerIndex
		index     int
	)

	//Handle FILE direction
	if direction == FILE {
		pawnMoves = (0x1 << (pinPieceIndex + 8)) & empty
		pawnMoves |= (0x1 << (pinPieceIndex + 16)) & empty & (empty << 8) & RANK_MASK[3]
		for pawnMoves != 0 {
			index = bits.TrailingZeros64(pawnMoves)
			moves = append(moves, move+FromIndexToAlgebraic(index))
			ClearBit(index, &pawnMoves)
		}
	} else if direction == DIAGONAL || direction == ANTI_DIAGONAL {
		//Logic for DIAGONAL and ANTI_DIAGONAL is the same, so we group them together
		pawnMoves = (0x1 << (pinPieceIndex + int(direction))) & attacker
		if pawnMoves != 0 {
			move += FromIndexToAlgebraic(bits.TrailingZeros64(pawnMoves))
			if 56 <= pseudoAttackerIndex && pseudoAttackerIndex <= 63 {
				moves = append(moves, []string{move + "Q", move + "R", move + "B", move + "N"}...)
			} else {
				moves = append(moves, move)
			}
		}
	}

	return moves
}

// Get all black pawn pin piece moves for FILE, DIAGONAL and ANTI_DIAGONAL direction
func CalculateBlackPawnPinMoves(pseudoAttackerIndex, pinPieceIndex int, direction Direction, empty uint64) []string {
	var (
		moves     []string
		move      = FromIndexToAlgebraic(pinPieceIndex)
		pawnMoves uint64
		attacker  uint64 = 0x1 << pseudoAttackerIndex
		index     int
	)

	//Handle FILE direction
	if direction == FILE {
		pawnMoves = (0x1 << (pinPieceIndex - 8)) & empty
		pawnMoves |= (0x1 << (pinPieceIndex - 16)) & empty & (empty >> 8) & RANK_MASK[4]
		for pawnMoves != 0 {
			index = bits.TrailingZeros64(pawnMoves)
			moves = append(moves, move+FromIndexToAlgebraic(index))
			ClearBit(index, &pawnMoves)
		}
	} else if direction == DIAGONAL || direction == ANTI_DIAGONAL {
		//Logic for DIAGONAL and ANTI_DIAGONAL is the same, so we group them together
		pawnMoves = (0x1 << (pinPieceIndex - int(direction))) & attacker
		if pawnMoves != 0 {
			move += FromIndexToAlgebraic(bits.TrailingZeros64(pawnMoves))
			if 0 <= pseudoAttackerIndex && pseudoAttackerIndex <= 7 {
				moves = append(moves, []string{move + "q", move + "r", move + "b", move + "n"}...)
			} else {
				moves = append(moves, move)
			}
		}
	}

	return moves
}

// Move generation, which will decide based on the side to move and call the appropiate method
func (chess *Chess) MoveGeneration() []string {
	// fmt.Println(chess.sideToMove)
	if chess.SideToMove {
		// utility.PrintList(chess.WhiteMoveGeneration())
		return chess.WhiteMoveGeneration()
	}
	return chess.BlackMoveGeneration()
}
