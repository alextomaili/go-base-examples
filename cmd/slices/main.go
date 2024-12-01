package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

/*
// https://www.w3schools.com/c/tryc.php?filename=demo_array_multi_access
#include <stdio.h>

int main() {
  int matrix[2][3] = { {1, 4, 2}, {11, 44, 22} };
  printf("%d\n", matrix[1][2]);

  int* p = &matrix;
  printf("%d\n", p[3*1+2]);

  int* p1 = p + 3;
  printf("%d\n", *p1);
  printf("%d\n", p1[2]);

  return 0;
}
*/

func playWithArrays() {
	a := [3]int{1, 2, 3}
	fmt.Println(a)

	b := [2][3]int{{1, 2, 3}, {11, 22, 33}}
	cb := b
	b[0][1] = 99
	fmt.Println(b)
	fmt.Println(cb)

	var p *int = (*int)(unsafe.Pointer(&b))
	fmt.Println(*p)

	var intExample int
	var p1 *int = (*int)(
		unsafe.Add(
			unsafe.Pointer(p),
			unsafe.Sizeof(intExample)*(1*3+1),
		),
	)
	fmt.Println(*p1)
}

/*
	type SliceHeader struct {
		Data uintptr
		Len  int
		Cap  int
	}
*/
func playWithSlices() {
	a := [6]int{1, 2, 3, 4, 5, 6}
	fmt.Println(a)

	s := a[0:6]
	fmt.Println(s, len(s), cap(s))

	s1 := s
	a[3] = 444
	s[2] = 333
	fmt.Println(s)
	fmt.Println(s1)

	s2 := s[3:6]
	fmt.Println(s2, len(s2), cap(s2))
}

func playWithSlices2() {
	s := make([]int, 5, 10)
	fmt.Println(s, len(s), cap(s))

	for i := 0; i < 5; i++ {
		s[i] = i * 10
	}
	fmt.Println(s, len(s), cap(s))

	if false {
		s[7] = 7000
		fmt.Println(s, len(s), cap(s))
	}

	s = append(s, 600, 700)
	fmt.Println(s, len(s), cap(s))

	for i := 0; i < 5; i++ {
		s = append(s, i*10)
	}
	fmt.Println(s, len(s), cap(s))

	s = s[0:0]
	fmt.Println(s, len(s), cap(s))

	for i := 0; i < 5; i++ {
		s = append(s, i*24)
	}
	fmt.Println(s, len(s), cap(s))

	s = make([]int, 0, 30)
	fmt.Println(s, len(s), cap(s))
}

func playWithSlices3() {
	s := make([]int, 5, 10)
	fmt.Println(s, len(s), cap(s))

	h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	fmt.Println(h.Len, h.Cap)
}

func playWithSlices4() {
	s := make([]int, 5, 10)

	f := func(s []int) {
		fmt.Println("f: ", s, len(s), cap(s))
		s[3] = 1024
		fmt.Println("f: ", s, len(s), cap(s))

		for i := 0; i < 24; i++ {
			s = append(s, 9)
		}
		fmt.Println("f: ", s, len(s), cap(s))
	}

	f(s)
	fmt.Println("main: ", s, len(s), cap(s))
}

func playWithSlices44() {
	s := make([]int, 5, 10)

	f := func(s *[]int) {
		fmt.Println("f: ", s, len(*s), cap(*s))
		(*s)[3] = 1024
		fmt.Println("f: ", s, len(*s), cap(*s))

		for i := 0; i < 24; i++ {
			*s = append(*s, 9)
		}
		fmt.Println("f: ", s, len(*s), cap(*s))
	}

	f(&s)
	fmt.Println("main: ", s, len(s), cap(s))
}

func main() {
	fmt.Println(">>>")

	playWithSlices44()
}
