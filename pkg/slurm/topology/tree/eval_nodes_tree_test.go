package tree

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yeahdongcn/topology/pkg/slurm"
)

func Benchmark_eval_nodes_tree_topo3(b *testing.B) {
	err := switch_record_validate("../../../../test/topology3.conf")
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		avns := []string{}
		for i := 0; i < 80; i++ {
			x := rand.Intn(100)
			avns = append(avns, "worker"+fmt.Sprintf("%03d", x))
		}
		node_map := bitstr_t(avns)
		eval := &topology_eval_t{
			node_map:  &node_map,
			req_nodes: 4,
		}
		err := eval_nodes_tree(eval, false)
		require.Equal(b, slurm.SUCCESS, err)
	}
}

func Test_eval_nodes_tree_topo3(t *testing.T) {
	err := switch_record_validate("../../../../test/topology3.conf")
	require.NoError(t, err)

	node_map := &bitstr_t{"worker001", "worker002", "worker003", "worker004", "worker005", "worker006", "worker007"}
	eval := &topology_eval_t{
		node_map:  node_map,
		req_nodes: 4,
	}
	rc := eval_nodes_tree(eval, false)
	require.Equal(t, slurm.SUCCESS, rc)
	require.Equal(t, &bitstr_t{"worker001", "worker002", "worker003", "worker004"}, eval.node_map)
	require.Equal(t, uint16(2), eval.leaf_switch_cnt)
}

func Test_eval_nodes_tree_topo2(t *testing.T) {
	err := switch_record_validate("../../../../test/topology2.conf")
	require.NoError(t, err)

	node_map := &bitstr_t{"tux0", "tux1", "tux2", "tux12", "tux13", "tux14", "tux15"}
	eval := &topology_eval_t{
		node_map:  node_map,
		req_nodes: 4,
	}
	rc := eval_nodes_tree(eval, false)
	require.Equal(t, slurm.SUCCESS, rc)
	require.Equal(t, &bitstr_t{"tux12", "tux13", "tux14", "tux15"}, eval.node_map)
	require.Equal(t, uint16(1), eval.leaf_switch_cnt)
}

func Test_eval_nodes_tree_topo1(t *testing.T) {
	err := switch_record_validate("../../../../test/topology1.conf")
	require.NoError(t, err)

	node_map := &bitstr_t{"tu-x1", "tu-x3", "tux5", "tux6", "tux7"}
	eval := &topology_eval_t{
		node_map:  node_map,
		req_nodes: 3,
	}
	rc := eval_nodes_tree(eval, false)
	require.Equal(t, slurm.SUCCESS, rc)
	require.Equal(t, &bitstr_t{"tux5", "tux6", "tux7"}, eval.node_map)
	require.Equal(t, uint16(2), eval.leaf_switch_cnt)

	node_map = &bitstr_t{"tu-x1"}
	eval = &topology_eval_t{
		node_map:  node_map,
		req_nodes: 3,
	}
	rc = eval_nodes_tree(eval, false)
	require.Equal(t, slurm.ERROR, rc)

	node_map = &bitstr_t{"tu-x0", "tu-x2", "tux4", "tux7"}
	eval = &topology_eval_t{
		node_map:  node_map,
		req_nodes: 3,
	}
	rc = eval_nodes_tree(eval, false)
	require.Equal(t, slurm.SUCCESS, rc)
	require.Equal(t, &bitstr_t{"tu-x0", "tu-x2", "tux4"}, eval.node_map)
	require.Equal(t, uint16(3), eval.leaf_switch_cnt)

	// With req_node_bitmap
	node_map = &bitstr_t{"tu-x0", "tu-x2", "tux4", "tux7"}
	eval = &topology_eval_t{
		node_map:        node_map,
		req_node_bitmap: &bitstr_t{"tu-x0", "tu-x2", "tux4"},
		req_nodes:       3,
	}
	rc = eval_nodes_tree(eval, false)
	require.Equal(t, slurm.SUCCESS, rc)
	require.Equal(t, &bitstr_t{"tu-x0", "tu-x2", "tux4"}, eval.node_map)
	require.Equal(t, uint16(3), eval.leaf_switch_cnt)

	node_map = &bitstr_t{"tu-x0", "tu-x1", "tu-x2", "tux4", "tux5", "tux6"}
	eval = &topology_eval_t{
		node_map:        node_map,
		req_node_bitmap: &bitstr_t{"tux4"},
		req_nodes:       3,
	}
	rc = eval_nodes_tree(eval, false)
	require.Equal(t, slurm.SUCCESS, rc)
	require.Equal(t, &bitstr_t{"tux4", "tux5", "tux6"}, eval.node_map)
	require.Equal(t, uint16(2), eval.leaf_switch_cnt)
}
