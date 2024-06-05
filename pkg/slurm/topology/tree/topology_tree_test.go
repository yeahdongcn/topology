package tree

import (
	"testing"
)

func Test_topology_p_generate_node_ranking(t *testing.T) {
	for _, topo := range topologies {
		node_record_table = make([]*node_record_t, 0)
		node_record_cnt = 0

		topology_p_generate_node_ranking(topo)
	}
}
