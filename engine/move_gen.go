package engine

import (
	"math/bits"
)

func (chess *Chess) HAndVMoves(index int) uint64 {
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

func (chess *Chess) DAndAntiDMoves(index int) uint64 {
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
		whiteInDanger |= chess.HAndVMoves(index) & blacks_empty_whiteKing
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
		whiteInDanger |= chess.DAndAntiDMoves(index) & blacks_empty_whiteKing
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
		whiteCanAttack |= chess.HAndVMoves(index) & whitesCanLandTo
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
		whiteCanAttack |= chess.DAndAntiDMoves(index) & whitesCanLandTo
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
	return chess.Boards[WHITE_KING]&blackAttacks != 0
}

func (chess *Chess) IsBlackKingChecked() bool {
	return chess.Boards[BLACK_KING]&chess.GenerateWhiteAttacks() != 0
}

func (chess *Chess) CalculateWhiteKingAttackers() (uint64, bool) {
	var (
		FILE_A, FILE_H = FILE_MASK[0], FILE_MASK[7]
		kingIndex      = bits.TrailingZeros64(chess.Boards[WHITE_KING])
	)

	//Calculate attacker
	pawnAttackers := (chess.Boards[WHITE_KING] << 7) & ^FILE_A & chess.Boards[BLACK_PAWN]
	pawnAttackers |= (chess.Boards[WHITE_KING] << 9) & ^FILE_H & chess.Boards[BLACK_PAWN]
	rookAttackers := chess.HAndVMoves(kingIndex) & chess.Boards[BLACK_ROOK]
	knightAttackers := KNIGHT_ATTACK[kingIndex] & chess.Boards[BLACK_KNIGHT]
	bishopAttackers := chess.DAndAntiDMoves(kingIndex) & chess.Boards[BLACK_BISHOP]
	queenAttackers := (chess.HAndVMoves(kingIndex) | chess.DAndAntiDMoves(kingIndex)) & chess.Boards[BLACK_QUEEN]

	return pawnAttackers | rookAttackers | knightAttackers | bishopAttackers | queenAttackers,
		rookAttackers != 0 || bishopAttackers != 0 || queenAttackers != 0
}

func (chess *Chess) WhiteKingMoves() []Move {
	//Variable declaration
	var (
		kingIndex = bits.TrailingZeros64(chess.Boards[WHITE_KING])
		kingMove  = KING_ATTACK[kingIndex] & ^chess.GenerateWhiteKingInDanger()
		blacks    = chess.GenerateAllBlacks()
		empty     = ^(chess.GenerateAllWhites() | blacks)
		move      = Move{
			FromBoard: WHITE_KING,
			FromIndex: kingIndex,
			ToBoard:   WHITE_KING,
		}
		moves []Move
		index int
	)

	//Calculate moves
	kingMove &= (empty | blacks)
	for kingMove != 0 {
		index = bits.TrailingZeros64(kingMove)
		move.ToIndex = index
		moves = append(moves, move)
		ClearBit(index, &kingMove)
	}

	return moves
}

func (chess *Chess) EnPassantMoves() []Move {
	//Variables declaration
	var (
		FILE_A, FILE_H = FILE_MASK[0], FILE_MASK[7]
		move           = Move{
			FromBoard: WHITE_PAWN,
			ToBoard:   WHITE_PAWN,
			ToIndex:   chess.EnPassantTarget,
		}
		moves []Move
		index int
	)

	//Perform en passant move and check if the King is still vulnerable
	if 40 <= chess.EnPassantTarget && chess.EnPassantTarget <= 47 {
		//Record the old values of Whites and Black pawns
		wpTemp, bpTemp := chess.Boards[WHITE_PAWN], chess.Boards[BLACK_PAWN]

		//Check if there're White pawns at EnPassantTarget - 9 or EnPassantTarget - 7
		epMove := ((0x1 << chess.EnPassantTarget) >> 9) & chess.Boards[WHITE_PAWN] & ^FILE_A
		epMove |= ((0x1 << chess.EnPassantTarget) >> 7) & chess.Boards[WHITE_PAWN] & ^FILE_H

		//Loop though all ep move found
		for epMove != 0 {
			//Perform the en passant move
			index = bits.TrailingZeros64(epMove)
			ClearBit(index, &chess.Boards[WHITE_PAWN])                   //Move the White pawn from old position
			SetBit(chess.EnPassantTarget, &chess.Boards[WHITE_PAWN])     //Place the White pawn to new position
			ClearBit(chess.EnPassantTarget-8, &chess.Boards[BLACK_PAWN]) //Remove the Black pawn

			//If the en passant move not lead to a check, add them to the list
			if !chess.IsWhiteKingChecked() {
				move.FromIndex = index
				moves = append(moves, move)
			}

			//Restore the board after making a move
			chess.Boards[WHITE_PAWN] = wpTemp
			chess.Boards[BLACK_PAWN] = bpTemp

			//Clear the ep move
			ClearBit(index, &epMove)
		}
	}

	return moves
}

func (chess *Chess) PinSPMoves(movePiece, capturePiece, pseudoAttackerIndex, pinPieceIndex int, rayline uint64) []Move {
	//Variables declaration
	var (
		move = Move{
			FromBoard: movePiece,
			FromIndex: pinPieceIndex,
			ToBoard:   movePiece,
		}
		moves []Move
		index int
	)

	//Calculate capture move
	move.ToIndex = pseudoAttackerIndex
	moves = append(moves, move)

	//Remove the pin piece from the rayline
	ClearBit(pinPieceIndex, &rayline)

	//Calculate non-capture moves
	for rayline != 0 {
		index = bits.TrailingZeros64(rayline)
		move.ToIndex = index
		moves = append(moves, move)
		ClearBit(index, &rayline)
	}

	return moves
}

func (chess *Chess) PinPawnMovesInFile(pinPieceIndex int, empty uint64) []Move {
	//Variables declaration
	var (
		move = Move{
			FromBoard: WHITE_PAWN,
			FromIndex: pinPieceIndex,
			ToBoard:   WHITE_PAWN,
		}
		moves []Move
		index int
	)

	//In FILE, pin pawn can only advance to the front, or double push (it can't perform promotion)
	pawnMoves := (0x1 << (pinPieceIndex + 8)) & empty
	pawnMoves |= (0x1 << (pinPieceIndex + 16)) & empty & (empty << 8) & RANK_MASK[3]

	for pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move.ToIndex = index
		moves = append(moves, move)
		ClearBit(index, &pawnMoves)
	}

	return moves
}

