package jukebox

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/ebitengine/oto/v3"
	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/psi"
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

var otoContext *oto.Context
var otoPlayer *SoundPlayer

type SoundPlayer struct {
	player   *oto.Player
	requests chan []FreqDuration

	sampleRate int
	bitDepth   int
	channelNum int

	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter

	time uint64
}

func NewSoundPlayer() *SoundPlayer {
	pipeReader, pipeWriter := io.Pipe()

	sp := &SoundPlayer{
		sampleRate: 44100,
		requests:   make(chan []FreqDuration, 4096),

		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	}

	sp.player = otoContext.NewPlayer(pipeReader)

	goprocess.Go(sp.Run)

	return sp
}

func (sp *SoundPlayer) Run(proc goprocess.Process) {
	proc.Go(func(proc goprocess.Process) {
		sp.player.Play()
	})

	for notes := range sp.requests {
		for _, note := range notes {
			freq := note.Frequency
			duration := note.Duration

			writeSample := func(l, r float64) {
				var buffer [8]byte
				ivl := math.Float32bits(float32(l))
				ivr := math.Float32bits(float32(r))
				binary.LittleEndian.PutUint32(buffer[:4], ivl)
				binary.LittleEndian.PutUint32(buffer[4:], ivr)
				if _, err := sp.pipeWriter.Write(buffer[:]); err != nil {
					logger.Error(err)
				}

				sp.time++
			}

			numSamples := int(float64(sp.sampleRate) * (duration / 1000))

			for i := 0; i < numSamples; i++ {
				val := math.Sin(2.0 * math.Pi * freq * float64(sp.time) / float64(sp.sampleRate))

				writeSample(val, val)
			}
		}
	}
}

var logger = logging.GetLogger("jukebox")

func init() {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 2,
		Format:       oto.FormatFloat32LE,
	})

	if err != nil {
		panic(err)
	}

	otoContext = ctx

	go func() {
		_, _ = <-ready

		otoPlayer = NewSoundPlayer()
	}()
}

func midiNoteToFrequency(note int) float64 {
	// Convert MIDI note number to frequency
	return 440.0 * math.Pow(2.0, float64(note-69)/12.0)
}

type FreqDuration struct {
	Frequency float64 `json:"freq" jsonschema:"title=Freq,type=number,minimum=0.0,maximum=20000.0"`
	Duration  float64 `json:"duration_in_ms" jsonschema:"title=Duration,description=Duration in milliseconds,type=number,format=duration"`
}

type PlayToneRequest struct {
	FreqDurations []FreqDuration `json:"freq_durations" jsonschema:"title=FreqDurations,description=List of frequencies and duration in milliseconds to play"`
}

func (s *Song) PlayTone(req PlayToneRequest) error {
	otoPlayer.requests <- req.FreqDurations

	return nil
}

var SongInterface = psi.DefineNodeInterface[ISong]()
var SongType = psi.DefineNodeType[*Song](psi.WithInterfaceFromNode(SongInterface))
