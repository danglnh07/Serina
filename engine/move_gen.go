package engine

import (
	"math/bits"
)

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

func (chess *Chess) GenerateAllWhites() uint64 {
	var res uint64
	for i := WHITE_PAWN; i <= WHITE_KING; i++ {
		res |= chess.Boards[i]
	}
	return res
}

func (chess *Chess) GenerateAllBlacks() uint64 {
	var res uint64
	for i := BLACK_PAWN; i <= BLACK_KING; i++ {
		res |= chess.Boards[i]
	}
	return res
}

func (chess *Chess) GenerateWhiteKingInDanger() uint64 {
	var (
		blacks_empty_whiteKing  = ^chess.GenerateAllWhites() | chess.Boards[WHITE_KING]
		FILE_A, FILE_H          = FILE_MASK[0], FILE_MASK[7]
		temp, whiteInDanger, wk uint64
		index                   int
	)

	//Temporary remove the White King
	wk = chess.Boards[WHITE_KING]
	chess.Boards[WHITE_KING] = 0

	//Calculate for black pawns (en passant will never threaten a King -> ignored)
	whiteInDanger |= (chess.Boards[BLACK_PAWN] >> 9) & blacks_empty_whiteKing & ^FILE_A
	whiteInDanger |= (chess.Boards[BLACK_PAWN] >> 7) & blacks_empty_whiteKing & ^FILE_H

	//Calculate for black rooks and black queens (horizontal and vertical only)
	temp = chess.Boards[BLACK_ROOK] | chess.Boards[BLACK_QUEEN]
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteInDanger |= HAndVMoves(index, chess) & blacks_empty_whiteKing
		ClearBit(index, &temp)
	}

	//Calculate for black knights
	temp = chess.Boards[BLACK_KNIGHT]
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteInDanger |= KNIGHT_ATTACK[index] & blacks_empty_whiteKing
		ClearBit(index, &temp)
	}

	//Calculate for black bishops and black queens (diagonal and anti diagonal only)
	temp = chess.Boards[BLACK_BISHOP] | chess.Boards[BLACK_QUEEN]
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteInDanger |= DAndAntiDMoves(index, chess) & blacks_empty_whiteKing
		ClearBit(index, &temp)
	}

	//Calculate for black king (there can be only 1 King -> no need for loop)
	index = bits.TrailingZeros64(chess.Boards[BLACK_KING])
	whiteInDanger |= KING_ATTACK[index] & blacks_empty_whiteKing

	//Restore the value of whiteKing
	chess.Boards[WHITE_KING] = wk
	return whiteInDanger
}

func (chess *Chess) GenerateWhiteAttacks() uint64 {
	var (
		FILE_A, FILE_H = FILE_MASK[0], FILE_MASK[7]
		//whites cannot land to squares that is occupied by another whites, so we negate whites bitboard
		whitesCanLandTo, blacks = ^chess.GenerateAllWhites(), chess.GenerateAllBlacks()
		temp, whiteCanAttack    uint64
		index                   int
	)

	//Calculate white pawns attack squares
	whiteCanAttack |= (chess.Boards[WHITE_PAWN] << 9) & blacks & ^FILE_H
	whiteCanAttack |= (chess.Boards[WHITE_PAWN] << 7) & blacks & ^FILE_A
	if 47 >= chess.EnPassantTarget && chess.EnPassantTarget >= 40 {
		whiteCanAttack |= 0x1 << chess.EnPassantTarget
	}

	//Calculate white rooks and white queens (horizontal and vertical only) attack squares
	temp = chess.Boards[WHITE_ROOK] | chess.Boards[WHITE_QUEEN]
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteCanAttack |= HAndVMoves(index, chess) & whitesCanLandTo
		ClearBit(index, &temp)
	}

	//Calculate white knights attack squares
	temp = chess.Boards[WHITE_KNIGHT]
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteCanAttack |= KNIGHT_ATTACK[index] & whitesCanLandTo
		ClearBit(index, &temp)
	}

	//Calculate white bishops and white queens (diagonal and anti diagonal only) attack squares
	temp = chess.Boards[WHITE_BISHOP] | chess.Boards[WHITE_QUEEN]
	for temp != 0 {
		index = bits.TrailingZeros64(temp)
		whiteCanAttack |= DAndAntiDMoves(index, chess) & whitesCanLandTo
		ClearBit(index, &temp)
	}

	//Calculate white king attack squares (since there is only one King, no need to use a loop)
	index = bits.TrailingZeros64(chess.Boards[WHITE_KING])
	whiteCanAttack |= KING_ATTACK[index] & whitesCanLandTo

	return whiteCanAttack
}

