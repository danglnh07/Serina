package engine

import (
	"math/bits"
)

// Method that calculate all white pieces bitboard
func (chess *Chess) GenerateAllWhites() uint64 {
	return chess.WhitePawns | chess.WhiteRooks | chess.WhiteKnights | chess.WhiteBishops | chess.WhiteQueens | chess.WhiteKing
}

/*
 * Danger squares bitboard: the bitboard of all squares that is dangerous for the King to move to at current state
 * Danger squares doesn't mean the opponent can ACTUALLY attack to those squares
 * One example is guarding, the opponent piece protecting another opponent's piece
 * NOTE: for danger squares calculation, we have to temporary removed our King, because there are cases like
 * the King moving further away from sliding attacker, but they still in the ray line attack
 */

// Method that calculate danger squares bitboard of White King
func (chess *Chess) GenerateWhiteKingInDanger() uint64 {
	var (
		blacks_empty_whiteKing  = ^chess.GenerateAllWhites() | chess.WhiteKing
		FILE_A, FILE_H          = FILE_MASK[0], FILE_MASK[7]
		temp, whiteInDanger, wk uint64
		index                   int
	)

	//Temporary remove the White King
	wk = chess.WhiteKing
	chess.WhiteKing = 0

	//Calculate for black pawns (en passant will never threaten a King -> ignored)
	whiteInDanger |= (chess.BlackPawns >> 9) & blacks_empty_whiteKing & ^FILE_A
	whiteInDanger |= (chess.BlackPawns >> 7) & blacks_empty_whiteKing & ^FILE_H

	//Calculate for black rooks and black queens (horizontal and vertical only)
	temp = chess.BlackRooks | chess.BlackQueens
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteInDanger |= HAndVMoves(index, chess) & blacks_empty_whiteKing
		ClearBit(index, &temp)
	}

	//Calculate for black knights
	temp = chess.BlackKnights
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteInDanger |= KNIGHT_ATTACK[index] & blacks_empty_whiteKing
		ClearBit(index, &temp)
	}

	//Calculate for black bishops and black queens (diagonal and anti diagonal only)
	temp = chess.BlackBishops | chess.BlackQueens
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteInDanger |= DAndAntiDMoves(index, chess) & blacks_empty_whiteKing
		ClearBit(index, &temp)
	}

	//Calculate for black king (there can be only 1 King -> no need for loop)
	index = bits.TrailingZeros64(chess.BlackKing)
	whiteInDanger |= KING_ATTACK[index] & blacks_empty_whiteKing

	//Restore the value of whiteKing
	chess.WhiteKing = wk
	return whiteInDanger
}

/*
 * Attack bitboard: all squares that side can attack to. Since this is attack to, pawn advance and double push
 * do not count
 * This bitboard can be used to determine if the King is being checked
 */

// Calculate White attack bitboard
func (chess *Chess) GenerateWhiteAttacks() uint64 {
	var (
		FILE_A, FILE_H = FILE_MASK[0], FILE_MASK[7]
		//whites cannot land to squares that is occupied by another whites, so we negate whites bitboard
		whitesCanLandTo, blacks = ^chess.GenerateAllWhites(), chess.GenerateAllBlacks()
		temp, whiteCanAttack    uint64
		index                   int
	)

	//Calculate white pawns attack squares
	whiteCanAttack |= (chess.WhitePawns << 9) & blacks & ^FILE_H
	whiteCanAttack |= (chess.WhitePawns << 7) & blacks & ^FILE_A
	if 47 >= chess.EnPassantTarget && chess.EnPassantTarget >= 40 {
		whiteCanAttack |= 0x1 << chess.EnPassantTarget
	}

	//Calculate white rooks and white queens (horizontal and vertical only) attack squares
	temp = chess.WhiteRooks | chess.WhiteQueens
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteCanAttack |= HAndVMoves(index, chess) & whitesCanLandTo
		ClearBit(index, &temp)
	}

	//Calculate white knights attack squares
	temp = chess.WhiteKnights
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteCanAttack |= KNIGHT_ATTACK[index] & whitesCanLandTo
		ClearBit(index, &temp)
	}

	//Calculate white bishops and white queens (diagonal and anti diagonal only) attack squares
	temp = chess.WhiteBishops | chess.WhiteQueens
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteCanAttack |= DAndAntiDMoves(index, chess) & whitesCanLandTo
		ClearBit(index, &temp)
	}

	//Calculate white king attack squares (since there is only one King, no need to use a loop)
	index = bits.TrailingZeros64(chess.WhiteKing)
	whiteCanAttack |= KING_ATTACK[index] & whitesCanLandTo

	return whiteCanAttack
}

