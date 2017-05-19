package main

import (
	"container/heap"
	"sync"
//	"fmt"
)

type Key struct {
	Name string
	Hits int
}

type KeyStat struct {
	RCount uint64
	WCount uint64
	RBytes uint64
	WBytes uint64
	Name string
}

type Stat struct {
	Lock sync.Mutex
	keys map[string]KeyStat
}

type OutputHeap []*KeyStat

func (h OutputHeap) Len() int { return len(h) }

func (h OutputHeap) Less(i, j int) bool {
	// Use global var to sort by
	switch Config_.SortBy {
	case "rbytes":	return h[i].RBytes > h[j].RBytes
	case "wbytes":	return h[i].WBytes > h[j].WBytes
	case "rcount":	return h[i].RCount > h[j].RCount
	case "wcount":	return h[i].WCount > h[j].WCount
	}
	return h[i].RCount > h[j].RCount
}
func (h OutputHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *OutputHeap) Push(x interface{}) {
	*h = append(*h, x.(*KeyStat))
}

func (h *OutputHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func NewStat() *Stat {
	s := &Stat{}
	s.keys = make(map[string]KeyStat)
	return s
}

func (s *Stat) Add(keys []KeyStat) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	for _, key := range keys {
		if k, ok := s.keys[key.Name]; ok {
			k.WBytes += key.WBytes
			k.RBytes += key.RBytes
			k.WCount += key.WCount
			k.RCount += key.RCount
			s.keys[key.Name] = k
		} else {
			s.keys[key.Name] = key
		}
	}
}

func (s *Stat) GetTopKeys() *OutputHeap {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	top := &OutputHeap{}
	heap.Init(top)

	for _, key := range s.keys {
		heap.Push(top, &KeyStat{key.RCount, key.WCount, key.RBytes, key.WBytes, key.Name})
	}
	return top
}

func (s *Stat) Rotate(clear bool) *Stat {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	new_stat := NewStat()
	new_stat.keys = s.keys

	if clear {
		s.keys = make(map[string]KeyStat)
	}
	return new_stat
}
