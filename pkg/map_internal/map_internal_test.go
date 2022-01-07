// Идея взята отсюда https://hackernoon.com/some-insights-on-maps-in-golang-rm5v3ywh

package map_internal

import (
	"fmt"
	"runtime"
	"testing"
)

const (
	keysInBucket = 8
	kvCount      = 100_000_000
	gb           = 1024 * 1024 * 1024
)

type cmap map[int64]int64

func TestMapInside(t *testing.T) {
	m := make(cmap)
	mt, _ := mapTypeAndValue(m)

	kvSize := uint16(mt.keysize + mt.elemsize)
	fmt.Printf("bucketsize: %v bytes\n", mt.bucketsize)
	fmt.Printf("overhead: %v\n", (mt.bucketsize - kvSize*keysInBucket))
	fmt.Printf("overhead: %v per entry\n", (mt.bucketsize-kvSize*keysInBucket)/keysInBucket)
}

func TestMapSize(t *testing.T) {
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
	mapSize = float32(mt.bucketsize) * float32(bucketsCount+int(hm.noverflow))
	bytesPerItem := mapSize / kvCount
	mapSizeByGC := float32(mAfter - mBefore)
	bytesPerItemByGc := mapSizeByGC / kvCount

	fmt.Printf("\nafter fill:\n")
	fmt.Printf("buckets count: %v\n", bucketsCount)
	fmt.Printf("noverflow: %v\n", hm.noverflow)
	fmt.Printf("bucketsize: %v bytes\n", mt.bucketsize)
	fmt.Printf("size: %v Gb\n", mapSize/gb)
	fmt.Printf("bytes per item: %v\n", bytesPerItem)
	fmt.Printf("size by GC: %v Gb\n", mapSizeByGC/gb)
	fmt.Printf("bytes per item by Gc: %v\n\n", bytesPerItemByGc)
}

func _alloc() uint64 {
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}
