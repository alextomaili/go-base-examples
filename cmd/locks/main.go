package main

import (
	"fmt"
	"github.com/alextomaili/go-base-examples/pkg/accounts_store"
)

func main() {
	a := accounts_store.NewAccount(10242048)
	a.Apply(accounts_store.NewOperation(1, 10242048, +10, "deposit"))
	a.Apply(accounts_store.NewOperation(2, 10242048, +100, "transfer"))
	a.Apply(accounts_store.NewOperation(3, 10242048, +200, "transfer"))
	a.Apply(accounts_store.NewOperation(4, 10242048, +200, "transfer"))

	sn := a.Get(2)
	fmt.Println(sn)

	op := sn.GetLastOps()[0]
	fmt.Println(op)
}
