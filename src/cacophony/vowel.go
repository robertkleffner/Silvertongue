package main

type VHeight int
type VPosition int
type VLength int

type Vowel struct {
	Height     VHeight
	Position   VPosition
	Length     VLength
	Rounded    bool
	Nasalized  bool
	Rhoticized bool
	IpaSymbol  rune
}

const (
	HEIGHT_CLOSE VHeight = iota
	HEIGHT_NEAR_CLOSE
	HEIGHT_CLOSE_MID
	HEIGHT_OPEN_MID
	HEIGHT_NEAR_OPEN
	HEIGHT_OPEN
)

const (
	POSITION_FRONT VPosition = iota
	POSITION_NEAR_FRONT
	POSITION_CENTRAL
	POSITION_NEAR_BACK
	POSITION_BACK
)