// Check if the White King is under attacked (is checked)
func (chess *Chess) IsWhiteKingChecked() bool {
	return chess.WhiteKing&chess.GenerateBlackAttacks() != 0
}

/*
 * Attacker bitboard: all current attackers to the King
 * The idea is simple: we assume that the King is an 'omni-piece' that can perform every attack move, and check if for each attack move type,
 * does its capture corresponding to the correct piece type
 */

// Calculate all White King attackers bitboard, and return if there is at least one sliding piece attacker
func (chess *Chess) CalculateWhiteKingAttackers() (uint64, bool) {
	var (
		FILE_A, FILE_H = FILE_MASK[0], FILE_MASK[7]
		kingIndex      = bits.TrailingZeros64(chess.WhiteKing)
	)

	//Calculate attacker
	pawnAttackers := (chess.WhiteKing << 7) & ^FILE_A & chess.BlackPawns
	pawnAttackers |= (chess.WhiteKing << 9) & ^FILE_H & chess.BlackPawns
	rookAttackers := HAndVMoves(kingIndex, chess) & chess.BlackRooks
	knightAttackers := KNIGHT_ATTACK[kingIndex] & chess.BlackKnights
	bishopAttackers := DAndAntiDMoves(kingIndex, chess) & chess.BlackBishops
	queenAttackers := (HAndVMoves(kingIndex, chess) | DAndAntiDMoves(kingIndex, chess)) & chess.BlackQueens

	return pawnAttackers | rookAttackers | knightAttackers | bishopAttackers | queenAttackers, rookAttackers != 0 || bishopAttackers != 0 || queenAttackers != 0
}

/*
 * Move generation: there are 3 cases:
 * 1. Double check (more than 1 attacker): only King evasion is valid
 * 2. Single check (only 1 attacker): there are 3 different move:
 * 		King evasion
 * 		Capture attacker
 * 		Block attacker attack line (only possible is attacker is SP)
 * 3. No check: all pseudo legal moves of NON-PINNED pieces and pin pieces' moves are legal
 * NOTE:
 * 1. For capture/block attacker in single check, only non-pinned pieces can do this
 * 2. Castling has its own condition of King checking, so they will be handled differently.
 * 3. En passant, is the most tricky move in this case, but since the likelihood of en passant is rare, we can simply perform
 * en passant, check if the King is still being checked or not, then unmake that move
 */

