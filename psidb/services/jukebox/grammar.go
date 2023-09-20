package jukebox

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type CommandNode interface {
	BuildCommand(ctx CommandContext) EvaluableCommand
}

type CommandSheet struct {
	Commands []*Command `(@@ ("\n" @@)*)`
}

type Command struct {
	Timecode float64 `("@" @Number)? (`

	SetBPM    *SetBPMNode    `@@|`
	Volume    *PlayNoteNode  `@@|`
	Balance   *BalanceNode   `@@|`
	PitchBend *PitchBendNode `@@|`
	PlayNote  *PlayNoteNode  `@@)`
}

func (cmd *Command) BuildCommand(ctx CommandContext) EvaluableCommand {
	if cmd.SetBPM != nil {
		return cmd.SetBPM.BuildCommand(ctx)
	} else if cmd.Volume != nil {
		return cmd.Volume.BuildCommand(ctx)
	} else if cmd.Balance != nil {
		return cmd.Balance.BuildCommand(ctx)
	} else if cmd.PitchBend != nil {
		return cmd.PitchBend.BuildCommand(ctx)
	} else if cmd.PlayNote != nil {
		return cmd.PlayNote.BuildCommand(ctx)
	} else {
		panic("unknown command")
	}
}

type VolumeNode struct {
	Volume float64 `"SetVolume" @Number`
}

func (v *VolumeNode) BuildCommand(ctx CommandContext) EvaluableCommand {
	return &VolumeCommand{
		Volume: v.Volume,
	}
}

type BalanceNode struct {
	Balance float64 `"SetBalance" @Number`
}

func (b *BalanceNode) BuildCommand(ctx CommandContext) EvaluableCommand {
	return &BalanceCommand{
		Balance: b.Balance,
	}
}

type PitchBendNode struct {
	PitchBend float64 `"PitchBend" @Number`
}

func (p *PitchBendNode) BuildCommand(ctx CommandContext) EvaluableCommand {
	return &PitchBendCommand{
		PitchBend: p.PitchBend,
	}
}

type SetBPMNode struct {
	BPM float64 `"SetBPM" @Number`
}

func (s *SetBPMNode) BuildCommand(ctx CommandContext) EvaluableCommand {
	return &SetBPMCommand{
		BPM: s.BPM,
	}
}

type PlayNoteNode struct {
	Octave      int      `"PlayNote" @Number`
	Note        string   `@("C"|"D"|"E"|"F"|"G"|"A"|"B")`
	Accidentals []string `(@("#"|"b"|"m"|"º"))*`
	Duration    float64  `@Number`
	Velocity    *float64 `(@Number)?`
}

func (n *PlayNoteNode) BuildCommand(ctx CommandContext) EvaluableCommand {
	return &NoteCommand{
		Octave:      n.Octave,
		Note:        n.Note,
		Accidentals: n.Accidentals,
		Duration:    n.Duration,
		Velocity:    n.Velocity,
	}
}

var CommandSheetLexer = lexer.MustSimple([]lexer.SimpleRule{
	{`Symbol`, `([#bm@°]|C|D|E|F|G|A|B|SetBPM|Note|PlayNote|PitchBend|SetVolume|SetBalance)`},
	{"Number", `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
	{"Int", `[-+]?\d+?`},
	{"whitespace", `\s+`},
})

var CommandSheetParser = participle.MustBuild[CommandSheet](
	participle.Lexer(CommandSheetLexer),
	participle.UseLookahead(participle.MaxLookahead),
	participle.Elide("whitespace"),
)
