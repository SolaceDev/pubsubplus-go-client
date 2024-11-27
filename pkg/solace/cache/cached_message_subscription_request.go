// Package cache does stuff
package cache

type FooInterface interface {
        GetNum() int64
        GetString() string
}

type fooImpl struct {
        num int64
        str string
}

func (f *fooImpl) GetNum() int64 {
        return f.num
}

func (f *fooImpl) GetString() string {
        return f.str
}

func NewFoo(num int64, s string) FooInterface {
        return &fooImpl{num: num, str: s}
}
