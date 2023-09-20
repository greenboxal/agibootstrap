package jukebox

import (
	"context"
	"strings"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

// IInstrument defines the interface for a musical instrument. It provides methods to play, pause, and reset the instrument.
// It also includes a method to handle the next tick of the instrument's timeline.
type IInstrument interface {
	// Play initiates the playback of the instrument. It returns an error if the playback process encounters any issues.
	Play(ctx context.Context) error

	// Pause halts the playback of the instrument. It returns an error if the pausing process encounters any issues.
	Pause(ctx context.Context) error

	// Reset stops the playback and resets the instrument to its initial state. It returns an error if the reset process encounters any issues.
	Reset(ctx context.Context) error

	// PlayCommandSheet initiates the playback of a command sheet. It takes a request parameter of type PlayCommandSheetRequest.
	// It returns an error if the playback process encounters any issues.
	PlayCommandSheet(ctx context.Context, req PlayCommandSheetRequest) error

	// PlayPrompt initiates the playback of a prompt. It takes a request parameter of type PlayPromptRequest.
	// It returns an error if the playback process encounters any issues.
	PlayPrompt(ctx context.Context, req PlayPromptRequest) error
}

// PlayPromptRequest defines the structure for a prompt request. It includes a prompt and a start timecode.
type PlayPromptRequest struct {
	Prompt        string  `json:"prompt"`         // Prompt is the text of the prompt to be played.
	StartTimecode float64 `json:"start_timecode"` // StartTimecode is the timecode at which the prompt should start playing.
}

// OnNextTickRequest defines the structure for a request to handle the next tick of the instrument's timeline. It includes a next timecode.
type OnNextTickRequest struct {
	NextTimecode float64 `json:"next_timecode"` // NextTimecode is the timecode for the next tick of the instrument's timeline.
}

// PlayCommandSheetRequest defines the structure for a command sheet request. It includes a command sheet and an array of command sheets.
type PlayCommandSheetRequest struct {
	CommandSheet  string   `json:"command_script" jsonschema:"title=Command Sheet Script,type=string"`  // CommandSheet is the script of the command sheet to be played.
	CommandSheets []string `json:"command_scripts" jsonschema:"title=Command Sheet Scripts,type=array"` // CommandSheets is an array of scripts of the command sheets to be played.
}

func (req *PlayCommandSheetRequest) Parse() CommandSheet {
	var result CommandSheet

	src := req.CommandSheet + strings.Join(req.CommandSheets, "\n")
	lines := strings.Split(src, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		sheet, err := CommandSheetParser.ParseString("", line)

		if err != nil {
			logger.Error(err)
			continue
		}

		logger.Infow("Play line", "line", line)

		result.Commands = append(result.Commands, sheet.Commands...)
	}

	return result
}

// Tone is a frequency and duration pair
//
//	Be sure to provide each Tone as { freq: float64, duration: float64 /* 0.0-1.0, 1.0 = 1 beat unit */, note: /[A-F]#?/, bpm?: float64 }
type Tone struct {
	Octave     int     `json:"octave" jsonschema:"title=Octave,type=integer,minimum=0,maximum=10"`
	Note       string  `json:"note" jsonschema:"title=Note,type=string,enum=C,C#,D,D#,E,F,F#,G,G#,A,A#,B"`
	Frequency  float64 `json:"-" j2sonschema:"title=Frequency,type=number,minimum=0.0,maximum=20000.0"`
	Frequency2 float64 `json:"-" j2sonschema:"title=Frequency,type=number,minimum=0.0,maximum=20000.0"`

	Duration float64 `json:"duration" jsonschema:"title=Duration,description=Duration of the tone,type=number,format=duration"`
	BPM      float64 `json:"bpm" jsonschema:"title=BPM,description=Beats per minute,type=number,minimum=0.0,maximum=1000.0"`
}

type PlayToneRequest struct {
	Tones []Tone `json:"tones" jsonschema:"title=Tones,description=List of frequencies and duration in milliseconds to play"`
}

type PlayPlayableRequest struct {
	Channel uint8 `json:"channel"`
}

type IPlayable interface {
	Play(ctx context.Context, req PlayPlayableRequest) error
}

type SongRegenerateRequest struct {
	Prompt string `json:"prompt"`
}

type ISong interface {
	Regenerate(ctx context.Context, req *SongRegenerateRequest) error
	OnRegenerate(ctx context.Context, req *SongRegenerateRequest) error
}

var PlayableInterface = psi.DefineNodeInterface[IPlayable]()
var SongInterface = psi.DefineNodeInterface[ISong]()
var InstrumentInterface = psi.DefineNodeInterface[IInstrument]()

var SongType = psi.DefineNodeType[*Song](
	psi.WithInterfaceFromNode(PlayableInterface),
	psi.WithInterfaceFromNode(SongInterface),
)
