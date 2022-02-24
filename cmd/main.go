package main

import (
	"log"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/cmd/comment"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/cmd/video"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "nthu-distributed-system [module]",
		Short: "NTHU Distributed System module entrypoints",
	}

	cmd.AddCommand(video.NewVideoCommand())
	cmd.AddCommand(comment.NewCommentCommand())

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
