package cache

import (
	"testing"
)

func TestNewFoo(t *testing.T) {
        var num1 int64 = 1
        s1 := "string_one"
        f := NewFoo(num1, s1)
        if f.GetNum() != num1 {
                t.Errorf("got int err for NewFoo(num1)")
        }
        if f.GetString() != s1 {
                t.Errorf("got string err for NewFoo(num1)")
        }
}
