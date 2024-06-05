package tree

import (
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

type bitstr_t []string

func bit_super_set(b1, b2 *bitstr_t) bool {
	for _, bit := range *b1 {
		if !bit_test(b2, bit) {
			return false
		}
	}
	return true
}

func bit_overlap_any(b1, b2 *bitstr_t) bool {
	for _, bit := range *b1 {
		if bit_test(b2, bit) {
			return true
		}
	}
	return false
}

func bit_set(b *bitstr_t, bit string) bool {
	increased := false
	if !bit_test(b, bit) {
		*b = append(*b, bit)
		increased = true
	}
	sort.Slice(*b, func(i, j int) bool {
		numI, _ := strconv.Atoi(strings.TrimLeftFunc((*b)[i], unicode.IsLetter))
		numJ, _ := strconv.Atoi(strings.TrimLeftFunc((*b)[j], unicode.IsLetter))
		return numI < numJ
	})
	return increased
}

func bit_test(b *bitstr_t, bit string) bool {
	return slices.Contains(*b, bit)
}

func bit_or(b1, b2 *bitstr_t) {
	set := make(map[string]struct{})
	for _, v := range *b1 {
		set[v] = struct{}{}
	}
	for _, v := range *b2 {
		set[v] = struct{}{}
	}

	new := make([]string, 0, len(set))
	for k := range set {
		new = append(new, k)
	}

	sort.Slice(new, func(i, j int) bool {
		numI, _ := strconv.Atoi(strings.TrimLeftFunc(new[i], unicode.IsLetter))
		numJ, _ := strconv.Atoi(strings.TrimLeftFunc(new[j], unicode.IsLetter))
		return numI < numJ
	})
	*b1 = new
}

func bit_and(b1, b2 *bitstr_t) {
	var new bitstr_t
	for _, bit := range *b1 {
		if bit_test(b2, bit) {
			new = append(new, bit)
		} else {
			log.Tracef("Removing %s from hostlist", bit)
		}
	}
	*b1 = new
}

func bit_copy(b *bitstr_t) *bitstr_t {
	new := make(bitstr_t, len(*b))
	copy(new, *b)
	return &new
}

func bit_set_count(b *bitstr_t) int {
	return len(*b)
}

func bit_clear_all(b *bitstr_t) {
	*b = (*b)[:0]
}