func (chess *Chess) PinPawnMovesInDiagonals(direction, pinPieceIndex, pseudoAttackerIndex, capturePiece int) []Move {
	//Variables declaration
	var (
		move = Move{
			FromBoard: WHITE_PAWN,
			FromIndex: pinPieceIndex,
		}
		moves []Move
		index int
	)

	//In DIAGONAL or ANTI_DIAGONAL, pin pawns can only perform capture (it can perform promotion)
	pawnMoves := (0x1 << (pinPieceIndex + direction)) & uint64(0x1<<pseudoAttackerIndex)
	if pawnMoves != 0 {
		index = bits.TrailingZeros64(pawnMoves)
		move.ToIndex = index
		switch {
		case 56 <= pseudoAttackerIndex && pseudoAttackerIndex <= 63:
			move.ToBoard = WHITE_QUEEN //Promote to Queen
			moves = append(moves, move)
			move.ToBoard = WHITE_ROOK //Promote to Rook
			moves = append(moves, move)
			move.ToBoard = WHITE_BISHOP //Promote to Bishop
			moves = append(moves, move)
			move.ToBoard = WHITE_KNIGHT //Promote to Knight
			moves = append(moves, move)
		default:
			move.ToBoard = WHITE_PAWN
			moves = append(moves, move)
		}
	}

	return moves
}

