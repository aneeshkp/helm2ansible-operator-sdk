package main

import (
	"os"

	"github.com/redhat-nfvpe/helm2ansible-operator-sdk/cmd/helm2ansible-operator-sdk/new"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "helm2ansible-operator-sdk",
		Short: "A Kit to Convert Helm Chart Operators to Go Operators, Yee-Haw! 🏇",
	}
	root.AddCommand(new.GetNewCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
