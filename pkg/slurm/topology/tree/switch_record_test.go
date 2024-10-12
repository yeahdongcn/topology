package tree

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var topologies = []string{
	"../../../../test/topology1.conf",
	"../../../../test/topology2.conf",
	"../../../../test/topology3.conf",
}

func Test__node_name2bitmap(t *testing.T) {
	bitmap, err := _node_name2bitmap("node[1-3,5,7-9]")
	require.NoError(t, err)
	require.Equal(t, &bitstr_t{"node1", "node2", "node3", "node5", "node7", "node8", "node9"}, bitmap)

	bitmap, err = _node_name2bitmap("node[1]")
	require.NoError(t, err)
	require.Equal(t, &bitstr_t{"node1"}, bitmap)

	bitmap, err = _node_name2bitmap("node1")
	require.Error(t, err)
	require.Nil(t, bitmap)

	bitmap, err = _node_name2bitmap("worker[001,134-136,140]")
	require.NoError(t, err)
	require.Equal(t, &bitstr_t{"worker001", "worker134", "worker135", "worker136", "worker140"}, bitmap)
}

func Test__parse_switches(t *testing.T) {
	for i, topo := range topologies {
		f, err := os.Open(topo)
		if err != nil {
			require.NoError(t, err)
		}
		defer f.Close()

		list, err := _parse_switches(f)
		require.NoError(t, err)

		expected := 7
		if i == 1 {
			expected = 8
		} else if i == 2 {
			expected = 24
		}
		require.Equal(t, expected, len(list))
	}
}

func Test__read_topo_file(t *testing.T) {
	for i, topo := range topologies {
		list, err := _read_topo_file(topo)
		require.NoError(t, err)

		expected := 7
		if i == 1 {
			expected = 8
		} else if i == 2 {
			expected = 24
		}
		require.Equal(t, expected, len(list))
	}
}

func Test__merge_switches_array(t *testing.T) {
	a1 := []uint16{1, 2, 3}
	a2 := []uint16{4, 5, 6}
	a3 := _merge_switches_array(a1, a2)
	require.Equal(t, []uint16{1, 2, 3, 4, 5, 6}, a3)

	a1 = []uint16{1, 2, 3}
	a2 = []uint16{3, 4, 5}
	a3 = _merge_switches_array(a1, a2)
	require.Equal(t, []uint16{1, 2, 3, 4, 5}, a3)
}

func Test_switch_record_validate(t *testing.T) {
	for i, topo := range topologies {
		switch_record_table = make([]*switch_record_t, 0)
		switch_record_cnt = 0

		ptr_array, err := _read_topo_file(topo)
		require.NoError(t, err)
		err = switch_record_validate(ptr_array)
		require.NoError(t, err)
		require.NotEmpty(t, switch_record_table)
		expected := 7
		if i == 1 {
			expected = 8
		} else if i == 2 {
			expected = 24
		}
		require.Equal(t, expected, switch_record_cnt)
	}
}
