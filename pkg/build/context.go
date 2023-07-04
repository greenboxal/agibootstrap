package build

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
)

type Context struct {
	builder *Builder

	log     *Log
	project project.Project
	cfg     Configuration

	totalEpochs  int
	totalSteps   int
	totalChanges int

	errors []error
}

func (bctx *Context) Log() *Log                { return bctx.Log() }
func (bctx *Context) Project() project.Project { return bctx.project }
func (bctx *Context) Config() Configuration    { return bctx.cfg }

func (bctx *Context) ReportError(err error) {
	bctx.errors = append(bctx.errors, err)

	bctx.log.Error(err)
}

func (bctx *Context) Build(ctx context.Context) (*Result, error) {
	err := bctx.runBuild(ctx)

	if err != nil {
		bctx.ReportError(err)
	}

	// Return the total number of totalChanges and no error
	return &Result{
		Errors:      bctx.errors,
		ChangeCount: bctx.totalChanges,
		TotalEpochs: bctx.totalEpochs,
		TotalSteps:  bctx.totalSteps,
	}, nil
}

func (bctx *Context) runBuild(ctx context.Context) error {
	// Execute the build steps until no further totalChanges are made
	for ; bctx.cfg.MaxEpochs == 0 || bctx.totalEpochs < bctx.cfg.MaxEpochs; bctx.totalEpochs++ {
		stepChanges := 0

		// Execute each build step
		for _, step := range bctx.cfg.BuildSteps {
			if bctx.cfg.MaxSteps != 0 && bctx.totalSteps >= bctx.cfg.MaxSteps {
				break
			}

			var result StepResult

			// Wrap the build step execution in a recover function
			task := bctx.project.TaskManager().SpawnTask(ctx, func(progress tasks.TaskProgress) (err error) {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(error); ok {
							err = e
						} else {
							err = fmt.Errorf("%v", r)
						}
					}
				}()

				// Execute the build step and return the result
				result, err = step.Process(ctx, bctx)

				return
			})

			if err := task.Error(); err != nil {
				return err
			}

			// Update the total number of totalChanges
			stepChanges += result.ChangeCount
			bctx.totalSteps++
		}

		// Stage all the totalChanges in the file system
		if err := bctx.project.FS().StageAll(); err != nil {
			return err
		}

		// Commit the totalChanges to the file system
		if err := bctx.project.Commit(); err != nil {
			return err
		}

		// If no totalChanges were made, exit the loop
		if stepChanges == 0 {
			break
		}

		// Update the total number of totalChanges
		bctx.totalChanges += stepChanges

		if bctx.cfg.MaxSteps != 0 && bctx.totalSteps >= bctx.cfg.MaxSteps {
			break
		}
	}

	// Push the totalChanges to the remote repository
	if err := bctx.project.FS().Push(); err != nil {
		return err
	}

	return nil
}

func (bctx *Context) Close() {
	if bctx.log != nil {
		_ = bctx.log.Close()
	}
}