func (chess *Chess) CaptureAttackerMoves(attacker, wp, wr, wn, wb, wq uint64, attackerIndex int) []Move {
	//Variables declaration
	var (
		move = Move{
			ToIndex: attackerIndex,
		}
		moves          []Move
		index          int
		FILE_A, FILE_H = FILE_MASK[0], FILE_MASK[7]
	)

	//Pawn capture
	move.FromBoard = WHITE_PAWN
	capture := (attacker >> 7) & ^FILE_H & wp
	capture |= (attacker >> 9) & ^FILE_A & wp
	if 56 <= attackerIndex && attackerIndex <= 63 { //Handle promotion cases
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
	} else { //Non promotion cases
		move.ToBoard = WHITE_PAWN
		for capture != 0 {
			index = bits.TrailingZeros64(capture)
			move.FromIndex = index
			moves = append(moves, move)
			ClearBit(index, &capture)
		}
	}

	//Rook capture
	move.FromBoard, move.ToBoard = WHITE_ROOK, WHITE_ROOK
	capture |= chess.HAndVMoves(attackerIndex) & wr
	for capture != 0 {
		index = bits.TrailingZeros64(capture)
		move.FromIndex = index
		moves = append(moves, move)
		ClearBit(index, &capture)
	}

	//Knight capture
	move.FromBoard, move.ToBoard = WHITE_KNIGHT, WHITE_KNIGHT
	capture |= KNIGHT_ATTACK[attackerIndex] & wn
	for capture != 0 {
		index = bits.TrailingZeros64(capture)
		move.FromIndex = index
		moves = append(moves, move)
		ClearBit(index, &capture)
	}

	//Bishop capture
	move.FromBoard, move.ToBoard = WHITE_BISHOP, WHITE_BISHOP
	capture |= chess.DAndAntiDMoves(attackerIndex) & wb
	for capture != 0 {
		index = bits.TrailingZeros64(capture)
		move.FromIndex = index
		moves = append(moves, move)
		ClearBit(index, &capture)
	}

	//Queen capture
	move.FromBoard, move.ToBoard = WHITE_QUEEN, WHITE_QUEEN
	capture |= (chess.HAndVMoves(attackerIndex) | chess.DAndAntiDMoves(attackerIndex)) & wq
	for capture != 0 {
		index = bits.TrailingZeros64(capture)
		move.FromIndex = index
		moves = append(moves, move)
		ClearBit(index, &capture)
	}

	return moves
}

func (chess *Chess) BlockingAttackerMoves(wp, wr, wn, wb, wq uint64, attackerIndex, kingIndex int) []Move {
	//Variables declaration
	var (
		move             = Move{}
		moves            []Move
		blockMoves       uint64
		min              = Min(attackerIndex, kingIndex)
		max              = Max(attackerIndex, kingIndex)
		direction, index int
		empty            = ^(chess.GenerateAllWhites() | chess.GenerateAllBlacks())
		RANK_2           = RANK_MASK[1]
	)

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

		blockMoves |= chess.HAndVMoves(i) & wr
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

		blockMoves |= chess.DAndAntiDMoves(i) & wb
		move.FromBoard = WHITE_BISHOP
		move.ToBoard = WHITE_BISHOP
		move.ToIndex = i
		for blockMoves != 0 {
			index = bits.TrailingZeros64(blockMoves)
			move.FromIndex = index
			moves = append(moves, move)
			ClearBit(index, &blockMoves)
		}

		blockMoves |= (chess.HAndVMoves(i) | chess.DAndAntiDMoves(i)) & wq
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

	return moves
}

func (chess *Chess) PseudoLegalMoves(wp, wr, wn, wb, wq uint64) []Move {
	//No check
	var (
		pawnMoves, rookMoves, knightMoves, bishopMoves, queenMoves uint64
		pieceIndex, index                                          int
		move                                                       = Move{}
		moves                                                      []Move
		whites, blacks                                             = chess.GenerateAllWhites(), chess.GenerateAllBlacks()
		empty                                                      = ^(whites | blacks)
		RANK_4, FILE_A, FILE_H                                     = RANK_MASK[3], FILE_MASK[0], FILE_MASK[7]
	)

	/*===Pawns moves===*/

	move.FromBoard = WHITE_PAWN
	pawnMoves = (wp << 8) & empty
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

	//Pawn attack to the right
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

	//Pawn attack to the left
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

	/*===Rooks moves===*/

	move.FromBoard, move.ToBoard = WHITE_ROOK, WHITE_ROOK
	for wr != 0 {
		pieceIndex = bits.TrailingZeros64(wr)
		move.FromIndex = pieceIndex

		//Handle non-capture move
		rookMoves = chess.HAndVMoves(pieceIndex) & ^whites
		for rookMoves != 0 {
			index = bits.TrailingZeros64(rookMoves)
			move.ToIndex = index
			moves = append(moves, move)
			ClearBit(index, &rookMoves)
		}

		ClearBit(pieceIndex, &wr)
	}

	/*===Knights moves===*/

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

	/*===Bishops moves===*/

	move.FromBoard, move.ToBoard = WHITE_BISHOP, WHITE_BISHOP
	for wb != 0 {
		pieceIndex = bits.TrailingZeros64(wb)
		move.FromIndex = pieceIndex

		bishopMoves = chess.DAndAntiDMoves(pieceIndex) & ^whites
		for bishopMoves != 0 {
			index = bits.TrailingZeros64(bishopMoves)
			move.ToIndex = index
			moves = append(moves, move)
			ClearBit(index, &bishopMoves)
		}

		ClearBit(pieceIndex, &wb)
	}

	/*===Queens moves===*/

	move.FromBoard, move.ToBoard = WHITE_QUEEN, WHITE_QUEEN
	for wq != 0 {
		pieceIndex = bits.TrailingZeros64(wq)
		move.FromIndex = pieceIndex

		queenMoves = (chess.HAndVMoves(pieceIndex) | chess.DAndAntiDMoves(pieceIndex)) & ^whites
		for queenMoves != 0 {
			index = bits.TrailingZeros64(queenMoves)
			move.ToIndex = index
			moves = append(moves, move)
			ClearBit(index, &queenMoves)
		}

		ClearBit(pieceIndex, &wq)
	}

	return moves
}