// WHITE move generation
func (chess *Chess) WhiteMoveGeneration() []string {
	//Variables declaration
	var (
		moves     []string
		move      string
		kingIndex = bits.TrailingZeros64(chess.WhiteKing)

		//Temporary bitboard (since pin pieces can only moving along the line, we remove them from the bitboard)
		wp, wr, wn, wb, wq = chess.WhitePawns, chess.WhiteRooks, chess.WhiteKnights, chess.WhiteBishops, chess.WhiteQueens

		//All whites, blacks, empty bitboard
		whites, blacks = chess.GenerateAllWhites(), chess.GenerateAllBlacks()
		empty          = ^(whites | blacks)

		//Temporary index used for looping through bitboard and direction temporary variable
		index     int
		direction Direction

		//RANK, FILE constant
		RANK_2, RANK_4, FILE_A, FILE_H = RANK_MASK[1], RANK_MASK[3], FILE_MASK[0], FILE_MASK[7]
	)

	//Calculate King movement/evasion
	kingMove := ^chess.GenerateWhiteKingInDanger() & KING_ATTACK[kingIndex] & ^whites // = KING_ATTACK & !(kingInDanger | whites)
	move = FromIndexToAlgebraic(kingIndex)
	for kingMove != 0 {
		index = bits.TrailingZeros64(kingMove)
		moves = append(moves, move+FromIndexToAlgebraic(index))
		ClearBit(index, &kingMove)
	}

	//Get King's attackers
	attackers, hasSPAttacker := chess.CalculateWhiteKingAttackers()
	//If this is double check, or there is only the King left, then we stop here
	if bits.OnesCount64(attackers) > 1 || wp|wr|wn|wb|wq == 0 {
		return moves
	}

	//Handling en passant
	if 40 <= chess.EnPassantTarget && chess.EnPassantTarget <= 47 {
		wpTemp, bpTemp := chess.WhitePawns, chess.BlackPawns
		epMove := ((0x1 << chess.EnPassantTarget) >> 9) & wp & ^FILE_A
		epMove |= ((0x1 << chess.EnPassantTarget) >> 7) & wp & ^FILE_H

		//Loop though all ep move found
		for epMove != 0 {
			//Perform the en passant move
			index = bits.TrailingZeros64(epMove)
			ClearBit(index, &chess.WhitePawns)
			SetBit(chess.EnPassantTarget, &chess.WhitePawns)
			ClearBit(chess.EnPassantTarget-8, &chess.BlackPawns)

			//If the en passant move not lead to a check, add them to the list
			if !chess.IsWhiteKingChecked() {
				moves = append(moves, FromIndexToAlgebraic(index)+FromIndexToAlgebraic(chess.EnPassantTarget)+"EP")
			}

			//Restore the board after making a move
			chess.WhitePawns = wpTemp
			chess.BlackPawns = bpTemp

			//Clear the ep move
			ClearBit(index, &epMove)
		}

	}

	//Calculate pin pieces' moves (whether it's single check or no check, we still need to remove the pin pieces beforehand)
	var (
		temp, rayline                                uint64
		min, max, pseudoAttackerIndex, pinPieceIndex int
		pinMoves                                     []string
	)

	temp = chess.BlackRooks | chess.BlackQueens
	for temp != 0 {
		pseudoAttackerIndex = bits.TrailingZeros64(temp)
		min = Min(kingIndex, pseudoAttackerIndex)
		max = Max(kingIndex, pseudoAttackerIndex)

		if IsAtSameRank(kingIndex, pseudoAttackerIndex) || IsAtSameFile(kingIndex, pseudoAttackerIndex) {
			//Get the direction
			if IsAtSameRank(kingIndex, pseudoAttackerIndex) {
				direction = RANK
			} else {
				direction = FILE
			}

			rayline = CalculateRayAttackLine(min, max, direction)
			if bits.OnesCount64(rayline&blacks) == 0 && bits.OnesCount64(rayline&whites) == 1 {
				pinPieceIndex = bits.TrailingZeros64(rayline & whites)
				//For Rook pin attacker, only Rook and Queen can move. For FILE specifically, pawn also can move
				if IsPieceAtIndex(wr, pinPieceIndex) || IsPieceAtIndex(wq, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculateSPPinMoves(min, max, pseudoAttackerIndex, pinPieceIndex, direction)...)
				}

				if direction == FILE && IsPieceAtIndex(wp, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculatePawnPinMoves(pseudoAttackerIndex, pinPieceIndex, direction, empty, WHITE)...)
				}

				ClearBitAcrossBoards(pinPieceIndex, &wp, &wr, &wn, &wb, &wq)
			}
		}

		ClearBit(pseudoAttackerIndex, &temp)
	}

	temp = chess.BlackBishops | chess.BlackQueens
	for temp != 0 {
		pseudoAttackerIndex = bits.TrailingZeros64(temp)
		min = Min(kingIndex, pseudoAttackerIndex)
		max = Max(kingIndex, pseudoAttackerIndex)

		if IsAtSameDiagonal(kingIndex, pseudoAttackerIndex) || IsAtSameAntiDiagonal(kingIndex, pseudoAttackerIndex) {
			//Calculate direction
			if IsAtSameDiagonal(kingIndex, pseudoAttackerIndex) {
				direction = DIAGONAL
			} else {
				direction = ANTI_DIAGONAL
			}

			rayline = CalculateRayAttackLine(min, max, direction)
			if bits.OnesCount64(rayline&blacks) == 0 && bits.OnesCount64(rayline&whites) == 1 {
				pinPieceIndex = bits.TrailingZeros64(rayline & whites)
				//For DIAGONAL and ANTI_DIAGONAL, only Bishop, Queen and Pawn (capture) can move
				if IsPieceAtIndex(wb, pinPieceIndex) || IsPieceAtIndex(wq, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculateSPPinMoves(min, max, pseudoAttackerIndex, pinPieceIndex, direction)...)
				} else if IsPieceAtIndex(wp, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculatePawnPinMoves(pseudoAttackerIndex, pinPieceIndex, direction, empty, WHITE)...)
				}
				ClearBitAcrossBoards(pinPieceIndex, &wp, &wr, &wn, &wb, &wq)
			}
		}

		ClearBit(pseudoAttackerIndex, &temp)
	}

	//Handling single check
	if bits.OnesCount64(attackers) == 1 {
		attackerIndex := bits.TrailingZeros64(attackers)
		move = FromIndexToAlgebraic(attackerIndex)

		/*---Calculate attacker capturing moves---*/
		capture := (attackers >> 7) & ^FILE_H & wp
		capture |= (attackers >> 9) & ^FILE_A & wp
		if 56 <= attackerIndex && attackerIndex <= 63 {
			var pawnPromotion string
			for capture != 0 {
				index = bits.TrailingZeros64(capture)
				pawnPromotion = FromIndexToAlgebraic(index) + move
				moves = append(moves, []string{pawnPromotion + "Q", pawnPromotion + "R", pawnPromotion + "B", pawnPromotion + "N"}...)
				ClearBit(index, &capture)
			}
		}
		capture |= HAndVMoves(attackerIndex, chess) & wr
		capture |= KNIGHT_ATTACK[attackerIndex] & wn
		capture |= DAndAntiDMoves(attackerIndex, chess) & wb
		capture |= (HAndVMoves(attackerIndex, chess) | DAndAntiDMoves(attackerIndex, chess)) & wq
		for capture != 0 {
			index = bits.TrailingZeros64(capture)
			moves = append(moves, FromIndexToAlgebraic(index)+move)
			ClearBit(index, &capture)
		}

		/*---Calculate attacker blocking moves---*/
		if hasSPAttacker {
			var blockMoves uint64
			min, max = Min(attackerIndex, kingIndex), Max(attackerIndex, kingIndex)

			switch {
			case IsAtSameRank(attackerIndex, kingIndex):
				direction = RANK
				//Handle pawn advance
				for i := min + int(RANK); i < max; i += int(RANK) {
					move = FromIndexToAlgebraic(i)
					blockMoves |= ((0x1 << i) >> 8) & wp
					blockMoves |= ((0x1 << i) >> 16) & (empty >> 8) & wp & RANK_2
					for blockMoves != 0 {
						index = bits.TrailingZeros64(blockMoves)
						if 56 <= i && i <= 63 {
							promotion := FromIndexToAlgebraic(index) + move
							moves = append(moves, []string{promotion + "Q", promotion + "R", promotion + "B", promotion + "N"}...)
						} else {
							moves = append(moves, FromIndexToAlgebraic(index)+move)
						}
						ClearBit(index, &blockMoves)
					}
				}
			case IsAtSameFile(attackerIndex, kingIndex):
				direction = FILE
			case IsAtSameDiagonal(attackerIndex, kingIndex):
				direction = DIAGONAL
			case IsAtSameAntiDiagonal(attackerIndex, kingIndex):
				direction = ANTI_DIAGONAL
			}

			for i := min + int(direction); i < max; i += int(direction) {
				if direction == DIAGONAL || direction == ANTI_DIAGONAL {
					blockMoves |= ((0x1 << i) >> 8) & wp
					blockMoves |= ((0x1 << i) >> 16) & (empty >> 8) & wp & RANK_2
				}

				blockMoves |= HAndVMoves(i, chess) & wr
				blockMoves |= KNIGHT_ATTACK[i] & wn
				blockMoves |= DAndAntiDMoves(i, chess) & wb
				blockMoves |= (HAndVMoves(i, chess) | DAndAntiDMoves(i, chess)) & wq
				for blockMoves != 0 {
					index = bits.TrailingZeros64(blockMoves)
					//moves = append(moves, FromIndexToAlgebraic(index)+move)
					moves = append(moves, FromIndexToAlgebraic(index)+FromIndexToAlgebraic(i))
					ClearBit(index, &blockMoves)
				}
			}
		}

		return moves
	}

	//No check
	var (
		pawnMoves, rookMoves, knightMoves, bishopMoves uint64
		pieceIndex                                     int
	)

	pawnMoves = (wp << 8) & empty
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move = FromIndexToAlgebraic(index-8) + FromIndexToAlgebraic(index)
		//If the destination stay at RANK_8, it's pawn promotion
		if 56 <= index && index <= 63 {
			moves = append(moves, []string{move + "Q", move + "R", move + "B", move + "N"}...)
		} else {
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (wp << 16) & empty & (empty << 8) & RANK_4
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		moves = append(moves, FromIndexToAlgebraic(index-16)+FromIndexToAlgebraic(index))
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (wp << 7) & blacks & ^FILE_A
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move = FromIndexToAlgebraic(index-7) + FromIndexToAlgebraic(index)
		//If the destination stay at RANK_8, it's pawn promotion
		if 56 <= index && index <= 63 {
			moves = append(moves, []string{move + "Q", move + "R", move + "B", move + "N"}...)
		} else {
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (wp << 9) & blacks & ^FILE_H
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move = FromIndexToAlgebraic(index-9) + FromIndexToAlgebraic(index)
		//If the destination stay at RANK_8, it's pawn promotion
		if 56 <= index && index <= 63 {
			moves = append(moves, []string{move + "Q", move + "R", move + "B", move + "N"}...)
		} else {
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}

	temp = wr | wq
	for temp != 0 {
		pieceIndex = bits.TrailingZeros64(temp)
		rookMoves = HAndVMoves(pieceIndex, chess) & ^whites
		move = FromIndexToAlgebraic(pieceIndex)
		for rookMoves != 0 {
			index = bits.TrailingZeros64(rookMoves)
			moves = append(moves, move+FromIndexToAlgebraic(index))
			ClearBit(index, &rookMoves)
		}
		ClearBit(pieceIndex, &temp)
	}

	for wn != 0 {
		pieceIndex = bits.TrailingZeros64(wn)
		knightMoves = KNIGHT_ATTACK[pieceIndex] & ^whites
		move = FromIndexToAlgebraic(pieceIndex)
		for knightMoves != 0 {
			index = bits.TrailingZeros64(knightMoves)
			moves = append(moves, move+FromIndexToAlgebraic(index))
			ClearBit(index, &knightMoves)
		}
		ClearBit(pieceIndex, &wn)
	}

	temp = wb | wq
	for temp != 0 {
		pieceIndex = bits.TrailingZeros64(temp)
		bishopMoves = DAndAntiDMoves(pieceIndex, chess) & ^whites
		move = FromIndexToAlgebraic(pieceIndex)
		for bishopMoves != 0 {
			index = bits.TrailingZeros64(bishopMoves)
			moves = append(moves, move+FromIndexToAlgebraic(index))
			ClearBit(index, &bishopMoves)
		}
		ClearBit(pieceIndex, &temp)
	}

	//Append pin moves into list of moves
	moves = append(moves, pinMoves...)

	//Handling castling
	whiteKingInDanger := chess.GenerateWhiteKingInDanger()
	if (chess.CastlingPrivilege&int(WHITE_KING_SIDE)) == int(WHITE_KING_SIDE) && (empty&0x6) == 0x6 && (whiteKingInDanger&0xE) == 0 {
		moves = append(moves, "O-O")
	}

	if (chess.CastlingPrivilege&int(WHITE_QUEEN_SIDE)) == int(WHITE_QUEEN_SIDE) && (empty&0x70) == 0x70 && (whiteKingInDanger&0x38) == 0 {
		moves = append(moves, "O-O-O")
	}

	return moves
}
