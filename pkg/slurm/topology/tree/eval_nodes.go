package tree

type topo_weight_info_t struct {
	node_bitmap *bitstr_t
	node_cnt    int
	weight      uint64
}
