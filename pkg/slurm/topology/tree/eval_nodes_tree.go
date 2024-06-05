package tree

import (
	"container/list"

	log "github.com/sirupsen/logrus"

	"github.com/yeahdongcn/topology/pkg/slurm"
)

func eval_nodes_enough_nodes(avail_nodes, rem_nodes int) bool {
	return avail_nodes >= rem_nodes
}

func _topo_add_dist(dist *[]uint32, inx int) {
	for i := 0; i < switch_record_cnt; i++ {
		if switch_record_table[inx].switches_dist[i] == INFINITE ||
			(*dist)[i] == INFINITE {
			(*dist)[i] = INFINITE
		} else {
			(*dist)[i] += switch_record_table[inx].switches_dist[i]
		}
	}
}

func _topo_compare_switches(i, j uint16, switch_node_cnt *[]int, rem_nodes int) int {
	for {
		i_fit := (*switch_node_cnt)[i] >= rem_nodes
		j_fit := (*switch_node_cnt)[j] >= rem_nodes
		if i_fit && j_fit {
			if (*switch_node_cnt)[i] < (*switch_node_cnt)[j] {
				return 1
			}
			if (*switch_node_cnt)[i] > (*switch_node_cnt)[j] {
				return -1
			}
			break
		} else if i_fit {
			return 1
		} else if j_fit {
			return -1
		}

		if ((switch_record_table[i].parent != i) ||
			(switch_record_table[j].parent != j)) &&
			(switch_record_table[i].parent !=
				switch_record_table[j].parent) {
			i = switch_record_table[i].parent
			j = switch_record_table[j].parent
			continue
		}

		break
	}

	if (*switch_node_cnt)[i] > (*switch_node_cnt)[j] {
		return 1
	}
	if (*switch_node_cnt)[i] < (*switch_node_cnt)[j] {
		return -1
	}
	if switch_record_table[i].level < switch_record_table[j].level {
		return 1
	}
	if switch_record_table[i].level > switch_record_table[j].level {
		return -1
	}
	return 0
}

func _topo_choose_best_switch(dist *[]uint32, switch_node_cnt *[]int, rem_nodes int, i int, best_switch *int) {
	if *best_switch == -1 || (*dist)[i] == INFINITE || (*switch_node_cnt)[i] == 0 {
		/*
		 * If first possibility
		 */
		if (*switch_node_cnt)[i] > 0 && (*dist)[i] < INFINITE {
			*best_switch = i
		}
		return
	}

	tcs := _topo_compare_switches(uint16(i), uint16(*best_switch), switch_node_cnt, rem_nodes)
	if ((*dist)[i] < (*dist)[*best_switch] && tcs >= 0) ||
		((*dist)[i] == (*dist)[*best_switch] && tcs > 0) {
		/*
		 * If closer and fit request OR
		 * same distance and tightest fit (less resource waste)
		 */
		*best_switch = i
	}
}

