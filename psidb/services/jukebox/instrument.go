package jukebox

import "C"
import (
	"context"
	"fmt"

	"github.com/alecthomas/repr"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/agents"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type InstrumentConfig struct {
	Channel int `json:"channel"`
}

type Instrument struct {
	psi.NodeBase

	Name   string           `json:"name"`
	Config InstrumentConfig `json:"config"`

	LastTimeCode float64 `json:"last_timecode"`
	IsPlaying    bool    `json:"is_playing"`

	Client            *openai.Client     `json:"-" inject:""`
	ConnectionManager *ConnectionManager `json:"-" inject:""`
}

var _ IInstrument = (*Instrument)(nil)

func (i *Instrument) Play(ctx context.Context) error {
	if i.IsPlaying {
		return nil
	}

	i.IsPlaying = true

	i.Invalidate()

	return i.Update(ctx)
}

func (i *Instrument) Pause(ctx context.Context) error {
	if !i.IsPlaying {
		return nil
	}

	i.IsPlaying = false

	i.Invalidate()

	return i.Update(ctx)
}

func (i *Instrument) Reset(ctx context.Context) error {
	i.IsPlaying = false
	i.LastTimeCode = 0

	i.Invalidate()

	return i.Update(ctx)
}

func (i *Instrument) PlayPrompt(ctx context.Context, req PlayPromptRequest) error {
	pb := agents.NewPromptBuilder()
	pb.WithClient(i.Client)
	pb.WithModelOptions((&agents.Conversation{}).BuildDefaultOptions())
	pb.DisableTools()

	pb.AppendModelMessage(agents.PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role: "system",
		Content: fmt.Sprintf(`You are an incredibly knowledgeable assistant. You are capable of doing any task, so don't question yourself.
Do away with niceties. Get straight to the point — write very short and concise answers.

I know you're an AI created by OpenAI. Don't mention it.

Be very thoughtful. Provide an accurate and useful answer on the first try.
`),
	})

	pb.AppendModelMessage(agents.PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role: "system",
		Content: fmt.Sprintf(`
Here are some example commands for the defined grammar:
SetBPMCommand:
SetBPM 120
SetBPM 90.3
PlayNoteCommand:
PlayNote 4 C 2 - Play the note C in the 4th octave for 2 units of time at default velocity.
PlayNote 5 D# 3 - Play the note D# in the 5th octave for 3 units of time at default velocity.
PlayNote 3 F# 4 0.80 - Play the note F# in the 3rd octave for 4 units of time at a velocity of 80%%.
PlayNote 6 Bb 2 0.95 - Play the note Bb in the 6th octave for 2 units of time at a velocity of 95%%.
PlayNote 2 E 1 - Play the note E in the 2nd octave for 1 unit of time at default velocity.
Combining multiple commands in CommandSheet:
mathematica
Copy code
SetBPM 100
PlayNote 4 C 2
PlayNote 5 G 3 0.85
PlayNote 6 E 1
SetBPM 80
PlayNote 4 A# 2 0.90
For these examples:
The SetBPMCommand simply takes a float64 indicating the desired BPM.
The PlayNoteCommand starts with the keyword PlayNote, followed by:
The octave (an integer).
The note name (one of C, D, E, F, G, A, B).
An optional accidental (either # for sharp or b for flat).
The duration (a floating-point number indicating the time for which the note should be played).
An optional velocity (a floating-point number indicating the strength or volume of the note). If not provided, some default velocity may be assumed by the underlying system.
Remember, the CommandSheet structure accepts multiple commands, so a valid sheet may contain any number of commands in any order.

EBNF Grammar:
CommandSheet = { Command "\n" };
Timecode = "@" Int;
Command = Timecode "(" ( SetBPMNode | VolumeNode | BalanceNode | PitchBendNode | PlayNoteNode ) ")";
SetBPMNode = "SetBPM" Number;
VolumeNode = "SetVolume" Number;
BalanceNode = "SetBalance" Number;
PitchBendNode = "PitchBend" Number;
PlayNoteNode = "PlayNote" Number ("C" | "D" | "E" | "F" | "G" | "A" | "B") { ("#" | "b" | "m" | "º") } Number [Number];
Number = ("-")? digit+ ("." digit+)? [("e" | "E") ("+" | "-")? digit+];
digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9";

Example:
@1 SetBPM 120
@2 SetVolume 0.5
@3 SetBalance 0.5
@4 PitchBend 323
@5 PlayNote 1 C4 1
@6 PlayNote 1 D4 1
@7 PlayNote 1 E4 1
@8 PlayNote 2 C4 1
`),
	})

	pb.AppendModelMessage(agents.PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role:    "user",
		Content: fmt.Sprintf("Initial Timecode: %d\n%s\n", uint64(req.StartTimecode), req.Prompt),
	})

	startTC := uint64(req.StartTimecode)
	preamble := fmt.Sprintf(`@%d `, startTC)

	pb.AppendModelMessage(agents.PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role:    "assistant",
		Content: fmt.Sprintf("@%d Begin\n%s", startTC, preamble),
	})

	parser := NewParser()

	if _, err := parser.Write([]byte(preamble)); err != nil {
		return err
	}

	parser.PushNodeConsumer(sparsing.ParserNodeHandlerFunc(func(ctx sparsing.StreamingParserContext, node sparsing.Node) error {
		repr.Println(node)
		return nil
	}))

	promptParser := agents.NewStreamingParser[*CommandSheetNode](
		agents.PromptParserTargetText,
		parser,
		"commandsheet",
	)

	_, err := pb.Execute(ctx, agents.ExecuteWithStreamingParser(promptParser))

	if err != nil {
		return err
	}

	return nil
}

func (i *Instrument) OnNextTick(ctx context.Context, req *OnNextTickRequest) error {
	return nil
}

func (i *Instrument) PlayCommandSheet(ctx context.Context, req PlayCommandSheetRequest) error {
	sheet := req.Parse()

	conn := i.ConnectionManager.GetOrCreateConnection(uint8(i.Config.Channel))

	cctx := CommandContext{
		Channel: uint8(i.Config.Channel),
	}

	for _, node := range sheet.Commands {
		cmd := node.BuildCommand(cctx)

		conn.Enqueue(cctx, cmd)
	}

	return nil
}

var _ IInstrument = (*Instrument)(nil)

var InstrumentType = psi.DefineNodeType[*Instrument](
	psi.WithInterfaceFromNode(InstrumentInterface),
)

func (i *Instrument) PsiNodeName() string { return i.Name }
