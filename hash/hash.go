package hash

import (
	"fmt"
	jump "github.com/renstrom/go-jump-consistent-hash"
)

type Hash struct {
	nodeSize int32
	offset   int32
	failNode map[int]int
}

func New(node int32) *Hash {
	return &Hash{nodeSize: node, failNode: make(map[int]int)}
}

func (h *Hash) SetFailNode(node []int) {
	for k := range h.failNode {
		delete(h.failNode, k)
	}
	for _, v := range node {
		h.failNode[v] = 1
	}
}

func (h Hash) GetFailNode() map[int]int {
	return h.failNode
}

func (h Hash) GetNode(key uint64) int32 {
	for {
		n := jump.Hash(key, h.nodeSize)
		n = n + h.offset
		fmt.Println("N:", n)
		if _, ok := h.failNode[int(n)]; !ok {
			return n
		}
		fmt.Println("This is Fail Node ", n)
		key++
	}
}

func (h *Hash) SetNodeSize(node int32) {
	if h.nodeSize != node {
		h.nodeSize = node
	}
}

func (h *Hash) Offset(offset int32) {
	h.offset = offset
}
