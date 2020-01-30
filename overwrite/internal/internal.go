package internal

type X struct {
	val string
	err error
}

func (x X) GetVal() string {
	return x.val
}

func (x X) GetErr() error {
	return x.err
}

func NewX(a string, err error) X {
	return X{a, err}
}
