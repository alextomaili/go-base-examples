package main

import (
	"fmt"
	"github.com/alextomaili/go-base-examples/cmd/interfaces/api2"
	"reflect"
)

type (
	Foo struct {
		a int
		b string
	}
)

func main() {
	f := Foo{a: 10, b: "f-10"}

	t := reflect.TypeOf(f)

	api2.PrintName(t)
	fmt.Println(t.Name(), " ", t.Size(), " ", t.String())
}

/*
// any is an alias for interface{} and is equivalent to interface{} in all ways.
type any = interface{}

// efaceWords is interface{} internal representation.
type efaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

// EmptyInterface describes the layout of a "interface{}" or a "any."
// These are represented differently than non-empty interface, as the first
// word always points to an abi.Type.
type EmptyInterface struct {
	Type *Type
	Data unsafe.Pointer
}


*/
