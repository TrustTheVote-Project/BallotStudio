package main

import (
	"container/heap"
)

type cacheEntry struct {
	key   string
	data  interface{}
	size  uint64
	seen  uint64
	seeni int
	//expires time.Time
}

type seenHeap struct {
	they []*cacheEntry
}

// heap.Interface sort.Interface
func (sh *seenHeap) Len() int {
	return len(sh.they)
}

// heap.Interface sort.Interface
func (sh *seenHeap) Less(i, j int) bool {
	return sh.they[i].seen < sh.they[j].seen
}

// heap.Interface sort.Interface
func (sh *seenHeap) Swap(i, j int) {
	t := sh.they[i]
	sh.they[i] = sh.they[j]
	sh.they[i].seeni = i
	sh.they[j] = t
	sh.they[j].seeni = j
}

// heap.Interface
func (sh *seenHeap) Push(x interface{}) {
	it := x.(*cacheEntry)
	sh.they = append(sh.they, it)
}

// heap.Interface
func (sh *seenHeap) Pop() interface{} {
	last := len(sh.they) - 1
	out := sh.they[last]
	sh.they = sh.they[:last]
	return out
}

type expireHeap []*cacheEntry

type Cache struct {
	MaxSize     uint64
	byKey       map[string]*cacheEntry
	currentSize uint64
	bySeen      seenHeap
	ai          uint64 // access counter by which get/put are seen
}

func (c *Cache) Put(key string, v interface{}, size int) {
	ent := &cacheEntry{
		key:  key,
		data: v,
		size: uint64(size),
		seen: c.ai,
	}
	c.ai++
	if c.byKey == nil {
		c.byKey = make(map[string]*cacheEntry)
	}
	prev := c.byKey[key]
	if prev != nil {
		c.currentSize -= prev.size
		c.currentSize += uint64(size)
		c.bySeen.they[prev.seeni] = ent
		heap.Fix(&c.bySeen, prev.seeni)
		c.byKey[key] = ent
	} else {
		c.byKey[key] = ent
		c.currentSize += uint64(size)
		heap.Push(&c.bySeen, ent)
	}
	if c.MaxSize == 0 {
		c.MaxSize = 10000000
	}
	for c.currentSize > c.MaxSize {
		oldest := heap.Pop(&c.bySeen).(cacheEntry)
		delete(c.byKey, oldest.key)
		c.currentSize -= oldest.size
	}
}

func (c *Cache) Invalidate(key string) {
	delete(c.byKey, key)
}

func (c *Cache) Get(key string) interface{} {
	ent := c.byKey[key]
	if ent == nil {
		return nil
	}
	ent.seen = c.ai
	c.ai++
	heap.Fix(&c.bySeen, ent.seeni)
	return ent.data
}