// Check if the White King is under attacked (is checked)
func (chess *Chess) IsWhiteKingChecked() bool {
	chess.Flip()
	blackAttacks := FlipVertical(chess.GenerateWhiteAttacks())
	chess.Flip()
	return chess.Boards[5]&blackAttacks != 0
}

func (chess *Chess) CalculateWhiteKingAttackers() (uint64, bool) {
	var (
		FILE_A, FILE_H = FILE_MASK[0], FILE_MASK[7]
		kingIndex      = bits.TrailingZeros64(chess.Boards[WHITE_KING])
	)

	//Calculate attacker
	pawnAttackers := (chess.Boards[WHITE_KING] << 7) & ^FILE_A & chess.Boards[BLACK_PAWN]
	pawnAttackers |= (chess.Boards[WHITE_KING] << 9) & ^FILE_H & chess.Boards[BLACK_PAWN]
	rookAttackers := HAndVMoves(kingIndex, chess) & chess.Boards[BLACK_ROOK]
	knightAttackers := KNIGHT_ATTACK[kingIndex] & chess.Boards[BLACK_KNIGHT]
	bishopAttackers := DAndAntiDMoves(kingIndex, chess) & chess.Boards[BLACK_BISHOP]
	queenAttackers := (HAndVMoves(kingIndex, chess) | DAndAntiDMoves(kingIndex, chess)) & chess.Boards[BLACK_QUEEN]

	return pawnAttackers | rookAttackers | knightAttackers | bishopAttackers | queenAttackers,
		rookAttackers != 0 || bishopAttackers != 0 || queenAttackers != 0
}

