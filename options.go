package btree

const (
	minDegree     = 2
	defaultDegree = 32
)

type LessFunc[K any] func(a, b K) int

type Options[K any] struct {
	Degree int
	Less   LessFunc[K]
}

func DefaultOptions[K any](less LessFunc[K]) Options[K] {
	return Options[K]{
		Degree: defaultDegree,
		Less:   less,
	}
}

func OptionsWithDegree[K any](degree int, less LessFunc[K]) Options[K] {
	return Options[K]{
		Degree: degree,
		Less:   less,
	}
}
