package tree

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const INFINITE = 0xffffffff

var (
	switch_record_table = make([]*switch_record_t, 0)
	switch_record_cnt   = 0

	node_record_table = make([]*node_record_t, 0)
	node_record_cnt   = 0
)

type node_record_t struct {
	name string /* name of the node. NULL==defunct */

	/* Hilbert number based on node name,
	 * or other sequence number used to
	 * order nodes by location,
	 * no need to save/restore */
	node_rank int
}

type slurm_conf_switches_t struct {
	link_speed  uint32 /* link speed, arbitrary units */
	nodes       string /* names of nodes directly connect to this switch, if any */
	switch_name string /* name of this switch */
	switches    string /* names if child switches directly connected to this switch, if any */
}

type switch_record_t struct {
	level             int       /* level in hierarchy, leaf=0 */
	link_speed        uint32    /* link speed, arbitrary units */
	name              string    /* switch name */
	node_bitmap       *bitstr_t /* bitmap of all nodes descended from this switch */
	nodes             string    /* name of direct descendant nodes */
	num_desc_switches uint16    /* number of descendant switches */
	num_switches      uint16    /* number of direct descendant switches */
	parent            uint16    /* index of parent switch */
	switch_bitmap     *bitstr_t /* XXX: bitmap of all switches descended from this switch */
	switches          string    /* name of direct descendant switches */
	switches_dist     []uint32  /* distance to other switches */
	switch_desc_index []uint16  /* indexes of child descendant switches */
	switch_index      []uint16  /* indexes of child direct descendant switches */
}

func _parse_switches(f *os.File) ([]*slurm_conf_switches_t, error) {
	list := []*slurm_conf_switches_t{}

	s := bufio.NewScanner(f)
	for s.Scan() {
		txt := s.Text()
		if !strings.HasPrefix(txt, "#") && strings.TrimSpace(txt) != "" {
			table := strings.Split(txt, " ")

			switches := &slurm_conf_switches_t{}

			for _, column := range table {
				pair := strings.Split(column, "=")

				switch pair[0] {
				case "LinkSpeed":
					linkSpeed, err := strconv.ParseUint(pair[1], 10, 32)
					if err != nil {
						log.Errorf("Failed to parse LinkSpeed: %s", pair[1])
						continue
					}
					switches.link_speed = uint32(linkSpeed)
				case "SwitchName":
					switches.switch_name = pair[1]
				case "Nodes":
					switches.nodes = pair[1]
				case "Switches":
					switches.switches = pair[1]
				}
			}
			if len(switches.nodes) > 0 && len(switches.switches) > 0 {
				log.Errorf("switch %s has both child switches and nodes", switches.switch_name)
				continue
			}

			if len(switches.nodes) == 0 && len(switches.switches) == 0 {
				log.Errorf("switch %s has neither child switches nor nodes", switches.switch_name)
				continue
			}

			list = append(list, switches)
		}
	}
	return list, nil
}

func _read_topo_file(filename string) ([]*slurm_conf_switches_t, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open the %s file", filename)
	}
	defer f.Close()

	log.Tracef("Reading the %s file", filename)

	return _parse_switches(f)
}

/*
 * _find_child_switches creates an array of indexes to the
 * immediate descendants of switch sw.
 */
func _find_child_switches(sw int) {
	switch_record_table[sw].num_switches = uint16(bit_set_count(switch_record_table[sw].switch_bitmap))
	switch_record_table[sw].switch_index = make([]uint16, switch_record_table[sw].num_switches)

	cldx := 0
	for _, swname := range *switch_record_table[sw].switch_bitmap {
		for i := 0; i < switch_record_cnt; i++ {
			if swname == switch_record_table[i].name {
				switch_record_table[sw].switch_index[cldx] = uint16(i)
				switch_record_table[i].parent = uint16(sw)
				cldx++
				break
			}
		}
	}
}

/*
 * _find_desc_switches creates an array of indexes to the
 * all descendants of switch sw.
 */
func _find_desc_switches(sw int) {
	switchDescIndex := _merge_switches_array(
		switch_record_table[sw].switch_desc_index,
		switch_record_table[sw].switch_index)
	switch_record_table[sw].switch_desc_index = switchDescIndex
	switch_record_table[sw].num_desc_switches = uint16(len(switchDescIndex))

	for k := uint16(0); k < switch_record_table[sw].num_switches; k++ {
		child_index := switch_record_table[sw].switch_index[k]
		switchDescIndex = _merge_switches_array(
			switch_record_table[sw].switch_desc_index,
			switch_record_table[child_index].switch_desc_index,
		)
		switch_record_table[sw].switch_desc_index = switchDescIndex
		switch_record_table[sw].num_desc_switches = uint16(len(switchDescIndex))
	}
}

