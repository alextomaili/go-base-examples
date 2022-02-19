// Идея взята отсюда https://hackernoon.com/some-insights-on-maps-in-golang-rm5v3ywh

package map_internal

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"
)

const (
	keysInBucket = 8
	kvCount      = 100_000_000
	gb           = 1024 * 1024 * 1024
)

type (
	cmap map[int64]int64

	node struct {
		left   int32
		right  int32
		parent int32
		color  int32
		id     int64
		value  int64
	}
)

func TestMapInside(t *testing.T) {
	m := make(cmap)
	mt, _ := mapTypeAndValue(m)

	kvSize := uint16(mt.keysize + mt.elemsize)
	fmt.Printf("bucketsize: %v bytes\n", mt.bucketsize)
	fmt.Printf("overhead: %v\n", (mt.bucketsize - kvSize*keysInBucket))
	fmt.Printf("overhead: %v per entry\n", (mt.bucketsize-kvSize*keysInBucket)/keysInBucket)
}

func TestMapSizeA(t *testing.T) {
	mBefore := _alloc()

	m := make(cmap, kvCount/10)
	mt, hm := mapTypeAndValue(m)

	bucketsCount := 1 << hm.B
	mapSize := float32(mt.bucketsize) * float32(bucketsCount+int(hm.noverflow)) / gb

	fmt.Printf("before fill:\n")
	fmt.Printf("buckets count: %v\n", bucketsCount)
	fmt.Printf("size: %v Gb\n\n", mapSize)

	fmt.Printf("Elements | h.B | Buckets\n")
	var prevB uint8
	var i int64
	for i = 0; i < kvCount; i++ {
		m[i] = i * 3
		if hm.B != prevB {
			fmt.Printf("%8d | %3d | %8d\n", hm.count, hm.B, 1<<hm.B)
			prevB = hm.B
		}
	}

	mAfter := _alloc()

	bucketsCount = 1 << hm.B
	oldbucketsCount := 0
	if hm.oldbuckets != nil {
		oldbucketsCount = bucketsCount / 2
	}

	mapSize = float32(mt.bucketsize) * float32(bucketsCount+int(hm.noverflow))
	bytesPerItem := mapSize / kvCount
	mapSizeByGC := float32(mAfter - mBefore)
	bytesPerItemByGc := mapSizeByGC / kvCount

	fmt.Printf("\nafter fill:\n")
	fmt.Printf("buckets count: %v\n", bucketsCount)
	fmt.Printf("old buckets count: %v\n", oldbucketsCount)
	fmt.Printf("noverflow: %v\n", hm.noverflow)
	fmt.Printf("bucketsize: %v bytes\n", mt.bucketsize)
	fmt.Printf("size: %v Gb\n", mapSize/gb)
	fmt.Printf("bytes per item: %v\n", bytesPerItem)
	fmt.Printf("size by GC: %v Gb\n", mapSizeByGC/gb)
	fmt.Printf("bytes per item by Gc: %v\n\n", bytesPerItemByGc)
}

func TestMapSize(t *testing.T) {
	mBefore := _alloc()
	m := make(cmap, kvCount/10)
	for i := 0; i < kvCount; i++ {
		m[int64(i)] = int64(i * 3)
	}
	mAfter := _alloc()

	mt, hm := mapTypeAndValue(m)

	bucketsCount := 1 << hm.B
	mapSize := float32(mt.bucketsize) * float32(bucketsCount+int(hm.noverflow))
	bytesPerItem := mapSize / kvCount
	mapSizeByGC := float32(mAfter - mBefore)
	bytesPerItemByGc := mapSizeByGC / kvCount

	fmt.Printf("\nmap map[int64]int64:\n")
	fmt.Printf("buckets count: %v\n", bucketsCount)
	fmt.Printf("noverflow: %v\n", hm.noverflow)
	fmt.Printf("bucketsize: %v bytes\n", mt.bucketsize)
	fmt.Printf("size: %v Gb\n", mapSize/gb)
	fmt.Printf("bytes per item: %v\n", bytesPerItem)
	fmt.Printf("size by GC: %v Gb\n", mapSizeByGC/gb)
	fmt.Printf("bytes per item by Gc: %v\n\n", bytesPerItemByGc)

	nodeSize := unsafe.Sizeof(node{})
	treeSize := (float32(kvCount) + float32(kvCount/10)) * float32(nodeSize)
	fmt.Printf("\n[]node:\n")
	fmt.Printf("bytes per item: %v\n", nodeSize)
	fmt.Printf("size: %v Gb\n", treeSize/gb)
}

func _alloc() uint64 {
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}

var cK, cV int64

//go:noinline
func fillMap(count int) cmap {
	m := make(cmap, count)
	for i := 0; i < count; i++ {
		m[int64(i)] = int64(i * 3)
	}
	return m
}

//go:noinline
func fillSlice(count int) []node {
	a := make([]node, 0, count+count/10)
	for i := 0; i < count; i++ {
		a = append(a, node{
			id:    int64(i),
			value: int64(i * 3),
		})
	}
	return a
}

//go:noinline
func iterateMap(m cmap) (cK, cV int64) {
	for k, v := range m {
		cK += k
		cV += v
	}
	return
}

//go:noinline
func iterateSlice(a []node) (cK, cV int64) {
	for _, n := range a {
		cK += n.id
		cV += n.value
	}
	return
}

//go:noinline
func iterateSliceI(a []node) (cK, cV int64) {
	for i := 0; i < len(a); i++ {
		n := a[i]
		cK += n.id
		cV += n.value
	}
	return
}

func BenchmarkIterate(b *testing.B) {
	b.Run("map[int64]int64", func(b *testing.B) {
		b.StopTimer()
		m := fillMap(kvCount)

		runtime.GC()
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cK, cV = iterateMap(m)
		}
	})
	b.Run("[]node", func(b *testing.B) {
		b.StopTimer()
		a := fillSlice(kvCount)

		runtime.GC()
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cK, cV = iterateSlice(a)
		}
	})
	b.Run("[]nodeI", func(b *testing.B) {
		b.StopTimer()
		a := fillSlice(kvCount)

		runtime.GC()
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cK, cV = iterateSliceI(a)
		}
	})
}
