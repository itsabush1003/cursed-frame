package model

type TeamColor uint32

const (
	UNDEFINED TeamColor = iota
	RED
	BLUE
	GREEN
	YELLOW
	CYAN
	MAGENTA
	WHITE
	GLAY
	BLACK
	teamNum
)

const MaxTeamNum int = int(teamNum) - 1
const MinTeamNum int = 2
const MinTeamUser int = 3

func (tc TeamColor) Raw() uint32 {
	return uint32(tc)
}

func (tc TeamColor) ToInt() int {
	return int(tc)
}

func (tc TeamColor) String() string {
	switch tc {
	case RED:
		return "RED"
	case BLUE:
		return "BLUE"
	case GREEN:
		return "GREEN"
	case YELLOW:
		return "YELLOW"
	case CYAN:
		return "CYAN"
	case MAGENTA:
		return "MAGENTA"
	case WHITE:
		return "WHITE"
	case GLAY:
		return "GLAY"
	case BLACK:
		return "BLACK"
	default:
		return "UNDEFINED"
	}
}

type Team struct {
	teamID    uint32
	teamColor TeamColor
	member    []User
}
