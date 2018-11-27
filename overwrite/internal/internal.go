package internal

type X struct {
	val string
}

func (x X) Get()string {
	return x.val
}

func NewX(a string) X{
	return X{a}
}