/* Allocate resources to job using a minimal leaf switch count */
func _eval_nodes_topo(topo_eval *topology_eval_t) int {
	var (
		switch_node_bitmap []*bitstr_t /* nodes on this switch */
		switch_node_cnt    []int       /* total nodes on switch */
		switch_required    []int       /* set if has required node */
		req_nodes_bitmap   *bitstr_t   /* required node bitmap */
		best_nodes_bitmap  *bitstr_t   /* required+low priority nodes */
		best_node_cnt      = 0
		node_weight_list   *list.List
		rc                 = slurm.SUCCESS
		top_switch_inx     = -1
		rem_nodes          = 0 /* remaining resources desired */
		prev_rem_nodes     = 0
		switches_dist      []uint32
		sufficient         = false
	)

	rem_nodes = int(topo_eval.req_nodes)

	/* Validate availability of required nodes */
	if topo_eval.req_node_bitmap != nil {
		if !bit_super_set(topo_eval.req_node_bitmap, topo_eval.node_map) {
			log.Error("requires nodes which are not currently available")
			rc = slurm.ERROR
			goto fini
		}

		req_node_cnt := bit_set_count(topo_eval.req_node_bitmap)
		if req_node_cnt == 0 {
			log.Error("required node list has no nodes")
			rc = slurm.ERROR
			goto fini
		}

		max_nodes := bit_set_count(topo_eval.node_map)
		if req_node_cnt > max_nodes {
			log.Errorf("requires more nodes than currently available (%d>%d)",
				req_node_cnt, max_nodes)
			rc = slurm.ERROR
			goto fini
		}

		req_nodes_bitmap = topo_eval.req_node_bitmap
	}

	/*
	 * Add required nodes to job allocation and
	 * build list of node bitmaps, sorted by weight
	 */
	if bit_set_count(topo_eval.node_map) == 0 {
		log.Error("node_map is empty")
		rc = slurm.ERROR
		goto fini
	}
	node_weight_list = list.New()
	for i := 0; i < bit_set_count(topo_eval.node_map); i++ {
		node_ptr := (*topo_eval.node_map)[i]
		if req_nodes_bitmap != nil && bit_test(req_nodes_bitmap, node_ptr) {
			rem_nodes--
		}

		var nw *topo_weight_info_t
		if node_weight_list.Front() == nil {
			nw = &topo_weight_info_t{
				node_bitmap: &bitstr_t{},
				node_cnt:    0,
				weight:      0,
			}

			node_weight_list.PushBack(nw)
		} else {
			nw = node_weight_list.Front().Value.(*topo_weight_info_t)
		}
		bit_set(nw.node_bitmap, node_ptr)
		nw.node_cnt++
	}

	/*
	 * Identify the highest level switch to be used.
	 * Note that nodes can be on multiple non-overlapping switches.
	 */
	switch_node_bitmap = make([]*bitstr_t, switch_record_cnt)
	switch_node_cnt = make([]int, switch_record_cnt)
	switch_required = make([]int, switch_record_cnt)

	for i := 0; i < switch_record_cnt; i++ {
		switch_ptr := switch_record_table[i]
		switch_node_bitmap[i] = bit_copy(switch_ptr.node_bitmap)
		bit_and(switch_node_bitmap[i], topo_eval.node_map)
		switch_node_cnt[i] = bit_set_count(switch_node_bitmap[i])

		if req_nodes_bitmap != nil && bit_overlap_any(req_nodes_bitmap, switch_node_bitmap[i]) {
			switch_required[i] = 1
			if (top_switch_inx == -1) ||
				(switch_record_table[i].level > switch_record_table[top_switch_inx].level) {
				top_switch_inx = i
			}
		}

		if !eval_nodes_enough_nodes(switch_node_cnt[i], rem_nodes) {
			continue
		}

		if req_nodes_bitmap == nil {
			if (top_switch_inx == -1) ||
				switch_record_table[i].level >= switch_record_table[top_switch_inx].level {
				top_switch_inx = i
			}
		}
	}

	if req_nodes_bitmap == nil {
		bit_clear_all(topo_eval.node_map)
	}

	/*
	 * Top switch is highest level switch containing all required nodes
	 * OR all nodes of the lowest scheduling weight
	 * OR -1 if can not identify top-level switch, which may be due to a
	 * disjoint topology and available nodes living on different switches.
	 */
	if top_switch_inx == -1 {
		log.Error("unable to identify top level switch")
		rc = slurm.ERROR
		goto fini
	}

	/* Check that all specifically required nodes are on shared network */
	if req_nodes_bitmap != nil &&
		!bit_super_set(req_nodes_bitmap,
			switch_node_bitmap[top_switch_inx]) {
		log.Error("required nodes are not on shared network")
		rc = slurm.ERROR
		goto fini
	}

	/*
	 * Remove nodes from consideration that can not be reached from this
	 * top level switch.
	 */
	for i := 0; i < switch_record_cnt; i++ {
		if top_switch_inx != i {
			bit_and(switch_node_bitmap[i], switch_node_bitmap[top_switch_inx])
		}
	}

	if req_nodes_bitmap != nil {
		bit_and(topo_eval.node_map, req_nodes_bitmap)
		if rem_nodes <= 0 {
			/* Required nodes completely satisfied the request */
			rc = slurm.SUCCESS
			goto fini
		}
	}

	goto try_again

try_again:
	/*
	 * Identify the best set of nodes (i.e. nodes with the lowest weight,
	 * in addition to the required nodes) that can be used to satisfy the
	 * job request. All nodes must be on a common top-level switch. The
	 * logic here adds groups of nodes, all with the same weight, so we
	 * usually identify more nodes than required to satisfy the request.
	 * Later logic selects from those nodes to get the best topology.
	 */
	best_node_cnt = 0
	best_nodes_bitmap = &bitstr_t{}
	for e := node_weight_list.Front(); e != nil; e = e.Next() {
		nw := e.Value.(*topo_weight_info_t)
		if bit_set_count(nw.node_bitmap) == 0 {
			continue
		}

		for i := 0; i < bit_set_count(nw.node_bitmap); i++ {
			node_ptr := (*nw.node_bitmap)[i]
			if !bit_test(switch_node_bitmap[top_switch_inx], node_ptr) {
				continue
			}
			if bit_set(best_nodes_bitmap, node_ptr) {
				best_node_cnt++
			}
		}

		if !sufficient {
			sufficient = eval_nodes_enough_nodes(best_node_cnt, rem_nodes)
		}
	}

	if !sufficient {
		log.Error("insufficient resources currently available")
		rc = slurm.ERROR
		goto fini
	}

	/*
	 * Construct a set of switch array entries.
	 * Use the same indexes as switch_record_table in slurmctld.
	 */
	bit_or(best_nodes_bitmap, topo_eval.node_map)
	for i := 0; i < switch_record_cnt; i++ {
		bit_and(switch_node_bitmap[i], best_nodes_bitmap)
		switch_node_cnt[i] = bit_set_count(switch_node_bitmap[i])
	}

	/* Add additional resources for already required leaf switches */
	if req_nodes_bitmap != nil {
		for i := 0; i < switch_record_cnt; i++ {
			if switch_required[i] == 0 || switch_node_bitmap[i] == nil ||
				switch_record_table[i].level != 0 {
				continue
			}
			for j := 0; j < bit_set_count(switch_node_bitmap[i]); j++ {
				if bit_test(topo_eval.node_map, (*switch_node_bitmap[i])[j]) {
					continue
				}
				rem_nodes--
				bit_set(topo_eval.node_map, (*switch_node_bitmap[i])[j])
				if rem_nodes <= 0 {
					rc = slurm.SUCCESS
					goto fini
				}
			}
		}
	}

	switches_dist = make([]uint32, switch_record_cnt)

	for i := 0; i < switch_record_cnt; i++ {
		if switch_required[i] == 1 {
			_topo_add_dist(&switches_dist, i)
		}
	}
	/* Add additional resources as required from additional leaf switches */
	prev_rem_nodes = rem_nodes + 1
	for {
		best_switch_inx := -1
		if prev_rem_nodes == rem_nodes {
			break /* Stalled */
		}
		prev_rem_nodes = rem_nodes

		for i := 0; i < switch_record_cnt; i++ {
			if switch_record_table[i].level != 0 {
				continue
			}
			_topo_choose_best_switch(&switches_dist, &switch_node_cnt, rem_nodes, i, &best_switch_inx)
		}
		if best_switch_inx == -1 {
			break
		}

		_topo_add_dist(&switches_dist, best_switch_inx)
		/*
		 * NOTE: Ideally we would add nodes in order of resource
		 * availability rather than in order of bitmap position, but
		 * that would add even more complexity and overhead.
		 */
		for i := 0; i < bit_set_count(switch_node_bitmap[best_switch_inx]); i++ {
			node_ptr := (*switch_node_bitmap[best_switch_inx])[i]
			if bit_test(topo_eval.node_map, node_ptr) {
				continue
			}
			if bit_set(topo_eval.node_map, node_ptr) {
				rem_nodes--
			}
			if rem_nodes <= 0 {
				rc = slurm.SUCCESS
				goto fini
			}
		}
		switch_node_cnt[best_switch_inx] = 0 /* Used all */
	}

fini:
	if rc == slurm.SUCCESS {
		leaf_switch_cnt := uint16(0)
		/* Count up leaf switches. */
		for i := 0; i < switch_record_cnt; i++ {
			if switch_record_table[i].level != 0 {
				continue
			}
			if bit_overlap_any(switch_node_bitmap[i], topo_eval.node_map) {
				leaf_switch_cnt++
			}
		}
		log.Debugf("Allocated %d nodes on %d leaf switches", bit_set_count(topo_eval.node_map), leaf_switch_cnt)
		topo_eval.leaf_switch_cnt = leaf_switch_cnt
	}

	return rc
}

/*
 * Allocate resources to the job on one leaf switch if possible,
 * otherwise distribute the job allocation over many leaf switches.
 */
func _eval_nodes_dfly(topo_eval *topology_eval_t) int {
	return 0
}

func eval_nodes_tree(topo_eval *topology_eval_t, have_dragonfly bool) int {
	if have_dragonfly {
		return _eval_nodes_dfly(topo_eval)
	} else {
		return _eval_nodes_topo(topo_eval)
	}
}
