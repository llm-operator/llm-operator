package batch

import (
	"github.com/llmariner/llmariner/cli/internal/batch/jobs"
	"github.com/spf13/cobra"
)

// Cmd is the root command for batch.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "batch",
		Short:              "batch commands",
		Args:               cobra.NoArgs,
		DisableFlagParsing: true,
	}
	cmd.AddCommand(jobs.Cmd())
	return cmd
}
