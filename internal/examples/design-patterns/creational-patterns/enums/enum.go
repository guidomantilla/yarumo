package main

type Season int64

const (
	Summer Season = 0
	Autumn        = 1
	Winter        = 2
	Spring        = 3
)

func (s Season) String() string {
	switch s {
	case Summer:
		return "summer"
	case Autumn:
		return "autumn"
	case Winter:
		return "winter"
	case Spring:
		return "spring"
	}
	return "unknown"
}
