package engine

// BIT_MASK constant
var (
	// RANK constant (from RANK_1 to RANK_8)
	RANK_MASK = [8]uint64{
		0xFF, 0xFF00, 0xFF0000, 0xFF000000, 0xFF00000000, 0xFF0000000000, 0xFF000000000000, 0xFF00000000000000,
	}

	// FILE constant (from FILE_1 to FILE_8)
	FILE_MASK = [8]uint64{
		0x8080808080808080, 0x4040404040404040, 0x2020202020202020, 0x1010101010101010,
		0x808080808080808, 0x404040404040404, 0x202020202020202, 0x101010101010101,
	}

	// DIAGONAL constant (from DIAGONAL_1 to DIAGONAL_15) (DIAGONAL is the y = x line, from top-left to bottom-right)
	DIAGONAL_MASK = [15]uint64{
		0x8000000000000000, 0x4080000000000000, 0x2040800000000000, 0x1020408000000000, 0x810204080000000,
		0x408102040800000, 0x204081020408000, 0x102040810204080, 0x1020408102040, 0x10204081020,
		0x102040810, 0x1020408, 0x10204, 0x102, 0x1,
	}

	// ANTI_DIAGONAL constant (from ANTI_DIAGONAL_1 to ANTI_DIAGONAL_15) (ANTI_DIAGONAL is the y = -x line, from top-right to bottom-left)
	ANTI_DIAGONAL_MASK = [15]uint64{
		0x100000000000000, 0x201000000000000, 0x402010000000000, 0x804020100000000, 0x1008040201000000,
		0x2010080402010000, 0x4020100804020100, 0x8040201008040201, 0x80402010080402, 0x804020100804,
		0x8040201008, 0x80402010, 0x804020, 0x8040, 0x80,
	}

	// Knight attack bit pattern array. This pattern array map to the bitboard representation (bitboard has index 0 in the LSB)
	// For example, if we have a knight bitboard = 1, then it attack pattern is the first element (0x20400 in this case)
	KNIGHT_ATTACK = [64]uint64{
		0x20400, 0x50800, 0xA1100, 0x142200, 0x284400, 0x508800, 0xA01000, 0x402000,
		0x2040004, 0x5080008, 0xA110011, 0x14220022, 0x28440044, 0x50880088, 0xA0100010, 0x40200020,
		0x204000402, 0x508000805, 0xA1100110A, 0x1422002214, 0x2844004428, 0x5088008850, 0xA0100010A0, 0x4020002040,
		0x20400040200, 0x50800080500, 0xA1100110A00, 0x142200221400, 0x284400442800, 0x508800885000, 0xA0100010A000, 0x402000204000,
		0x2040004020000, 0x5080008050000, 0xA1100110A0000, 0x14220022140000, 0x28440044280000, 0x50880088500000, 0xA0100010A00000, 0x40200020400000,
		0x204000402000000, 0x508000805000000, 0xA1100110A000000, 0x1422002214000000, 0x2844004428000000, 0x5088008850000000, 0xA0100010A0000000, 0x4020002040000000,
		0x400040200000000, 0x800080500000000, 0x1100110A00000000, 0x2200221400000000, 0x4400442800000000, 0x8800885000000000, 0x100010A000000000, 0x2000204000000000,
		0x4020000000000, 0x8050000000000, 0x110A0000000000, 0x22140000000000, 0x44280000000000, 0x88500000000000, 0x10A00000000000, 0x20400000000000,
	}

	// King attack bit pattern array. This pattern array map to the bitboard representation (bitboard has index 0 in the LSB)
	// For example, if we have a king bitboard = 1, then it attack pattern is the first element (0x302 in this case)
	KING_ATTACK = [64]uint64{
		0x302, 0x705, 0xE0A, 0x1C14, 0x3828, 0x7050, 0xE0A0, 0xC040,
		0x30203, 0x70507, 0xE0A0E, 0x1C141C, 0x382838, 0x705070, 0xE0A0E0, 0xC040C0,
		0x3020300, 0x7050700, 0xE0A0E00, 0x1C141C00, 0x38283800, 0x70507000, 0xE0A0E000, 0xC040C000,
		0x302030000, 0x705070000, 0xE0A0E0000, 0x1C141C0000, 0x3828380000, 0x7050700000, 0xE0A0E00000, 0xC040C00000,
		0x30203000000, 0x70507000000, 0xE0A0E000000, 0x1C141C000000, 0x382838000000, 0x705070000000, 0xE0A0E0000000, 0xC040C0000000,
		0x3020300000000, 0x7050700000000, 0xE0A0E00000000, 0x1C141C00000000, 0x38283800000000, 0x70507000000000, 0xE0A0E000000000, 0xC040C000000000,
		0x302030000000000, 0x705070000000000, 0xE0A0E0000000000, 0x1C141C0000000000, 0x3828380000000000, 0x7050700000000000, 0xE0A0E00000000000, 0xC040C00000000000,
		0x203000000000000, 0x507000000000000, 0xA0E000000000000, 0x141C000000000000, 0x2838000000000000, 0x5070000000000000, 0xA0E0000000000000, 0x40C0000000000000,
	}
)

// Chess struct
type Chess struct {
	WhitePawns        uint64
	WhiteRooks        uint64
	WhiteKnights      uint64
	WhiteBishops      uint64
	WhiteQueens       uint64
	WhiteKing         uint64
	BlackPawns        uint64
	BlackRooks        uint64
	BlackKnights      uint64
	BlackBishops      uint64
	BlackQueens       uint64
	BlackKing         uint64
	SideToMove        bool // True if White turn, false if Black turn
	EnPassantTarget   int  // Range from 0 to 63
	CastlingPrivilege int  // Use 4 bit integer to represent: KQkq (exactly in this order)
	Halfmove          int
	Fullmove          int
}

// Factory method to create new chess instance
func NewChess() *Chess {
	return &Chess{}
}

/*---Define type Direction*/

// Direction in a chess: RANK (1), FILE (8), DIAGONAL (7), ANTI_DIAGONAL (9)
type Direction int

// Method to validate value for Direction
func (d Direction) IsValid() bool {
	return d == RANK || d == FILE || d == DIAGONAL || d == ANTI_DIAGONAL
}

const (
	RANK          Direction = 1
	FILE          Direction = 8
	DIAGONAL      Direction = 7
	ANTI_DIAGONAL Direction = 9
)

/*---Define type Color*/

// Color in chess: WHITE (8) and BLACK (16) (the value is not realy necessary though)
type Color int

// Method to validate value for Color
func (c Color) IsValid() bool {
	return c == BLACK || c == WHITE
}

const (
	WHITE Color = 8
	BLACK Color = 16
)

/*---Define type Castling side*/

// Castling side in chess: WHITE_KING_SIDE (8), WHITE_QUEEN_SIDE (4), BLACK_KING_SIDE(2), BLACK_QUEEN_SIDE (1)
type CastlingSide int

// Method to validate value for Castling side
func (cs CastlingSide) IsValid() bool {
	return cs == WHITE_KING_SIDE || cs == WHITE_QUEEN_SIDE || cs == BLACK_KING_SIDE || cs == BLACK_QUEEN_SIDE
}

const (
	WHITE_KING_SIDE  CastlingSide = 8
	WHITE_QUEEN_SIDE CastlingSide = 4
	BLACK_KING_SIDE  CastlingSide = 2
	BLACK_QUEEN_SIDE CastlingSide = 1
)
