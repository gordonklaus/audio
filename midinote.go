package audio

import "math"

func MIDINote(key string) (uint8, bool) {
	note, ok := midiNote[key]
	return note, ok
}

func MIDINotePitch(note uint8) float64 {
	return math.Log2(MIDINoteFrequency(note))
}

func MIDINoteFrequency(note uint8) float64 {
	return 440 * math.Exp2((float64(note)-69)/12)
}

var midiNote = map[string]uint8{
	"Z": 60,
	"S": 61,
	"X": 62,
	"D": 63,
	"C": 64,
	"V": 65,
	"G": 66,
	"B": 67,
	"H": 68,
	"N": 69,
	"J": 70,
	"M": 71,
	",": 72,
	"L": 73,
	".": 74,
	";": 75,
	"/": 76,
	"Q": 72,
	"2": 73,
	"W": 74,
	"3": 75,
	"E": 76,
	"R": 77,
	"5": 78,
	"T": 79,
	"6": 80,
	"Y": 81,
	"7": 82,
	"U": 83,
	"I": 84,
	"9": 85,
	"O": 86,
	"0": 87,
	"P": 88,
	"[": 89,
	"=": 90,
	"]": 91,
	`\`: 93,
}
