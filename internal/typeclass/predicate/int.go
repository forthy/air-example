package predicate

var (
	PositiveInt Predicate[int] = func(v int) bool {
		return v > 0
	}
	ZeroInt Predicate[int] = func(v int) bool {
		return v == 0
	}
)
