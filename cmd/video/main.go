package video

import "github.com/spf13/cobra"

func NewVideoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video [service]",
		Short: "start video's service",
	}

	cmd.AddCommand(newAPICommand())
	cmd.AddCommand(newGatewayCommand())
	cmd.AddCommand(newStreamCommand())

	return cmd
}
