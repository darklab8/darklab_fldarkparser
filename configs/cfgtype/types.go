package cfgtype

type Vector struct {
	X float64
	Y float64
	Z float64
}

type TractorID string

type FactionNick string

type Milliseconds = float64

type Seconds = float64

type BaseUniNick string

func (b BaseUniNick) ToStr() string { return string(b) }
