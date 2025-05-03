package main

type TPosition struct {
	SLine int
	SColm int
	ELine int
	EColm int
}

func InitPosition(sLine, sColm, eLine, eColm int) TPosition {
	return TPosition{
		SLine: sLine,
		SColm: sColm,
		ELine: eLine,
		EColm: eColm,
	}
}

func InitPositionFromLineAndColm(sLine, sColm int) TPosition {
	return TPosition{
		SLine: sLine,
		SColm: sColm,
		ELine: sLine,
		EColm: sColm,
	}
}

func CreatePosition(sLine, sColm, eLine, eColm int) *TPosition {
	pos := new(TPosition)
	pos.SLine = sLine
	pos.SColm = sColm
	pos.ELine = eLine
	pos.EColm = eColm
	return pos
}

func (position TPosition) Merge(other TPosition) TPosition {
	return TPosition{
		SLine: position.SColm,
		ELine: other.ELine,
		SColm: position.SColm,
		EColm: other.EColm,
	}
}
