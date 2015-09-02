package collections

func NewStringsSet(l []string) map[string]bool {
	result := map[string]bool{}
	for _, e := range l {
		result[e] = true
	}
	return result
}

func DedupStringsList(l []string) []string {
	var result []string
	for e := range NewStringsSet(l) {
		result = append(result, e)
	}
	return result
}

func DedupStableStringsList(l []string) []string {
	s := map[string]bool{}
	var result []string
	for _, e := range l {
		if _, ok := s[e]; !ok {
			result = append(result, e)
		}
		s[e] = true
	}
	return result
}
