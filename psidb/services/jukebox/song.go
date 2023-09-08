package jukebox

import (
	"math"
	"time"

	"github.com/hajimehoshi/oto"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

const (
	sampleRate = 44100
	channelNum = 2
	bitDepth   = 2
)

type MidiNote struct {
}

type MidiCommand struct {
}

// Note XX / F# /
// Note XX / G
// Note XX / G#
// Note XX / A
// Note XX / A#
// Note XX / B

type ISong interface {
	PlayTone(req PlayToneRequest) error
}

type Song struct {
	psi.NodeBase

	Name string `json:"name"`
}

func (s *Song) PsiNodeName() string { return s.Name }

var otoCtx *oto.Context
var otoPlayer *oto.Player
var otoRequests = make(chan []FreqDuration, 4096)

var logger = logging.GetLogger("jukebox")

func init() {
	ctx, err := oto.NewContext(sampleRate, channelNum, bitDepth, 8192)

	if err != nil {
		panic(err)
	}

	otoCtx = ctx
	otoPlayer = ctx.NewPlayer()

	go func() {
		for notes := range otoRequests {
			for _, note := range notes {
				freq := note.Frequency
				duration := note.Duration

				if freq == 0 {
					time.Sleep(time.Duration(duration) * time.Millisecond)
					continue
				}

				numSamples := int(float64(sampleRate) * (duration / 1000))
				samples := make([]byte, numSamples*channelNum*bitDepth)

				for totalSamples := 0; totalSamples < numSamples; totalSamples += sampleRate {
					sampleCount := sampleRate

					if totalSamples+sampleRate > numSamples {
						sampleCount = numSamples - totalSamples
					}

					for sampleIndex := 0; sampleIndex < sampleCount; sampleIndex++ {
						i := totalSamples + sampleIndex
						val := int16(32767.0 * math.Sin(2.0*math.Pi*freq*float64(i)/float64(sampleRate)))
						samples[4*sampleIndex], samples[4*sampleIndex+1] = byte(val), byte(val>>8)   // Left channel
						samples[4*sampleIndex+2], samples[4*sampleIndex+3] = byte(val), byte(val>>8) // Right channel
					}

					if _, err := otoPlayer.Write(samples); err != nil {
						logger.Error(err)
					}

					totalSamples += sampleRate
				}
			}
		}
	}()
}

func midiNoteToFrequency(note int) float64 {
	// Convert MIDI note number to frequency
	return 440.0 * math.Pow(2.0, float64(note-69)/12.0)
}

type FreqDuration struct {
	Frequency float64 `json:"freq" jsonschema:"title=Freq,type=number,minimum=0.0,maximum=20000.0"`
	Duration  float64 `json:"duration" jsonschema:"title=Duration,description=Duration in milliseconds,type=number,format=duration"`
}

type PlayToneRequest struct {
	FreqDurations []FreqDuration `json:"freq_durations" jsonschema:"title=FreqDurations,description=List of frequencies and duration in milliseconds to play"`
}

func (s *Song) PlayTone(req PlayToneRequest) error {
	otoRequests <- req.FreqDurations

	return nil
}

var SongInterface = psi.DefineNodeInterface[ISong]()
var SongType = psi.DefineNodeType[*Song](psi.WithInterfaceFromNode(SongInterface))
