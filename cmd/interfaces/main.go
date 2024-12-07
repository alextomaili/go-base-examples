package main

import (
	"fmt"
	"github.com/alextomaili/go-base-examples/cmd/interfaces/api2"
	"strconv"
)

type (
	Foo struct {
		a int
		b string
	}

	MyInt int
)

func (s *Foo) A() string {
	s.b = s.b + " after A()"
	fmt.Println(s.b)
	return s.b
}

func (s Foo) B() string {
	s.b = s.b + " after B()"
	fmt.Println(s.b)
	return s.b
}

func (mi MyInt) Name() string {
	return strconv.Itoa(int(mi))
}

type (
	Animal struct {
		name string
	}
)

func (a *Animal) Name() string {
	return a.name
}

func (a *Animal) Age() int {
	return 1
}

func main() {
	f := Foo{
		a: 10,
		b: "f-10",
	}

	f.A()
	f.B()
	fmt.Println(f.b)

	var m MyInt = 10
	fmt.Println(m.Name())

	a := &Animal{
		name: "Bobik",
	}
	api2.PrintName(a)
	api2.PrintAge(a)
	api2.PrintAll(a)

}
