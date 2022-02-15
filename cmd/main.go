package main

import (
	"log"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/cmd/video"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "nthu-distributed-system [module]",
		Short: "NTHU Distributed System module entrypoints",
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}

	cmd.AddCommand(video.NewVideoCommand())
}
