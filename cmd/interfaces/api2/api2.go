package api2

import (
	"fmt"
	"strconv"
)

type (
	Namer interface {
		Name() string
	}

	Ager interface {
		Age() int
	}
)

func PrintName(a Namer) {
	fmt.Println(a.Name())
}

func PrintAge(a Ager) {
	fmt.Println(a.Age())
}

func PrintAll(i interface{}) {
	r := ""

	n, cast := i.(Namer)
	if cast {
		r = r + "name: " + n.Name() + ", "
	}

	if a, cast := i.(Ager); cast {
		r = r + "age: " + strconv.Itoa(a.Age()) + ", "
	}

	fmt.Println(r)
}

/*
// efaceWords is interface{} internal representation.
type efaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}
*/
