package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	fusefs "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/spf13/cobra"

	"github.com/greenboxal/agibootstrap/pkg/platform/psifuse"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func buildPsiDbCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "psidb",
		Short: "PSI Database Management",
		Long:  "This command manages the PSI database.",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeProject(cmd.Context(), true)
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return teardownProject(cmd.Context())
		},
	}

	listCmd := &cobra.Command{
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
			edges, err := g.Store().ListNodeEdges(cmd.Context(), rootPath)

			if err != nil {
				return err
			}

			for edges.Next() {
				edge := edges.Value()

				if edge.Key.Kind == psi.EdgeKindChild {
					cmd.Printf("%s\n", edge.Key)
				} else {
					if edge.ToPath != nil {
						cmd.Printf("%s -> %s\n", edge.Key, *edge.ToPath)
					} else if edge.ToLink != nil {
						cmd.Printf("%s -> %s\n", edge.Key, edge.ToLink)
					} else {
						cmd.Printf("%s -> ?\n", edge.Key)
					}
				}
			}

			return nil
		},
	}

	dumpFrozenNode := &cobra.Command{
		Use:   "dump-frozen-node [path]",
		Short: "Dumps PSI frozen node",
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
			fn, _, err := g.Store().GetNodeByPath(cmd.Context(), rootPath)

			if err != nil {
				return err
			}

			fmt.Printf("%s\n", dumpJson(fn))

			return nil
		},
	}

	dumpFrozenEdge := &cobra.Command{
		Use:   "dump-frozen-edge [cid]",
		Short: "Dumps PSI frozen edge",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			contentId, err := cid.Parse(args[0])

			if err != nil {
				return err
			}

			cmd.SilenceUsage = true

			g := project.Graph()
			fe, err := g.Store().GetEdgeByCid(cmd.Context(), cidlink.Link{Cid: contentId})

			if err != nil {
				return err
			}

			fmt.Printf("%s\n", dumpJson(fe))

			return nil
		},
	}

	mountFuse := &cobra.Command{
		Use:   "fuse [dir]",
		Short: "Mounts fuse filesystem",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if _, err := os.Stat(args[0]); err != nil {
				return err
			}

			root := psifuse.NewPsiNodeDir(project.Graph(), project.Graph().Root().CanonicalPath())

			_ = exec.Command("diskutil", "unmount", "force", args[0]).Run()

			server, err := fusefs.Mount(args[0], root, &fusefs.Options{
				MountOptions: fuse.MountOptions{
					// Set to true to see how the file system works.
					Debug: true,
				},
			})

			if err != nil {

				return err
			}

			defer server.Unmount()

			cmd.Printf("Mounted at %s\n", args[0])
			cmd.Printf("Unmount by running: fusermount -u %s\n", args[0])

			server.Wait()

			return nil
		},
	}

	rootCmd.AddCommand(listCmd, dumpFrozenNode, dumpFrozenEdge, mountFuse)

	return rootCmd
}

func dumpJson(v any) string {
	var parsed any

	encoded, err := ipld.Encode(typesystem.Wrap(v), dagjson.Encode)

	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(encoded, &parsed); err != nil {
		panic(err)
	}

	encoded, err = json.MarshalIndent(parsed, "", "  ")

	if err != nil {
		panic(err)
	}

	return string(encoded)
}
