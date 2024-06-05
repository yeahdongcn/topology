package tree

type topology_eval_t struct {
	node_map        *bitstr_t /* available/selected nodes */
	req_nodes       uint32    /* number of requested nodes */
	leaf_switch_cnt uint16    /* number of leaf switches */
	// XXX: Originally from job_record_t
	req_node_bitmap *bitstr_t /* bitmap of required nodes */
}
