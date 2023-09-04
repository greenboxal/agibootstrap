package console

import (
	`bytes`

	`github.com/jbenet/goprocess`
	`github.com/openengineer/go-repl`
	`github.com/spf13/cobra`

	`github.com/mattn/go-shellwords`

	`github.com/greenboxal/agibootstrap/psidb/core/session`
)

type Console struct {
	repl    *repl.Repl
	proc    goprocess.Process
	session *session.Session

	rootCommand *cobra.Command
}

func NewConsole(session *session.Session) *Console {
	c := &Console{}

	c.session = session
	c.repl = repl.NewRepl(c)
	c.proc = goprocess.Go(c.Run)

	c.rootCommand = newRootCommand(c, session)

	return c
}

func (c *Console) Run(proc goprocess.Process) {
	go func() {
		select {
		case <-proc.Closing():
		case <-c.session.Closing():
		}

		c.repl.UnmakeRaw()
	}()

	if err := c.repl.Loop(); err != nil {
		panic(err)
	}
}

func (c *Console) Prompt() string {
	return "> "
}

func (c *Console) Eval(buffer string) string {
	words, err := shellwords.Parse(buffer)

	if err != nil {
		return err.Error()
	}

	if len(words) == 0 {
		return ""
	}

	result, err := c.runCommand(words)

	if err != nil {
		return err.Error()
	}

	return result
}

func (c *Console) Tab(buffer string) string {
	return ""
}

func (c *Console) Close() error {
	return c.proc.Close()
}

func (c *Console) runCommand(argv []string) (string, error) {
	out := &bytes.Buffer{}

	c.rootCommand.SetArgs(argv)
	c.rootCommand.SetErr(out)
	c.rootCommand.SetOut(out)

	if err := c.rootCommand.Execute(); err != nil {
		return "", err
	}

	return out.String(), nil
}

func newRootCommand(con *Console, sess *session.Session) *cobra.Command {
	root := &cobra.Command{
		Use: "psicli",

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.OutOrStdout().Write([]byte("Hello, world!\n"))
			return nil
		},
	}

	root.AddCommand(&cobra.Command{
		Use: "ps",

		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use: "help",

		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	})

	root.AddCommand(&cobra.Command{
		Use: "exit",

		RunE: func(cmd *cobra.Command, args []string) error {
			return con.Close()
		},
	})

	return root
}
