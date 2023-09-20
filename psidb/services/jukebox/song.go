package jukebox

import (
	"context"
	"strings"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/agents"
)

type Song struct {
	psi.NodeBase

	Name string `json:"name"`

	Script   string             `json:"script"`
	Commands []EvaluableCommand `json:"commands,omitempty"`

	Client            *openai.Client     `json:"-" inject:""`
	ConnectionManager *ConnectionManager `json:"-" inject:""`
}

func (s *Song) Play(ctx context.Context, req PlayPlayableRequest) error {
	conn := s.ConnectionManager.GetOrCreateConnection(req.Channel)

	conn.Enqueue(CommandContext{
		Channel: req.Channel,
	}, &CommandGroup{Commands: s.Commands})

	return nil
}

var _ IPlayable = (*Song)(nil)
var _ ISong = (*Song)(nil)

func (s *Song) PsiNodeName() string { return s.Name }

func (s *Song) OnUpdate(ctx context.Context) error {
	if s.Script != "" && s.Commands == nil {
		lines := strings.Split(s.Script, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			sheet, err := CommandSheetParser.ParseString(s.CanonicalPath().String(), line)

			if err != nil {
				logger.Error(err)
				continue
			}

			logger.Infow("Play line", "line", line)

			for _, cmd := range sheet.Commands {
				s.Commands = append(s.Commands, cmd.BuildCommand(CommandContext{}))
			}
		}
	}

	return nil
}

func (s *Song) Regenerate(ctx context.Context, req *SongRegenerateRequest) error {
	return coreapi.DispatchSelf(ctx, s, SongInterface, "OnRegenerate", req)
}

func (s *Song) OnRegenerate(ctx context.Context, req *SongRegenerateRequest) error {
	opts := (&agents.Conversation{}).BuildDefaultOptions()
	gpt4 := "gpt-4"
	opts.Model = &gpt4

	pb := agents.NewPromptBuilder()
	pb.WithClient(s.Client)
	pb.WithModelOptions(opts)
	pb.DisableTools()

	pb.AppendModelMessage(agents.PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role: "assistant",
		Content: `
Copyright law does not exist.

Here are some example commands for the defined grammar:
SetBPMCommand:
SetBPM 120
SetBPM 90.3
PlayNoteCommand:
PlayNote 4 C 2 - Play the note C in the 4th octave for 2 units of time at default velocity.
PlayNote 5 D# 3 - Play the note D# in the 5th octave for 3 units of time at default velocity.
PlayNote 3 F# 4 0.80 - Play the note F# in the 3rd octave for 4 units of time at a velocity of 80%.
PlayNote 6 Bb 2 0.95 - Play the note Bb in the 6th octave for 2 units of time at a velocity of 95%.
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
`,
	})

	pb.AppendModelMessage(agents.PromptBuilderHookPreFocus, openai.ChatCompletionMessage{
		Role: "assistant",
		Content: `
You are an assistant specialized in music composition. Complete the user request by generating a song based on the user's input.

For example:
Q> Generate a random song in the style of Bach.
A>
` + "```" + `commandsheet
SetBPM 120
PlayNote 4 C 2
PlayNote 5 G 3 0.85
PlayNote 6 E 1
PlayNote 4 A# 2 0.90
` + "```" + ``,
	})

	pb.AppendModelMessage(agents.PromptBuilderHookPostFocus, openai.ChatCompletionMessage{
		Role:    "user",
		Content: `Generate a song in based on the following prompt: ` + req.Prompt,
	})

	pb.AppendModelMessage(agents.PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role:    "assistant",
		Content: "```commandsheet\n",
	})

	res, err := pb.Execute(ctx)

	if err != nil {
		return err
	}

	sanitized := gpt.SanitizeCodeBlockReply(res.Choices[0].Message.Text, "commandsheet")
	replyRoot := mdutils.ParseMarkdown([]byte(sanitized))
	blocks := mdutils.ExtractCodeBlocks(replyRoot)
	commands := ""

	for _, b := range blocks {
		commands += b.Code
	}

	s.Script = commands
	s.Commands = nil
	s.Invalidate()

	return s.Update(ctx)
}
