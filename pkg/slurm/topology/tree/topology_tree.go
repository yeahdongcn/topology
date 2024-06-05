package tree

import (
	log "github.com/sirupsen/logrus"
)

/*
 * When TopologyParam=SwitchAsNodeRank is set, this plugin assigns a unique
 * node_rank for all nodes belonging to the same leaf switch.
 */
func topology_p_generate_node_ranking(filename string) error {
	/* By default, node_rank is 0, so start at 1 */
	switch_rank := 1

	log.Debugf("Generating node ranking %d", switch_rank)

	err := switch_record_validate(filename)
	if err != nil {
		return err
	}

	for sw := 0; sw < switch_record_cnt; sw++ {
		/* skip if not a leaf switch */
		if switch_record_table[sw].level != 0 {
			continue
		}

		for n := 0; n < node_record_cnt; n++ {
			if !bit_test(switch_record_table[sw].node_bitmap, node_record_table[n].name) {
				continue
			}

			node_record_table[n].node_rank = switch_rank
			log.Debugf("node=%s rank=%d", node_record_table[n].name, switch_rank)
		}

		switch_rank++
	}

	return nil
}
