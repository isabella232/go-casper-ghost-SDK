package shared

// IntersectionUint64 of any number of uint64 slices with time
// complexity of approximately O(n) leveraging a map to
// check for element existence off by a constant factor
// of underlying map efficiency.
func IntersectionUint64(s ...[]uint64) []uint64 {
	if len(s) == 0 {
		return []uint64{}
	}
	if len(s) == 1 {
		return s[0]
	}
	intersect := make([]uint64, 0)
	m := make(map[uint64]int)
	for _, k := range s[0] {
		m[k] = 1
	}
	for i, num := 1, len(s); i < num; i++ {
		for _, k := range s[i] {
			// Increment and check only if item is present in both, and no increment has happened yet.
			if _, found := m[k]; found && i == m[k] {
				m[k]++
				if m[k] == num {
					intersect = append(intersect, k)
				}
			}
		}
	}
	return intersect
}