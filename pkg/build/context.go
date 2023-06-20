package build

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/project"
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

		// Sync the project to ensure it is up to date
		if err := bctx.project.Sync(); err != nil {
			return err
		}

		// Execute each build step
		for _, step := range bctx.cfg.BuildSteps {
			if bctx.cfg.MaxSteps != 0 && bctx.totalSteps >= bctx.cfg.MaxSteps {
				break
			}

			// Wrap the build step execution in a recover function
			processWrapped := func() (result StepResult, err error) {
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
				return step.Process(ctx, bctx)
			}

			// Execute the build step and handle any errors
			result, err := processWrapped()
			if err != nil {
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
		_ = bctx.log.Sync()
	}
}
