package sudoku

func deduplicate(s []index) []index {
	seen := make(map[index]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}

// remove an element from a slice. will not modify the original slice
func remove(ls []index, s index) []index {
	var c []index
	for _, x := range ls {
		if x != s {
			c = append(c, x)
		}
	}
	return c
}

func contains(ls []index, s index) bool {
	for _, l := range ls {
		if l == s {
			return true
		}
	}
	return false
}

func cross(a, b index) []index {
	i := make([]index, len(a)*len(b))
	idx := 0
	for _, sa := range a {
		for _, sb := range b {
			i[idx] = index(sa) + index(sb)
			idx++
		}
	}
	return i
}