func (chess *Chess) WhiteMoveGeneration() []Move {
	//Variables declaration
	var (
		moves     []Move
		move      Move = Move{Castling: 0}
		kingIndex      = bits.TrailingZeros64(chess.Boards[WHITE_KING])

		//Temporary bitboard (since pin pieces can only moving along the line, we remove them from the bitboard)
		wp = chess.Boards[WHITE_PAWN]
		wr = chess.Boards[WHITE_ROOK]
		wn = chess.Boards[WHITE_KNIGHT]
		wb = chess.Boards[WHITE_BISHOP]
		wq = chess.Boards[WHITE_QUEEN]

		//All whites, blacks, empty bitboard
		whites, blacks = chess.GenerateAllWhites(), chess.GenerateAllBlacks()
		empty          = ^(whites | blacks)

		//Temporary index used for looping through bitboard and direction temporary variable
		index, direction int

		//RANK, FILE constant
		RANK_2, RANK_4, FILE_A, FILE_H = RANK_MASK[1], RANK_MASK[3], FILE_MASK[0], FILE_MASK[7]
	)

	//Calculate King movement/evasion
	kingMove := ^chess.GenerateWhiteKingInDanger() & KING_ATTACK[kingIndex] & ^whites // = KING_ATTACK & !(kingInDanger | whites)
	move.FromBoard = WHITE_KING
	move.FromIndex = kingIndex
	move.ToBoard = WHITE_KING
	for kingMove != 0 {
		index = bits.TrailingZeros64(kingMove)
		move.ToIndex = index
		moves = append(moves, move)
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
		wpTemp, bpTemp := chess.Boards[WHITE_PAWN], chess.Boards[BLACK_PAWN]
		epMove := ((0x1 << chess.EnPassantTarget) >> 9) & wp & ^FILE_A
		epMove |= ((0x1 << chess.EnPassantTarget) >> 7) & wp & ^FILE_H

		//Loop though all ep move found
		for epMove != 0 {
			//Perform the en passant move
			index = bits.TrailingZeros64(epMove)
			ClearBit(index, &chess.Boards[WHITE_PAWN])
			SetBit(chess.EnPassantTarget, &chess.Boards[WHITE_PAWN])
			ClearBit(chess.EnPassantTarget-8, &chess.Boards[BLACK_PAWN])

			//If the en passant move not lead to a check, add them to the list
			if !chess.IsWhiteKingChecked() {
				move.FromBoard = WHITE_PAWN
				move.FromIndex = index
				move.ToBoard = WHITE_PAWN
				move.ToIndex = chess.EnPassantTarget
				moves = append(moves, move)
			}

			//Restore the board after making a move
			chess.Boards[WHITE_PAWN] = wpTemp
			chess.Boards[BLACK_PAWN] = bpTemp

			//Clear the ep move
			ClearBit(index, &epMove)
		}
	}

	//Calculate pin pieces' moves (whether it's single check or no check, we still need to remove the pin pieces beforehand)
	var (
		temp, rayline                                uint64
		min, max, pseudoAttackerIndex, pinPieceIndex int
		pinMoves                                     []Move
	)

	temp = chess.Boards[BLACK_ROOK] | chess.Boards[BLACK_QUEEN]
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
				switch {
				case IsPieceAtIndex(wr, pinPieceIndex):
					move.FromBoard = WHITE_ROOK
					move.FromIndex = pinPieceIndex
					move.ToBoard = WHITE_ROOK
					//Add capture move first
					move.ToIndex = pseudoAttackerIndex
					pinMoves = append(pinMoves, move)
					//Add other move between the rayline
					for i := min + direction; i < max; i += direction {
						if i != pinPieceIndex {
							move.ToIndex = i
							pinMoves = append(pinMoves, move)
						}
					}
				case IsPieceAtIndex(wq, pinPieceIndex):
					move.FromBoard = WHITE_QUEEN
					move.FromIndex = pinPieceIndex
					move.ToBoard = WHITE_QUEEN
					//Add capture move first
					move.ToIndex = pseudoAttackerIndex
					pinMoves = append(pinMoves, move)
					//Add other move between the rayline
					for i := min + direction; i < max; i += direction {
						if i != pinPieceIndex {
							move.ToIndex = i
							pinMoves = append(pinMoves, move)
						}
					}
				case IsPieceAtIndex(wp, pinPieceIndex) && direction == FILE:
					pawnMoves := (0x1 << (pinPieceIndex + 8)) & empty
					pawnMoves |= (0x1 << (pinPieceIndex + 16)) & empty & (empty << 8) & RANK_4

					move.FromBoard = WHITE_PAWN
					move.FromIndex = pinPieceIndex
					move.ToBoard = WHITE_PAWN
					for pawnMoves != 0 {
						index = bits.TrailingZeros64(pawnMoves)
						move.ToIndex = index
						pinMoves = append(pinMoves, move)
						ClearBit(index, &pawnMoves)
					}
				}

				ClearBitAcrossBoards(pinPieceIndex, &wp, &wr, &wn, &wb, &wq)
			}
		}

		ClearBit(pseudoAttackerIndex, &temp)
	}

	temp = chess.Boards[BLACK_BISHOP] | chess.Boards[BLACK_QUEEN]
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
				switch {
				case IsPieceAtIndex(wb, pinPieceIndex):
					move.FromBoard = WHITE_BISHOP
					move.FromIndex = pinPieceIndex
					move.ToBoard = WHITE_BISHOP
					//Add capture move first
					move.ToIndex = pseudoAttackerIndex
					pinMoves = append(pinMoves, move)
					//Add other move between the rayline
					for i := min + direction; i < max; i += direction {
						if i != pinPieceIndex {
							move.ToIndex = i
							pinMoves = append(pinMoves, move)
						}
					}
				case IsPieceAtIndex(wq, pinPieceIndex):
					move.FromBoard = WHITE_QUEEN
					move.FromIndex = pinPieceIndex
					move.ToBoard = WHITE_QUEEN
					//Add capture move first
					move.ToIndex = pseudoAttackerIndex
					pinMoves = append(pinMoves, move)
					//Add other move between the rayline
					for i := min + direction; i < max; i += direction {
						if i != pinPieceIndex {
							move.ToIndex = i
							pinMoves = append(pinMoves, move)
						}
					}
				case IsPieceAtIndex(wp, pinPieceIndex):
					pawnMoves := (0x1 << (pinPieceIndex + direction)) & uint64(0x1<<pseudoAttackerIndex)
					move.FromBoard = WHITE_PAWN
					move.FromIndex = pinPieceIndex
					if pawnMoves != 0 {
						move.ToIndex = bits.TrailingZeros64(pawnMoves)
						switch {
						case 56 <= pseudoAttackerIndex && pseudoAttackerIndex <= 63:
							move.ToBoard = WHITE_QUEEN //Promote to Queen
							pinMoves = append(pinMoves, move)
							move.ToBoard = WHITE_ROOK //Promote to Rook
							pinMoves = append(pinMoves, move)
							move.ToBoard = WHITE_BISHOP //Promote to Bishop
							pinMoves = append(pinMoves, move)
							move.ToBoard = WHITE_KNIGHT //Promote to Knight
							pinMoves = append(pinMoves, move)
						default:
							move.ToBoard = WHITE_PAWN
							pinMoves = append(pinMoves, move)
						}
					}
				}

				ClearBitAcrossBoards(pinPieceIndex, &wp, &wr, &wn, &wb, &wq)
			}
		}

		ClearBit(pseudoAttackerIndex, &temp)
	}

	//Handling single check
	if bits.OnesCount64(attackers) == 1 {
		attackerIndex := bits.TrailingZeros64(attackers)
		move.ToIndex = attackerIndex

		/*---Calculate attacker capturing moves---*/
		capture := (attackers >> 7) & ^FILE_H & wp
		capture |= (attackers >> 9) & ^FILE_A & wp
		move.FromBoard = WHITE_PAWN
		if 56 <= attackerIndex && attackerIndex <= 63 {
			for capture != 0 {
				index = bits.TrailingZeros64(capture)
				move.FromIndex = index

				move.ToBoard = WHITE_QUEEN //Promote to Queen
				moves = append(moves, move)
				move.ToBoard = WHITE_ROOK //Promote to Rook
				moves = append(moves, move)
				move.ToBoard = WHITE_BISHOP //Promote to Bishop
				moves = append(moves, move)
				move.ToBoard = WHITE_KNIGHT //Promote to Knight
				moves = append(moves, move)

				ClearBit(index, &capture)
			}
		} else {
			move.ToBoard = WHITE_PAWN
			for capture != 0 {
				index = bits.TrailingZeros64(capture)
				move.FromIndex = index
				moves = append(moves, move)
				ClearBit(index, &capture)
			}
		}

		capture |= HAndVMoves(attackerIndex, chess) & wr
		move.FromBoard, move.ToBoard = WHITE_ROOK, WHITE_ROOK
		for capture != 0 {
			index = bits.TrailingZeros64(capture)
			move.FromIndex = index
			moves = append(moves, move)
			ClearBit(index, &capture)
		}

		capture |= KNIGHT_ATTACK[attackerIndex] & wn
		move.FromBoard, move.ToBoard = WHITE_KNIGHT, WHITE_KNIGHT
		for capture != 0 {
			index = bits.TrailingZeros64(capture)
			move.FromIndex = index
			moves = append(moves, move)
			ClearBit(index, &capture)
		}

		capture |= DAndAntiDMoves(attackerIndex, chess) & wb
		move.FromBoard, move.ToBoard = WHITE_BISHOP, WHITE_BISHOP
		for capture != 0 {
			index = bits.TrailingZeros64(capture)
			move.FromIndex = index
			moves = append(moves, move)
			ClearBit(index, &capture)
		}

		capture |= (HAndVMoves(attackerIndex, chess) | DAndAntiDMoves(attackerIndex, chess)) & wq
		move.FromBoard, move.ToBoard = WHITE_QUEEN, WHITE_QUEEN
		for capture != 0 {
			index = bits.TrailingZeros64(capture)
			move.FromIndex = index
			moves = append(moves, move)
			ClearBit(index, &capture)
		}

		/*---Calculate attacker blocking moves---*/
		if hasSPAttacker {
			var blockMoves uint64
			min = Min(attackerIndex, kingIndex)
			max = Max(attackerIndex, kingIndex)

			switch {
			case IsAtSameRank(attackerIndex, kingIndex):
				direction = RANK
				//Handle pawn advance
				move.FromBoard = WHITE_PAWN
				for i := min + RANK; i < max; i += RANK {
					move.ToIndex = i

					blockMoves |= ((0x1 << i) >> 8) & wp
					blockMoves |= ((0x1 << i) >> 16) & (empty >> 8) & wp & RANK_2

					switch {
					case 56 <= i && i <= 63:
						for blockMoves != 0 {
							index = bits.TrailingZeros64(blockMoves)
							move.FromIndex = index

							move.ToBoard = WHITE_QUEEN //Promote to Queen
							moves = append(moves, move)
							move.ToBoard = WHITE_ROOK //Promote to Rook
							moves = append(moves, move)
							move.ToBoard = WHITE_BISHOP //Promote to Bishop
							moves = append(moves, move)
							move.ToBoard = WHITE_KNIGHT //Promote to Knight
							moves = append(moves, move)

							ClearBit(index, &blockMoves)
						}
					default:
						for blockMoves != 0 {
							index = bits.TrailingZeros64(blockMoves)
							move.FromIndex = index
							move.ToBoard = WHITE_PAWN
							moves = append(moves, move)
							ClearBit(index, &blockMoves)
						}

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

					move.FromBoard = WHITE_PAWN
					move.ToBoard = WHITE_PAWN
					move.ToIndex = i
					for blockMoves != 0 {
						index = bits.TrailingZeros64(blockMoves)
						move.FromIndex = index
						moves = append(moves, move)
						ClearBit(index, &blockMoves)
					}
				}

				blockMoves |= HAndVMoves(i, chess) & wr
				move.FromBoard = WHITE_ROOK
				move.ToBoard = WHITE_ROOK
				move.ToIndex = i
				for blockMoves != 0 {
					index = bits.TrailingZeros64(blockMoves)
					move.FromIndex = index
					moves = append(moves, move)
					ClearBit(index, &blockMoves)
				}

				blockMoves |= KNIGHT_ATTACK[i] & wn
				move.FromBoard = WHITE_KNIGHT
				move.ToBoard = WHITE_KNIGHT
				move.ToIndex = i
				for blockMoves != 0 {
					index = bits.TrailingZeros64(blockMoves)
					move.FromIndex = index
					moves = append(moves, move)
					ClearBit(index, &blockMoves)
				}

				blockMoves |= DAndAntiDMoves(i, chess) & wb
				move.FromBoard = WHITE_BISHOP
				move.ToBoard = WHITE_BISHOP
				move.ToIndex = i
				for blockMoves != 0 {
					index = bits.TrailingZeros64(blockMoves)
					move.FromIndex = index
					moves = append(moves, move)
					ClearBit(index, &blockMoves)
				}

				blockMoves |= (HAndVMoves(i, chess) | DAndAntiDMoves(i, chess)) & wq
				move.FromBoard = WHITE_QUEEN
				move.ToBoard = WHITE_QUEEN
				move.ToIndex = i
				for blockMoves != 0 {
					index = bits.TrailingZeros64(blockMoves)
					move.FromIndex = index
					moves = append(moves, move)
					ClearBit(index, &blockMoves)
				}
			}
		}

		return moves
	}

	//No check
	var (
		pawnMoves, rookMoves, knightMoves, bishopMoves, queenMoves uint64
		pieceIndex                                                 int
	)

	pawnMoves = (wp << 8) & empty
	move.FromBoard = WHITE_PAWN
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move.FromIndex, move.ToIndex = index-8, index
		//If the destination stay at RANK_8, it's pawn promotion
		if 56 <= index && index <= 63 {
			move.ToBoard = WHITE_QUEEN //Promote to Queen
			moves = append(moves, move)
			move.ToBoard = WHITE_ROOK //Promote to Rook
			moves = append(moves, move)
			move.ToBoard = WHITE_BISHOP //Promote to Bishop
			moves = append(moves, move)
			move.ToBoard = WHITE_KNIGHT //Promote to Knight
			moves = append(moves, move)
		} else {
			move.ToBoard = WHITE_PAWN
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (wp << 16) & empty & (empty << 8) & RANK_4
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move.FromIndex, move.ToIndex, move.ToBoard = index-16, index, WHITE_PAWN
		moves = append(moves, move)
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (wp << 7) & blacks & ^FILE_A
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move.FromIndex, move.ToIndex = index-7, index
		//If the destination stay at RANK_8, it's pawn promotion
		if 56 <= index && index <= 63 {
			move.ToBoard = WHITE_QUEEN //Promote to Queen
			moves = append(moves, move)
			move.ToBoard = WHITE_ROOK //Promote to Rook
			moves = append(moves, move)
			move.ToBoard = WHITE_BISHOP //Promote to Bishop
			moves = append(moves, move)
			move.ToBoard = WHITE_KNIGHT //Promote to Knight
			moves = append(moves, move)
		} else {
			move.ToBoard = WHITE_PAWN
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}
	pawnMoves |= (wp << 9) & blacks & ^FILE_H
	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move.FromIndex, move.ToIndex = index-9, index
		//If the destination stay at RANK_8, it's pawn promotion
		if 56 <= index && index <= 63 {
			move.ToBoard = WHITE_QUEEN //Promote to Queen
			moves = append(moves, move)
			move.ToBoard = WHITE_ROOK //Promote to Rook
			moves = append(moves, move)
			move.ToBoard = WHITE_BISHOP //Promote to Bishop
			moves = append(moves, move)
			move.ToBoard = WHITE_KNIGHT //Promote to Knight
			moves = append(moves, move)
		} else {
			move.ToBoard = WHITE_PAWN
			moves = append(moves, move)
		}
		ClearBit(index, &pawnMoves)
	}

	move.FromBoard, move.ToBoard = WHITE_ROOK, WHITE_ROOK
	for wr != 0 {
		pieceIndex = bits.TrailingZeros64(wr)
		move.FromIndex = pieceIndex

		rookMoves = HAndVMoves(pieceIndex, chess) & ^whites
		for rookMoves != 0 {
			index = bits.TrailingZeros64(rookMoves)
			move.ToIndex = index
			moves = append(moves, move)
			ClearBit(index, &rookMoves)
		}

		ClearBit(pieceIndex, &wr)
	}

	move.FromBoard, move.ToBoard = WHITE_KNIGHT, WHITE_KNIGHT
	for wn != 0 {
		pieceIndex = bits.TrailingZeros64(wn)
		move.FromIndex = pieceIndex

		knightMoves = KNIGHT_ATTACK[pieceIndex] & ^whites
		for knightMoves != 0 {
			index = bits.TrailingZeros64(knightMoves)
			move.ToIndex = index
			moves = append(moves, move)
			ClearBit(index, &knightMoves)
		}

		ClearBit(pieceIndex, &wn)
	}

	move.FromBoard, move.ToBoard = WHITE_BISHOP, WHITE_BISHOP
	for wb != 0 {
		pieceIndex = bits.TrailingZeros64(wb)
		move.FromIndex = pieceIndex

		bishopMoves = DAndAntiDMoves(pieceIndex, chess) & ^whites
		for bishopMoves != 0 {
			index = bits.TrailingZeros64(bishopMoves)
			move.ToIndex = index
			moves = append(moves, move)
			ClearBit(index, &bishopMoves)
		}

		ClearBit(pieceIndex, &wb)
	}

	move.FromBoard, move.ToBoard = WHITE_QUEEN, WHITE_QUEEN
	for wq != 0 {
		pieceIndex = bits.TrailingZeros64(wq)
		move.FromIndex = pieceIndex

		queenMoves = (HAndVMoves(pieceIndex, chess) | DAndAntiDMoves(pieceIndex, chess)) & ^whites
		for queenMoves != 0 {
			index = bits.TrailingZeros64(queenMoves)
			move.ToIndex = index
			moves = append(moves, move)
			ClearBit(index, &queenMoves)
		}

		ClearBit(pieceIndex, &wq)
	}

	//Append pin moves into list of moves
	moves = append(moves, pinMoves...)

	//Handling castling
	whiteKingInDanger := chess.GenerateWhiteKingInDanger()
	if (chess.CastlingPrivilege&int(WHITE_KING_SIDE)) == int(WHITE_KING_SIDE) && (empty&0x6) == 0x6 && (whiteKingInDanger&0xE) == 0 {
		move.Castling = WHITE_KING_SIDE
		moves = append(moves, move)
	}

	if (chess.CastlingPrivilege&int(WHITE_QUEEN_SIDE)) == int(WHITE_QUEEN_SIDE) && (empty&0x70) == 0x70 && (whiteKingInDanger&0x38) == 0 {
		move.Castling = WHITE_QUEEN_SIDE
		moves = append(moves, move)
	}

	return moves
}

func (chess *Chess) MoveGeneration() []Move {
	if chess.SideToMove == BLACK {
		chess.Flip()
		moves := chess.WhiteMoveGeneration()
		chess.Flip()

		reflectMoves := []Move{}
		for _, move := range moves {
			switch move.Castling {
			case WHITE_KING_SIDE:
				move.Castling = BLACK_KING_SIDE
				reflectMoves = append(reflectMoves, move)
			case WHITE_QUEEN_SIDE:
				move.Castling = BLACK_QUEEN_SIDE
				reflectMoves = append(reflectMoves, move)
			default:
				move.FromBoard += 6
				move.ToBoard += 6
				move.FromIndex, move.ToIndex = FlipIndexVertical(move.FromIndex), FlipIndexVertical(move.ToIndex)
				reflectMoves = append(reflectMoves, move)
				//fmt.Printf("Flip move: %s\n", move.String())
			}
		}
		return reflectMoves
	}

	return chess.WhiteMoveGeneration()
}
