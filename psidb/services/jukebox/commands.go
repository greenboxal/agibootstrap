package jukebox

import (
	"math"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

const MasterChannel = 0
const FirstInstrumentChannel = 1

type CommandContext struct {
	Channel uint8   `json:"channel"`
	BPM     float64 `json:"BPM"`

	Out drivers.Out `json:"-"`
}

func (ctx *CommandContext) SendMidi(change midi.Message) error {
	return ctx.Out.Send(change)
}

type EvaluableCommand interface {
	Evaluate(ctx *CommandContext) error
}

type CommandGroup struct {
	Commands []EvaluableCommand `json:"commands"`
}

func (c *CommandGroup) Evaluate(ctx *CommandContext) error {
	for _, cmd := range c.Commands {
		if err := cmd.Evaluate(ctx); err != nil {
			return err
		}
	}

	return nil
}

type SetBPMCommand struct {
	BPM float64 `json:"bpm"`
}

func (s *SetBPMCommand) Evaluate(ctx *CommandContext) error {
	ctx.BPM = s.BPM

	ubpm := uint8((s.BPM - 20) / 979 * 127)

	if err := ctx.Out.Send(midi.ControlChange(MasterChannel, 14, ubpm)); err != nil {
		logger.Error(err)
	}

	return nil
}

type NoteCommand struct {
	Octave      int      `json:"octave"`
	Note        string   `json:"note"`
	Accidentals []string `json:"accidentals,omitempty"`
	Duration    float64  `json:"duration"`
	Velocity    *float64 `json:"velocity,omitempty"`
}

func (n *NoteCommand) Evaluate(ctx *CommandContext) error {
	var velocity float64

	if n.Velocity != nil {
		velocity = *n.Velocity
	} else {
		velocity = 1.0
	}

	octave := n.Octave
	cNote := midiNoteMap["C"]
	nNote := midiNoteMap[n.Note]
	semiTone := nNote - cNote

	for _, acc := range n.Accidentals {
		if acc == "#" {
			semiTone++
		} else if acc == "b" {
			semiTone--
		}
	}

	base := semiTone % cNote
	octave += int(math.Floor(float64(semiTone) / float64(cNote)))
	midiKey := shiftKey(uint8(base), uint8(octave))

	stepDuration := time.Duration((60/ctx.BPM)*1000.0) * time.Millisecond
	budget := stepDuration * time.Duration(n.Duration)
	deadline := time.Now().Add(budget)
	counter := 0

	defer func() {
		if err := ctx.SendMidi(midi.NoteOff(ctx.Channel, midiKey)); err != nil {
			logger.Error(err)
		}
	}()

	for time.Now().Before(deadline) {
		duration := stepDuration

		if duration > budget {
			duration = budget
		}

		v := uint8(velocity * (float64(duration) / float64(stepDuration)) * 127.0)

		if v >= 127 {
			v = 127
		}

		if err := ctx.SendMidi(midi.NoteOn(ctx.Channel, midiKey, v/3*2)); err != nil {
			logger.Error(err)
		}

		time.Sleep(duration)

		if err := ctx.SendMidi(midi.NoteOffVelocity(ctx.Channel, midiKey, v/3)); err != nil {
			logger.Error(err)
		}

		counter++
	}

	return nil
}

type VolumeCommand struct {
	Volume float64 `json:"volume"`
}

func (v *VolumeCommand) Evaluate(ctx *CommandContext) error {
	return ctx.SendMidi(midi.ControlChange(ctx.Channel, 7, uint8(v.Volume*127)))
}

type BalanceCommand struct {
	Balance float64 `json:"balance"`
}

func (b *BalanceCommand) Evaluate(ctx *CommandContext) error {
	return ctx.SendMidi(midi.ControlChange(ctx.Channel, 8, uint8(b.Balance*127)))
}

type PitchBendCommand struct {
	PitchBend float64 `json:"pitch_bend"`
}

func (p *PitchBendCommand) Evaluate(ctx *CommandContext) error {
	return ctx.SendMidi(midi.Pitchbend(0, int16(uint16(p.PitchBend*16383))))
}
