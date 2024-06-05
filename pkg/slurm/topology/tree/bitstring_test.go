package tree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHostlist(t *testing.T) {
	b1 := bitstr_t{"a1", "a2", "a10"}
	b2 := bitstr_t{"a2", "a3", "a11"}

	bit_and(&b1, &b2)
	require.Equal(t, 1, bit_set_count(&b1))
	require.Equal(t, bitstr_t{"a2"}, b1)

	b1 = bitstr_t{"a1", "a2", "a10"}
	b2 = bitstr_t{"a2", "a3", "a11"}

	bit_or(&b1, &b2)
	require.Equal(t, 5, bit_set_count(&b1))
	require.Equal(t, bitstr_t{"a1", "a2", "a3", "a10", "a11"}, b1)

	b1 = bitstr_t{}
	b2 = bitstr_t{"b", "c", "d"}
	b1 = *bit_copy(&b2)
	require.Equal(t, false, &b1 == &b2)

	require.Equal(t, true, bit_test(&b1, "b"))
	require.Equal(t, false, bit_test(&b1, "a"))

	b1 = bitstr_t{"a1", "a2", "a10"}
	increased := bit_set(&b1, "a3")
	require.Equal(t, true, increased)
	require.Equal(t, 4, bit_set_count(&b1))
	require.Equal(t, bitstr_t{"a1", "a2", "a3", "a10"}, b1)

	b2 = bitstr_t{"a3"}
	require.Equal(t, len(b1), bit_set_count(&b1))
	require.Equal(t, true, bit_overlap_any(&b1, &b2))

	b1 = bitstr_t{"a1", "a2", "a10"}
	bit_clear_all(&b1)
	require.Equal(t, 0, bit_set_count(&b1))
}
