package hash

import (
	"fmt"
	"github.com/go-redis/redis"
	jump "github.com/renstrom/go-jump-consistent-hash"
	"math/rand"
	"strconv"
)

type Hash struct {
	nodeSize int32
	offset   int32
	failNode []int
	redisCli *redis.Client
}

func New(node int32, redisCli *redis.Client) *Hash {
	return &Hash{nodeSize: node, failNode: make([]int, 1024), redisCli: redisCli}
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
	i := 0
	var (
		r     *rand.Rand
		pivot int64
	)
	for {
		n := jump.Hash(key, h.nodeSize)
		n = n + h.offset
		if h.failNode[int(n)] == 0 {
			return n
		}
		if i == 0 {
			r = rand.New(rand.NewSource(int64(key)))
		}
		pivot = r.Int63n(20481)
		key = uint64(pivot)
		i++
	}
}

func (h Hash) GetNodeRedis(key uint64) int32 {
	n := jump.Hash(key, h.nodeSize)
	n = n + h.offset
	if h.failNode[int(n)] == 0 {
		return n
	}
	keys := strconv.Itoa(int(key))
	cmd := h.redisCli.Get(keys)
	if cmd.Err() != nil {
		fmt.Println(cmd.Err().Error())
	}
	_ = cmd.String()
	return 1
}

func (h Hash) GetNode(key uint64) int32 {
	/*
		for {
			n := jump.Hash(key, h.nodeSize)
			n = n + h.offset
			if _, ok := h.failNode[int(n)]; !ok {
				return n
			}
			key++
		}
	*/
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
