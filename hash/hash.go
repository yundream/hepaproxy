package hash

import (
	"fmt"
	jump "github.com/renstrom/go-jump-consistent-hash"
	"math/rand"
)

type Hash struct {
	nodeSize int32
	offset   int32
	failNode []int
}

func New(node int32) *Hash {
	return &Hash{nodeSize: node, failNode: make([]int, 1024)}
}

func (h *Hash) AddFailNode(node int) {
	h.failNode[node] = 1
	fmt.Printf("%#v\n", h.failNode)
}

func (h *Hash) AddFailRange(from int, to int) {
	for i := from; i < to; i++ {
		h.failNode[i] = 1
	}
}

func (h Hash) GetFailNode() []int {
	return h.failNode
}

func (h Hash) GetNodeMulti(key uint64) int32 {
	r := rand.New(rand.NewSource(int64(key)))
	pivot := r.Int63n(20481)
	for {
		n := jump.Hash(key, h.nodeSize)
		n = n + h.offset
		if h.failNode[int(n)] == 0 {
			return n
		}
		key = uint64(pivot)
		pivot = r.Int63n(20481)
	}
}

func (h Hash) GetNode(key uint64) int32 {
	for {
		n := jump.Hash(key, h.nodeSize)
		n = n + h.offset
		if _, ok := h.failNode[int(n)]; !ok {
			return n
		}
		key++
	}
	return 1
}

func (h *Hash) SetNodeSize(node int32) {
	if h.nodeSize != node {
		h.nodeSize = node
	}
}

func (h *Hash) Offset(offset int32) {
	h.offset = offset
}
