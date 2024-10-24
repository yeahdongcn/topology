package tree

import (
	"fmt"

	"github.com/yeahdongcn/topology/pkg/slurm"
)

const (
	// DefaultConfigName is the default configuration file name for the topology tree.
	DefaultConfigName = "topology.conf"
	// DefaultConfigPath is the default configuration file path for the topology tree.
	DefaultConfigPath = "/etc/slurm-llnl/" + DefaultConfigName
)

// SwitchRecordValidate validates the switch records from the given configuration file.
func SwitchRecordValidate(filename string) error {
	return switch_record_validate(filename)
}

// EvalNodesTree evaluates the nodes tree.
// It returns the selected nodes, the number of leaf switches, and an error if any.
func EvalNodesTree(availableNodes []string, requiredNodes []string, requestedNodeCount uint32) ([]string, uint16, error) {

	availableNodesInNodeRecordTable := []string{}
	for _, availableNode := range availableNodes {
		if nodeInNodeRecordTable(availableNode, node_record_table) {
			availableNodesInNodeRecordTable = append(availableNodesInNodeRecordTable, availableNode)
		}
	}

	if len(availableNodesInNodeRecordTable) == 0 {
		return nil, 0, nil
	}

	var (
		node_map        *bitstr_t
		req_node_bitmap *bitstr_t
	)
	if len(availableNodesInNodeRecordTable)+len(requiredNodes) > 0 {
		b1 := bitstr_t(availableNodesInNodeRecordTable)
		b2 := bitstr_t(requiredNodes)
		bit_or(&b1, &b2)
		node_map = &b1
	}
	if len(requiredNodes) > 0 {
		bitmap := bitstr_t(requiredNodes)
		req_node_bitmap = &bitmap
	}
	eval := topology_eval_t{
		node_map:        node_map,
		req_node_bitmap: req_node_bitmap,
		req_nodes:       requestedNodeCount,
	}
	if eval_nodes_tree(&eval, false) == slurm.ERROR {
		return nil, 0, fmt.Errorf("failed to evaluate nodes tree")
	}
	return *eval.node_map, eval.leaf_switch_cnt, nil
}
