package tree

func nodeInNodeRecordTable(node string, node_record_table []*node_record_t) bool {
	for _, n := range node_record_table {
		if n.name == node {
			return true
		}
	}
	return false

}
