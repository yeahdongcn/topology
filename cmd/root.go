package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/yeahdongcn/topology/pkg/slurm/topology/tree"
)

var (
	topology       string
	availableNodes []string
	requiredNodes  []string
	requested      uint32
	rootCmd        = &cobra.Command{
		Use: "topology",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("Topology configuration file: %s", topology)
			log.Debugf("Available nodes: %#v", availableNodes)
			log.Debugf("Required nodes: %#v", requiredNodes)
			log.Debugf("Number of nodes requested: %d", requested)

			err := tree.SwitchRecordValidate(topology)
			if err != nil {
				return err
			}
			selectedNodes, leafSwitchCount, err := tree.EvalNodesTree(availableNodes, requiredNodes, requested)
			if err != nil {
				return err
			}

			log.Info("Selected nodes: ", selectedNodes)
			log.Info("Leaf switch count: ", leafSwitchCount)
			return nil
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&topology, "topology", "p", "", "Path to the topology configuration file")
	rootCmd.Flags().StringArrayVarP(&availableNodes, "available-nodes", "a", []string{}, "List of available nodes")
	rootCmd.Flags().StringArrayVarP(&requiredNodes, "required-nodes", "r", []string{}, "List of required nodes")
	rootCmd.Flags().Uint32VarP(&requested, "requested-node-count", "c", 0, "Number of nodes requested")
	rootCmd.MarkFlagRequired("topology")
	rootCmd.MarkFlagRequired("available-nodes")
	rootCmd.MarkFlagRequired("requested-node-count")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
