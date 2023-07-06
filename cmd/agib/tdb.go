package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func buildPsiDbCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "psidb",
		Short: "PSI Database Management",
		Long:  "This command manages the PSI database.",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeProject(cmd.Context(), true)
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return teardownProject()
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list [path]",
		Short: "List PSI nodes",
		Long:  "This command lists PSI nodes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var rootPath psi.Path

			cmd.SilenceUsage = true

			if len(args) > 0 {
				rootPath, err = psi.ParsePath(args[0])

				if err != nil {
					return err
				}
			} else {
				rootPath = project.CanonicalPath()
			}

			g := project.Graph()
			rootNode, err := g.ResolveNode(cmd.Context(), rootPath)

			if err != nil {
				return err
			}

			fmt.Printf("Node: %s\n", rootNode)

			fmt.Printf("\nChildren:\n")
			for it := rootNode.ChildrenIterator(); it.Next(); {
				child := it.Value()

				cmd.Printf("* %s\n", child.CanonicalPath())
			}

			fmt.Printf("\nEdges:\n")
			for it := rootNode.Edges(); it.Next(); {
				edge := it.Edge()

				cmd.Printf("* %s -> %s\n", edge.Key(), edge.To().CanonicalPath())
			}

			return nil
		},
	}

	rootCmd.AddCommand(listCmd)

	return rootCmd
}