func _merge_switches_array(a1, a2 []uint16) []uint16 {
	for i := 0; i < len(a2); i++ {
		j := 0
		for j = 0; j < len(a1); j++ {
			if a1[j] == a2[i] {
				break
			}
		}
		if j < len(a1) {
			continue
		}
		a1 = append(a1, a2[i])
	}
	return a1
}

func _check_better_path(i, j, k int) {
	tmp := uint32(0)
	if switch_record_table[j].switches_dist[i] == INFINITE ||
		switch_record_table[i].switches_dist[k] == INFINITE {
		tmp = INFINITE
	} else {
		tmp = switch_record_table[j].switches_dist[i] +
			switch_record_table[i].switches_dist[k]
	}
	if switch_record_table[j].switches_dist[k] > tmp {
		switch_record_table[j].switches_dist[k] = tmp
	}
}

func _node_name2bitmap(node_names string) (*bitstr_t, error) {
	// Regular expression to match patterns like "node[1-3,5,7-9]"
	re := regexp.MustCompile(`([a-zA-Z]+)\[([0-9,-]+)\]`)
	matches := re.FindStringSubmatch(node_names)

	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid hostlist expression")
	}

	prefix := matches[1]
	ranges := strings.Split(matches[2], ",")

	var my_bitmap bitstr_t
	for _, r := range ranges {
		if strings.Contains(r, "-") {
			// Handle range like "1-3"
			bounds := strings.Split(r, "-")
			lower, err := strconv.Atoi(bounds[0])
			if err != nil {
				return nil, err
			}
			upper, err := strconv.Atoi(bounds[1])
			if err != nil {
				return nil, err
			}
			for i := lower; i <= upper; i++ {
				format := "%s%d"
				if strings.HasPrefix(bounds[0], "0") {
					format = fmt.Sprintf("%%s%%0" + fmt.Sprintf("%d", len(bounds[0])) + "d")
				}
				bit_set(&my_bitmap, fmt.Sprintf(format, prefix, i))
			}
		} else {
			// Handle single number like "5"
			num, err := strconv.Atoi(r)
			if err != nil {
				return nil, err
			}
			format := "%s%d"
			if strings.HasPrefix(r, "0") {
				format = fmt.Sprintf("%%s%%0" + fmt.Sprintf("%d", len(r)) + "d")
			}
			bit_set(&my_bitmap, fmt.Sprintf(format, prefix, num))
		}
	}

	return &my_bitmap, nil
}

/* Return the index of a given switch name or -1 if not found */
func _get_switch_inx(table *map[string]int, name string) int {
	if index, ok := (*table)[name]; ok {
		return index
	}

	return -1
}

// XXX: Use a modern Go idiom to avoid this function
func _switch_record_reset() {
	switch_record_table = make([]*switch_record_t, 0)
	switch_record_cnt = 0

	node_record_table = make([]*node_record_t, 0)
	node_record_cnt = 0
}

