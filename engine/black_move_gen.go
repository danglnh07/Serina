package engine

import (
	"math/bits"
)

// Method that calculate all black pieces bitboard
func (chess *Chess) GenerateAllBlacks() uint64 {
	return chess.BlackPawns | chess.BlackRooks | chess.BlackKnights | chess.BlackBishops | chess.BlackQueens | chess.BlackKing
}

/*
 * Danger squares bitboard: the bitboard of all squares that is dangerous for the King to move to at current state
 * Danger squares doesn't mean the opponent can ACTUALLY attack to those squares
 * One example is guarding, the opponent piece protecting another opponent's piece
 * NOTE: for danger squares calculation, we have to temporary removed our King, because there are cases like
 * the King moving further away from sliding attacker, but they still in the ray line attack
 */

// Method that calculate danger squares bitboard of Black King
func (chess *Chess) GenerateBlackKingInDanger() uint64 {
	chess.Flip()
	defer chess.Flip()
	return FlipVertical(chess.GenerateWhiteKingInDanger())
}

/*
 * Attack bitboard: all squares that side can attack to. Since this is attack to, pawn advance and double push
 * do not count
 * This bitboard can be used to determine if the King is being checked
 */

// Calculate Black attack bitboard
func (chess *Chess) GenerateBlackAttacks() uint64 {
	chess.Flip()
	defer chess.Flip()
	return FlipVertical(chess.GenerateWhiteAttacks())
}

// Check if the Black King is under attacked (is checked)
func (chess *Chess) IsBlackKingChecked() bool {
	return chess.BlackKing&chess.GenerateWhiteAttacks() != 0
}

/*
 * Attacker bitboard: all current attackers to the King
 * The idea is simple: we assume that the King is an 'omni-piece' that can perform every attack move, and check if for each attack move type,
 * does its capture corresponding to the correct piece type
 */

