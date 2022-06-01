package comment

import "github.com/spf13/cobra"

func NewCommentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment [service]",
		Short: "start comment's service",
	}

	cmd.AddCommand(newAPICommand())
	cmd.AddCommand(newGatewayCommand())
	cmd.AddCommand(newMigrationCommand())

	return cmd
}
