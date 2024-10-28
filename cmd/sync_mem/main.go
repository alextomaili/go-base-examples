package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

type (
	F struct {
		f       uint32
		payload [32]int64
	}
)

func (f *F) SetReady() {
	if f == nil {
		return
	}
	f.f = 1
}

func (f *F) IsReady() bool {
	return f.f == 1
}

func (f *F) SetReadySync() {
	atomic.StoreUint32(&f.f, 1)
}

func (f *F) IsReadySync() bool {
	i := atomic.LoadUint32(&f.f)
	return i == 1
}

func main() {
	f := &F{}

	sync := true

	go func(f *F) {
		//calculate something
		for i := 0; i < len(f.payload)-1; i++ {
			f.payload[i] = time.Now().Unix()
		}

		//notify  when ready
		if sync {
			f.SetReadySync()
		} else {
			f.SetReady()
		}

		//do something else
		for {
		}
	}(f)

	//wait for ready
	w := 0
	for {
		if sync {
			if f.IsReadySync() {
				break
			}
		} else {
			if f.IsReady() {
				break
			}
		}
		w++
		fmt.Printf("waits: %d\n", w)
		time.Sleep(100 * time.Millisecond)
	}

	//use
	fmt.Printf("Done, w: %d, payload: %v \n", w, f.payload)
}

/*
https://stackoverflow.com/questions/2213428/invalidating-the-cpus-cache

When you perform a load without fences or mutexes, then the loaded value could potentially come from anywhere,
i.e, caches, registers (by way of compiler optimizations), or RAM...

In most mutex implementations, when you acquire a mutex, a fence is always applied, either explicitly
(e.g., mfence, barrier, etc.) or implicitly (e.g., lock prefix to lock the bus on x86). This causes the
cache-lines of all caches on the path to be invalidated.

Note that the entire cache isn't invalidated, just the respective cache-lines for the memory location.
This also includes the lines for the mutex (which is usually implemented as a value in memory).

Of course, there are architecture-specific details, but this is how it works in general.

Also note that this isn't the only reason for invalidating caches, as there may be operations on one
CPU that would need caches on another one to be invalidated.
Doing a google search for "cache coherence protocols" will provide you with a lot of
information on this subject.

https://www.d.umn.edu/~gshute/arch/cache-coherence.html

*/

/*
func (f *F) SetReady() {
  0x10c6ec0             4889442408              MOVQ AX, 0x8(SP)                     // mov %rax,0x8(%rsp)
        f.f = 1
  0x10c6ec5             8400                    TESTB AL, 0(AX)                      // test %al,(%rax)
  0x10c6ec7             c70001000000            MOVL $0x1, 0(AX)                     // movl $0x1,(%rax)
}
  0x10c6ecd             c3                      RET                                  // retq


func (f *F) SetReadySync() {
  0x10c6f20             55                      PUSHQ BP                             // push %rbp
  0x10c6f21             4889e5                  MOVQ SP, BP                          // mov %rsp,%rbp
  0x10c6f24             4883ec08                SUBQ $0x8, SP                        // sub $0x8,%rsp
  0x10c6f28             4889442418              MOVQ AX, 0x18(SP)                    // mov %rax,0x18(%rsp)
        atomic.StoreUint32(&f.f, 1)
  0x10c6f2d             8400                    TESTB AL, 0(AX)                      // test %al,(%rax)
  0x10c6f2f             48890424                MOVQ AX, 0(SP)                       // mov %rax,(%rsp)
  0x10c6f33             b901000000              MOVL $0x1, CX                        // mov $0x1,%ecx
  0x10c6f38             8708                    XCHGL CX, 0(AX)                      // xchg %ecx,(%rax)
}
  0x10c6f3a             4883c408                ADDQ $0x8, SP                        // add $0x8,%rsp
  0x10c6f3e             5d                      POPQ BP                              // pop %rbp
  0x10c6f3f             90                      NOPL                                 // nop
  0x10c6f40             c3                      RET                                  // retq

*/

/*
func (f *F) IsReady() bool {
  0x10c6ee0             55                      PUSHQ BP                             // push %rbp
  0x10c6ee1             4889e5                  MOVQ SP, BP                          // mov %rsp,%rbp
  0x10c6ee4             4883ec08                SUBQ $0x8, SP                        // sub $0x8,%rsp
  0x10c6ee8             4889442418              MOVQ AX, 0x18(SP)                    // mov %rax,0x18(%rsp)
  0x10c6eed             c644240700              MOVB $0x0, 0x7(SP)                   // movb $0x0,0x7(%rsp)
        return f.f == 1
  0x10c6ef2             488b4c2418              MOVQ 0x18(SP), CX                    // mov 0x18(%rsp),%rcx
  0x10c6ef7             8401                    TESTB AL, 0(CX)                      // test %al,(%rcx)
  0x10c6ef9             833901                  CMPL 0(CX), $0x1                     // cmpl $0x1,(%rcx)
  0x10c6efc             0f94c0                  SETE AL                              // sete %al
  0x10c6eff             88442407                MOVB AL, 0x7(SP)                     // mov %al,0x7(%rsp)
  0x10c6f03             4883c408                ADDQ $0x8, SP                        // add $0x8,%rsp
  0x10c6f07             5d                      POPQ BP                              // pop %rbp
  0x10c6f08             c3                      RET                                  // retq


func (f *F) IsReadySync() bool {
  0x10c6f60             55                      PUSHQ BP                             // push %rbp
  0x10c6f61             4889e5                  MOVQ SP, BP                          // mov %rsp,%rbp
  0x10c6f64             4883ec10                SUBQ $0x10, SP                       // sub $0x10,%rsp
  0x10c6f68             4889442420              MOVQ AX, 0x20(SP)                    // mov %rax,0x20(%rsp)
  0x10c6f6d             c644240300              MOVB $0x0, 0x3(SP)                   // movb $0x0,0x3(%rsp)
        i := atomic.LoadUint32(&f.f)
  0x10c6f72             488b4c2420              MOVQ 0x20(SP), CX                    // mov 0x20(%rsp),%rcx
  0x10c6f77             8401                    TESTB AL, 0(CX)                      // test %al,(%rcx)
  0x10c6f79             48894c2408              MOVQ CX, 0x8(SP)                     // mov %rcx,0x8(%rsp)
  0x10c6f7e             8b09                    MOVL 0(CX), CX                       // mov (%rcx),%ecx
  0x10c6f80             894c2404                MOVL CX, 0x4(SP)                     // mov %ecx,0x4(%rsp)
        return i == 1
  0x10c6f84             83f901                  CMPL CX, $0x1                        // cmp $0x1,%ecx
  0x10c6f87             0f94c0                  SETE AL                              // sete %al
  0x10c6f8a             88442403                MOVB AL, 0x3(SP)                     // mov %al,0x3(%rsp)
  0x10c6f8e             4883c410                ADDQ $0x10, SP                       // add $0x10,%rsp
  0x10c6f92             5d                      POPQ BP                              // pop %rbp
  0x10c6f93             c3                      RET                                  // retq

*/

/*
https://forum.osdev.org/viewtopic.php?t=33690

*/