// Calculate all Black King attackers bitboard, and also if there is at least one sliding piece attacker
func (chess *Chess) CalculateBlackKingAttackers() (uint64, bool) {
	chess.Flip()
	defer chess.Flip()
	attackers, hasSPAttackers := chess.CalculateWhiteKingAttackers()
	return FlipVertical(attackers), hasSPAttackers
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

// BLACK move generation
func (chess *Chess) BlackMoveGeneration() []string {
	//Variables declaration
	var (
		moves     []string
		move      string
		kingIndex = bits.TrailingZeros64(chess.BlackKing)

		//Temporary bitboard (since pin pieces can only moving along the line, we remove them from the bitboard)
		bp, br, bn, bb, bq = chess.BlackPawns, chess.BlackRooks, chess.BlackKnights, chess.BlackBishops, chess.BlackQueens

		//All whites, blacks, empty bitboard
		whites, blacks = chess.GenerateAllWhites(), chess.GenerateAllBlacks()
		empty          = ^(whites | blacks)

		//Index used for looping through the move bitboard
		index     int
		direction Direction

		//RANK, FILE constant
		RANK_5, RANK_7, FILE_A, FILE_H = RANK_MASK[4], RANK_MASK[6], FILE_MASK[0], FILE_MASK[7]
	)

	//Calculate King's moves/evasions
	kingMove := ^chess.GenerateBlackKingInDanger() & KING_ATTACK[kingIndex] & ^blacks
	move = FromIndexToAlgebraic(kingIndex)
	for kingMove != 0 {
		index = bits.TrailingZeros64(kingMove)
		moves = append(moves, move+FromIndexToAlgebraic(index))
		ClearBit(index, &kingMove)
	}

	//Get King's attackers
	attackers, hasSPAttacker := chess.CalculateBlackKingAttackers()
	//If this is double check, or there is only the King left, then we stop here
	if bits.OnesCount64(attackers) > 1 || bp|br|bn|bb|bq == 0 {
		return moves
	}

	//Handling en passant
	if 16 <= chess.EnPassantTarget && chess.EnPassantTarget <= 23 {
		wpTemp, bpTemp := chess.WhitePawns, chess.BlackPawns
		epMove := ((0x1 << chess.EnPassantTarget) << 9) & bp & ^FILE_H
		epMove |= ((0x1 << chess.EnPassantTarget) << 7) & bp & ^FILE_A

		//Loop though all ep move found
		for epMove != 0 {
			//Perform the en passant move
			index = bits.TrailingZeros64(epMove)
			ClearBit(index, &chess.BlackPawns)
			SetBit(chess.EnPassantTarget, &chess.BlackPawns)
			ClearBit(chess.EnPassantTarget+8, &chess.WhitePawns)

			//If the en passant move not lead to a check, add them to the list
			if !chess.IsBlackKingChecked() {
				moves = append(moves, FromIndexToAlgebraic(index)+FromIndexToAlgebraic(chess.EnPassantTarget)+"EP")
			}

			//Restore the original state of the board after making a move
			chess.WhitePawns = wpTemp
			chess.BlackPawns = bpTemp

			//Clear the ep move
			ClearBit(index, &epMove)
		}
	}

	//Calculate pin pieces' moves (whether it's single check or no check, we still need to remove the pin pieces first)
	var (
		temp, rayline                             uint64
		min, max, pinAttackerIndex, pinPieceIndex int
		pinMoves                                  []string
	)

	temp = chess.WhiteRooks | chess.WhiteQueens
	for temp != 0 {
		pinAttackerIndex = bits.TrailingZeros64(temp)
		min = Min(kingIndex, pinAttackerIndex)
		max = Max(kingIndex, pinAttackerIndex)

		if IsAtSameRank(kingIndex, pinAttackerIndex) || IsAtSameFile(kingIndex, pinAttackerIndex) {
			//Get the direction
			if IsAtSameRank(kingIndex, pinAttackerIndex) {
				direction = RANK
			} else {
				direction = FILE
			}

			rayline = CalculateRayAttackLine(min, max, direction)
			if bits.OnesCount64(rayline&whites) == 0 && bits.OnesCount64(rayline&blacks) == 1 {
				pinPieceIndex = bits.TrailingZeros64(rayline & blacks)
				//For Rook pin attacker, only Rook and Queen can move. For FILE specifically, pawn also can move
				if IsPieceAtIndex(br, pinPieceIndex) || IsPieceAtIndex(bq, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculateSPPinMoves(min, max, pinAttackerIndex, pinPieceIndex, direction)...)
				}

				if direction == FILE && IsPieceAtIndex(bp, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculatePawnPinMoves(pinAttackerIndex, pinPieceIndex, direction, empty, BLACK)...)
				}

				ClearBitAcrossBoards(pinPieceIndex, &bp, &br, &bn, &bb, &bq)
			}
		}

		ClearBit(pinAttackerIndex, &temp)
	}

	temp = chess.WhiteBishops | chess.WhiteQueens
	for temp != 0 {
		pinAttackerIndex = bits.TrailingZeros64(temp)
		min = Min(kingIndex, pinAttackerIndex)
		max = Max(kingIndex, pinAttackerIndex)

		if IsAtSameDiagonal(kingIndex, pinAttackerIndex) || IsAtSameAntiDiagonal(kingIndex, pinAttackerIndex) {
			//Calculate direction
			if IsAtSameDiagonal(kingIndex, pinAttackerIndex) {
				direction = DIAGONAL
			} else {
				direction = ANTI_DIAGONAL
			}

			rayline = CalculateRayAttackLine(min, max, direction)
			if bits.OnesCount64(rayline&whites) == 0 && bits.OnesCount64(rayline&blacks) == 1 {
				pinPieceIndex = bits.TrailingZeros64(rayline & blacks)
				//For DIAGONAL and ANTI_DIAGONAL, only Bishop, Queen and Pawn (capture) can move
				if IsPieceAtIndex(bb, pinPieceIndex) || IsPieceAtIndex(bq, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculateSPPinMoves(min, max, pinAttackerIndex, pinPieceIndex, direction)...)
				} else if IsPieceAtIndex(bp, pinPieceIndex) {
					pinMoves = append(pinMoves, CalculatePawnPinMoves(pinAttackerIndex, pinPieceIndex, direction, empty, BLACK)...)
				}
				ClearBitAcrossBoards(pinPieceIndex, &bp, &br, &bn, &bb, &bq)
			}
		}

		ClearBit(pinAttackerIndex, &temp)
	}

	//Handling single check
	if bits.OnesCount64(attackers) == 1 {
		attackerIndex := bits.TrailingZeros64(attackers)
		move = FromIndexToAlgebraic(attackerIndex)

		/*---Calculate attacker capturing moves---*/
		capture := (attackers << 7) & ^FILE_A & bp
		capture |= (attackers << 9) & ^FILE_H & bp
		if 0 <= attackerIndex && attackerIndex <= 7 {
			var pawnPromotion string
			for capture != 0 {
				index = bits.TrailingZeros64(capture)
				pawnPromotion = FromIndexToAlgebraic(index) + move
				moves = append(moves, []string{pawnPromotion + "q", pawnPromotion + "r", pawnPromotion + "b", pawnPromotion + "n"}...)
				ClearBit(index, &capture)
			}
		}
		capture |= HAndVMoves(attackerIndex, chess) & br
		capture |= KNIGHT_ATTACK[attackerIndex] & bn
		capture |= DAndAntiDMoves(attackerIndex, chess) & bb
		capture |= (HAndVMoves(attackerIndex, chess) | DAndAntiDMoves(attackerIndex, chess)) & bq
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
					blockMoves |= ((0x1 << i) << 8) & bp
					blockMoves |= ((0x1 << i) << 16) & (empty << 8) & bp & RANK_7
					for blockMoves != 0 {
						index = bits.TrailingZeros64(blockMoves)
						if 0 <= i && i <= 7 {
							promotion := FromIndexToAlgebraic(index)
							moves = append(moves, []string{promotion + "q", promotion + "r", promotion + "b", promotion + "n"}...)
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
					blockMoves |= ((0x1 << i) << 8) & bp
					blockMoves |= ((0x1 << i) << 16) & (empty << 8) & bp & RANK_7
				}

				blockMoves |= HAndVMoves(i, chess) & br
				blockMoves |= KNIGHT_ATTACK[i] & bn
				blockMoves |= DAndAntiDMoves(i, chess) & bb
				blockMoves |= (HAndVMoves(i, chess) | DAndAntiDMoves(i, chess)) & bq
				for blockMoves != 0 {
					index = bits.TrailingZeros64(blockMoves)
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

	pawnMoves = (bp >> 8) & empty
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move = FromIndexToAlgebraic(index+8) + FromIndexToAlgebraic(index)
		//If the destination stay at RANK_1, it's pawn promotion
		if 0 <= index && index <= 7 {
			moves = append(moves, []string{move + "q", move + "r", move + "b", move + "n"}...)
		} else {
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (bp >> 16) & empty & (empty >> 8) & RANK_5
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		moves = append(moves, FromIndexToAlgebraic(index+16)+FromIndexToAlgebraic(index))
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (bp >> 7) & whites & ^FILE_H
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move = FromIndexToAlgebraic(index+7) + FromIndexToAlgebraic(index)
		//If the destination stay at RANK_1, it's pawn promotion
		if 0 <= index && index <= 7 {
			moves = append(moves, []string{move + "q", move + "r", move + "b", move + "n"}...)
		} else {
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (bp >> 9) & whites & ^FILE_A
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move = FromIndexToAlgebraic(index+9) + FromIndexToAlgebraic(index)
		//If the destination stay at RANK_1, it's pawn promotion
		if 0 <= index && index <= 7 {
			moves = append(moves, []string{move + "q", move + "r", move + "b", move + "n"}...)
		} else {
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}

	temp = br | bq
	for temp != 0 {
		pieceIndex = bits.TrailingZeros64(temp)
		rookMoves = HAndVMoves(pieceIndex, chess) & ^blacks
		move = FromIndexToAlgebraic(pieceIndex)
		for rookMoves != 0 {
			index = bits.TrailingZeros64(rookMoves)
			moves = append(moves, move+FromIndexToAlgebraic(index))
			ClearBit(index, &rookMoves)
		}
		ClearBit(pieceIndex, &temp)
	}

	for bn != 0 {
		pieceIndex = bits.TrailingZeros64(bn)
		knightMoves = KNIGHT_ATTACK[pieceIndex] & ^blacks
		move = FromIndexToAlgebraic(pieceIndex)
		for knightMoves != 0 {
			index = bits.TrailingZeros64(knightMoves)
			moves = append(moves, move+FromIndexToAlgebraic(index))
			ClearBit(index, &knightMoves)
		}
		ClearBit(pieceIndex, &bn)
	}

	temp = bb | bq
	for temp != 0 {
		pieceIndex = bits.TrailingZeros64(temp)
		bishopMoves = DAndAntiDMoves(pieceIndex, chess) & ^blacks
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
	blackKingInDanger := chess.GenerateBlackKingInDanger()
	if (chess.CastlingPrivilege&int(BLACK_KING_SIDE)) == int(BLACK_KING_SIDE) && (empty&0x600000000000000) == 0x600000000000000 && (blackKingInDanger&0xE00000000000000) == 0 {
		moves = append(moves, "o-o")
	}

	if (chess.CastlingPrivilege&int(BLACK_QUEEN_SIDE)) == int(BLACK_QUEEN_SIDE) && (empty&0x7000000000000000) == 0x7000000000000000 && (blackKingInDanger&0x3800000000000000) == 0 {
		moves = append(moves, "o-o-o")
	}

	return moves
}
