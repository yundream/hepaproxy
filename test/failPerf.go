package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	nodeSize := flag.Int("node", 100, "Node size")
	count := flag.Int("count", 100, "Test loop count")
	flag.Init()
	fmt.Printf("Test Key : 0 ~ %d\n", count)

	api := fmt.Sprintf("http://localhost:8888/node/scale/%d", nodeSize)
	r, err := http.Post(api, "application/json", nil)
	if err != nil {
		panic(err)
	}
	if r.StatusCode != http.StatusOK {
		panic(err)
	}

	api := fmt.Sprintf("http://localhost:8888/node/scale/%d", nodeSize)
	r, err := http.Post(api, "application/json", nil)
	if err != nil {
		panic(err)
	}
	if r.StatusCode != http.StatusOK {
		panic(err)
	}

	for i := 0; i < count; i++ {
		api = fmt.Sprintf("http://localhost:8888/message/1/%d", i)
		r, err := http.Get(api)
		node := ioutil.ReadAll(r.Body)
		fmt.Println(i, ":", node)
	}
}
