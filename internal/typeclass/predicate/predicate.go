package predicate

type Predicate[T any] = func(v T) bool