func switch_record_validate(filename string) error {
	_switch_record_reset()

	ptr_array, err := _read_topo_file(filename)
	if err != nil {
		return err
	}

	switch_record_cnt = len(ptr_array)
	if switch_record_cnt == 0 {
		return fmt.Errorf("no switches configured")
	}

	// XXX: Global variable?
	switch_record_lookup_table := map[string]int{}
	node_record_lookup_table := map[string]int{}

	for i, ptr := range ptr_array {
		switch_ptr := &switch_record_t{}

		switch_ptr.name = ptr.switch_name
		/* See if switch name has already been defined. */
		if _, ok := switch_record_lookup_table[ptr.switch_name]; ok {
			log.Errorf("Switch (%s) has already been defined", switch_ptr.name)
			continue
		}

		switch_ptr.link_speed = ptr.link_speed
		if len(ptr.nodes) > 0 {
			switch_ptr.level = 0 /* leaf switch */
			switch_ptr.nodes = strings.Clone(ptr.nodes)
			node_bitmap, err := _node_name2bitmap(ptr.nodes)
			if err != nil {
				log.Fatalf("Invalid node name (%s) in switch config (%s)",
					ptr.nodes, ptr.switch_name)
			} else {
				switch_ptr.node_bitmap = node_bitmap

				// XXX: Add nodes to node_record_table
				for _, name := range *node_bitmap {
					if _, ok := node_record_lookup_table[name]; !ok {
						node_ptr := &node_record_t{
							name: name,
						}
						node_record_table = append(node_record_table, node_ptr)
						node_record_lookup_table[name] = node_record_cnt
						node_record_cnt++
					}
				}
			}
		} else if len(ptr.switches) > 0 {
			switch_ptr.level = -1 /* determine later */
			switch_ptr.switches = strings.Clone(ptr.switches)
			switch_bitmap, err := _node_name2bitmap(ptr.switches)
			if err != nil {
				log.Fatalf("Invalid switch name (%s) in switch config (%s)",
					ptr.switches, ptr.switch_name)
			} else {
				switch_ptr.switch_bitmap = switch_bitmap
			}
		} else {
			log.Fatalf("Switch configuration (%s) lacks children", ptr.switch_name)
		}

		switch_record_lookup_table[ptr.switch_name] = i
		switch_record_table = append(switch_record_table, switch_ptr)
	}

	for depth := 1; ; depth++ {
		resolved := true

		for i := 0; i < switch_record_cnt; i++ {
			switch_ptr := switch_record_table[i]
			if switch_ptr.level != -1 {
				continue
			}
			for p := bit_set_count(switch_ptr.switch_bitmap) - 1; p >= 0; p-- {
				child := (*switch_ptr.switch_bitmap)[p]
				j := _get_switch_inx(&switch_record_lookup_table, child)
				if j < 0 || j == i {
					log.Fatalf("Switch configuration %s has invalid child (%s)",
						switch_ptr.name, child)
				}
				if switch_record_table[j].level == -1 {
					/* Children not resolved */
					resolved = false
					switch_ptr.level = -1
					break
				}
				if switch_ptr.level == -1 {
					switch_ptr.level = switch_record_table[j].level + 1
					switch_ptr.node_bitmap = bit_copy(switch_record_table[j].node_bitmap)
				} else {
					switch_ptr.level = max(switch_ptr.level, switch_record_table[j].level+1)
					bit_or(switch_ptr.node_bitmap, switch_record_table[j].node_bitmap)
				}
			}
		}
		if resolved {
			break
		}
		if depth > 20 {
			log.Fatal("Switch configuration is not a tree")
		}
	}

	switchlevels := 0
	for i := 0; i < switch_record_cnt; i++ {
		switchlevels = max(switchlevels, switch_record_table[i].level)
		switch_ptr := switch_record_table[i]
		if bit_set_count(switch_ptr.node_bitmap) == 0 {
			log.Errorf("switch %s has no nodes", switch_ptr.name)
		}
	}

	log.Debugf("Switch levels: %d", switchlevels)

	/* Create array of indexes of children of each switch,
	 * and see if any switch can reach all nodes */
	for i := 0; i < switch_record_cnt; i++ {
		if switch_record_table[i].level != 0 {
			_find_child_switches(switch_record_lookup_table[switch_record_table[i].name])
		}
	}

	for i := 0; i < switch_record_cnt; i++ {
		switch_record_table[i].switches_dist = make([]uint32, switch_record_cnt)
		switch_record_table[i].switch_desc_index = make([]uint16, 0)
		switch_record_table[i].num_desc_switches = 0
	}

	for i := 0; i < switch_record_cnt; i++ {
		for j := i + 1; j < switch_record_cnt; j++ {
			switch_record_table[i].switches_dist[j] = INFINITE
			switch_record_table[j].switches_dist[i] = INFINITE
		}
		for j := 0; j < int(switch_record_table[i].num_switches); j++ {
			child := switch_record_table[i].switch_index[j]

			switch_record_table[i].switches_dist[child] = 1
			switch_record_table[child].switches_dist[i] = 1
		}
	}

	for i := 0; i < switch_record_cnt; i++ {
		for j := 0; j < switch_record_cnt; j++ {
			for k := 0; k < switch_record_cnt; k++ {
				_check_better_path(i, j, k)
			}
		}
	}

	for i := 1; i <= switchlevels; i++ {
		for j := 0; j < switch_record_cnt; j++ {
			if switch_record_table[j].level != i {
				continue
			}
			_find_desc_switches(j)
		}
	}

	return nil
}