func (chess *Chess) WhiteMoveGeneration() []Move {
	var (
		moves []Move

		//Temporary bitboard (since pin pieces can only moving along the line, we remove them from the bitboard)
		wp = chess.Boards[WHITE_PAWN]
		wr = chess.Boards[WHITE_ROOK]
		wn = chess.Boards[WHITE_KNIGHT]
		wb = chess.Boards[WHITE_BISHOP]
		wq = chess.Boards[WHITE_QUEEN]

		kingIndex = bits.TrailingZeros64(chess.Boards[WHITE_KING])

		whites, blacks = chess.GenerateAllWhites(), chess.GenerateAllBlacks()
		empty          = ^(whites | blacks)
	)

	//Generate King moves
	moves = append(moves, chess.WhiteKingMoves()...)

	//Get King's attackers
	attackers, hasSPAttacker := chess.CalculateWhiteKingAttackers()

	//If this is double check, or there is only the King left, then we stop here
	if bits.OnesCount64(attackers) > 1 || wp|wr|wn|wb|wq == 0 {
		return moves
	}

	//Handling en passant
	moves = append(moves, chess.EnPassantMoves()...)

	/*===Calculate pin pieces===*/
	var (
		temp, rayline                                                         uint64
		min, max, pseudoAttackerIndex, pinPieceIndex, direction, capturePiece int
		pinPieceMoves                                                         []Move
	)

	//Calculate pin moves in RANK & FILE directions
	temp = chess.Boards[BLACK_ROOK] | chess.Boards[BLACK_QUEEN]
	for temp != 0 {
		//Get pseudo attacker index
		pseudoAttackerIndex = bits.TrailingZeros64(temp)

		//Calculate capture piece
		if IsPieceAtIndex(chess.Boards[BLACK_ROOK], pseudoAttackerIndex) {
			capturePiece = BLACK_ROOK
		} else {
			capturePiece = BLACK_QUEEN
		}

		//Calculate the min and max of the rayline (include both the King and the enemy piece)
		min = Min(kingIndex, pseudoAttackerIndex)
		max = Max(kingIndex, pseudoAttackerIndex)

		//Check for direction
		switch {
		case IsAtSameRank(kingIndex, pseudoAttackerIndex):
			direction = RANK
		case IsAtSameFile(kingIndex, pseudoAttackerIndex):
			direction = FILE
		default:
			direction = -1
		}

		//If the direction is either RANK or FILE, calculate pin moves
		if direction != -1 {
			//Calculate rayline (not include the King and enemy piece)
			rayline = CalculateRayAttackLine(min, max, direction)

			//Check if there is only 1 ally piece (pin piece) between the King and enemy
			if bits.OnesCount64(rayline&blacks) == 0 && bits.OnesCount64(rayline&whites) == 1 {
				pinPieceIndex = bits.TrailingZeros64(rayline & whites)

				//For both RANK and FILE, only Rooks and Queens can move
				//For File specifically, Pawns can also move
				switch {
				case IsPieceAtIndex(wr, pinPieceIndex):
					pinPieceMoves = append(pinPieceMoves, chess.PinSPMoves(
						WHITE_ROOK, capturePiece, pseudoAttackerIndex, pinPieceIndex, rayline)...)
				case IsPieceAtIndex(wq, pinPieceIndex):
					pinPieceMoves = append(pinPieceMoves, chess.PinSPMoves(
						WHITE_QUEEN, capturePiece, pseudoAttackerIndex, pinPieceIndex, rayline)...)
				case direction == FILE && IsPieceAtIndex(wp, pinPieceIndex):
					pinPieceMoves = append(pinPieceMoves, chess.PinPawnMovesInFile(pinPieceIndex, empty)...)
				}

				//Clear the pin piece out of the temporary bitboards
				ClearBitAcrossBoards(pinPieceIndex, &wp, &wr, &wn, &wb, &wq)
			}

		}

		ClearBit(pseudoAttackerIndex, &temp)
	}

	//Calculate pin moves in DIAGONAL & ANTI_DIAGONAL directions
	temp = chess.Boards[BLACK_BISHOP] | chess.Boards[BLACK_QUEEN]
	for temp != 0 {
		//Get pseudo attacker index
		pseudoAttackerIndex = bits.TrailingZeros64(temp)

		//Calculate capture piece
		if IsPieceAtIndex(chess.Boards[BLACK_BISHOP], pseudoAttackerIndex) {
			capturePiece = BLACK_BISHOP
		} else {
			capturePiece = BLACK_QUEEN
		}

		//Calculate the min and max of the rayline (include both the King and the enemy piece)
		min = Min(kingIndex, pseudoAttackerIndex)
		max = Max(kingIndex, pseudoAttackerIndex)

		//Check for direction
		switch {
		case IsAtSameDiagonal(kingIndex, pseudoAttackerIndex):
			direction = DIAGONAL
		case IsAtSameAntiDiagonal(kingIndex, pseudoAttackerIndex):
			direction = ANTI_DIAGONAL
		default:
			direction = -1
		}

		//If the direction is either DIAGONAL or ANTI_DIAGONAL, calculate pin moves
		if direction != -1 {
			//Calculate rayline (not include the King and enemy piece)
			rayline = CalculateRayAttackLine(min, max, direction)

			//Check if there is only 1 ally piece (pin piece) between the King and enemy
			if bits.OnesCount64(rayline&blacks) == 0 && bits.OnesCount64(rayline&whites) == 1 {
				pinPieceIndex = bits.TrailingZeros64(rayline & whites)

				//For both DIAGONAL and ANTI_DIAGONAL, only Bishops, Queens and Pawns can move
				switch {
				case IsPieceAtIndex(wb, pinPieceIndex):
					pinPieceMoves = append(pinPieceMoves, chess.PinSPMoves(
						WHITE_BISHOP, capturePiece, pseudoAttackerIndex, pinPieceIndex, rayline)...)
				case IsPieceAtIndex(wq, pinPieceIndex):
					pinPieceMoves = append(pinPieceMoves, chess.PinSPMoves(
						WHITE_QUEEN, capturePiece, pseudoAttackerIndex, pinPieceIndex, rayline)...)
				case IsPieceAtIndex(wp, pinPieceIndex):
					pinPieceMoves = append(pinPieceMoves, chess.PinPawnMovesInDiagonals(direction, pinPieceIndex, pseudoAttackerIndex, BLACK_BISHOP)...)
				}

				ClearBitAcrossBoards(pinPieceIndex, &wp, &wr, &wn, &wb, &wq)
			}

		}

		ClearBit(pseudoAttackerIndex, &temp)
	}

	/*===Handling single check===*/
	if bits.OnesCount64(attackers) == 1 {
		attackerIndex := bits.TrailingZeros64(attackers)

		//Calculate capture attacker move
		moves = append(moves, chess.CaptureAttackerMoves(attackers, wp, wr, wn, wb, wq, attackerIndex)...)

		//Calculate blocking attacker move
		if hasSPAttacker {
			moves = append(moves, chess.BlockingAttackerMoves(wp, wr, wn, wb, wq, attackerIndex, kingIndex)...)
		}

		return moves
	}

	/*===No check case===*/
	//Append pseudo legal moves (which in this case, legal)
	moves = append(moves, chess.PseudoLegalMoves(wp, wr, wn, wb, wq)...)

	//Append pin pieces' moves
	moves = append(moves, pinPieceMoves...)

	/*===Handling castling cases===*/
	whiteKingInDanger := chess.GenerateWhiteKingInDanger()
	move := Move{}
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
		defer chess.Flip()

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
				move.FromIndex = FlipIndexVertical(move.FromIndex)
				move.ToIndex = FlipIndexVertical(move.ToIndex)
				reflectMoves = append(reflectMoves, move)
			}
		}
		return reflectMoves
	}

	return chess.WhiteMoveGeneration()
}
