package jukebox

var midiNoteMap = map[string]int{
	"C":  60,
	"C#": 61,
	"Db": 61,
	"D":  62,
	"D#": 63,
	"Eb": 63,
	"E":  64,
	"F":  65,
	"F#": 66,
	"Gb": 66,
	"G":  67,
	"G#": 68,
	"Ab": 68,
	"A":  69,
	"A#": 70,
	"Bb": 70,
	"B":  71,
}

func shiftKey(base uint8, oct uint8) uint8 {
	if oct > 10 {
		oct = 10
	}

	if oct == 0 {
		return base
	}

	res := base + uint8(12*oct)
	if res > 127 {
		res -= 12
	}

	return res
}
