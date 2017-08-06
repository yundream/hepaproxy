package handler

import (
	"bitbucket.org/dream_yun/hepaProxy/hash"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Handler struct {
	jump     *hash.Hash
	router   *mux.Router
	insCount int
	count    []int
	offset   int32
	scale    int
}

func New() *Handler {
	return &Handler{jump: hash.New(10)}
}

func (h Handler) Router() *mux.Router {
	return h.router
}
func (h *Handler) Regist() {
	h.count = make([]int, 2048)
	h.router = mux.NewRouter()
	h.router.HandleFunc("/node/scale/{scale}/{offset}", h.ScaleNode).Methods("PUT", "POST")
	h.router.HandleFunc("/node/fail", h.FailNode).Methods("PUT", "POST")
	h.router.HandleFunc("/node/fail/{node}", h.FailNodeAdd).Methods("PUT")
	h.router.HandleFunc("/node/fail/{from}/{to}", h.FailNoadAddRange).Methods("PUT")
	h.router.HandleFunc("/message/{from}/{to}", h.SendMessage).Methods("POST")
	h.router.HandleFunc("/message", h.RecvMessage).Methods("POST")
	h.router.HandleFunc("/count", h.GetCount).Methods("GET")
	h.router.HandleFunc("/count", h.DelCount).Methods("DELETE")
}

func (h *Handler) ScaleNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scale, err := strconv.Atoi(vars["scale"])
	offset, err := strconv.Atoi(vars["offset"])
	if err != nil {
		w.Write([]byte("ERROR"))
		return
	}
	h.jump.Offset(int32(offset))
	h.jump.SetNodeSize(int32(scale))
	h.scale = scale
	h.offset = int32(offset)
	fmt.Fprintf(w, "%d ~ %d", offset, offset+int(scale)-1)
}

func (h *Handler) FailNodeAdd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n := vars["node"]
	noden, _ := strconv.Atoi(n)
	h.jump.AddFailNode(noden)
	//node := h.jump.GetFailNode()
}

func (h *Handler) FailNoadAddRange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	froms := vars["from"]
	tos := vars["to"]
	from, _ := strconv.Atoi(froms)
	to, _ := strconv.Atoi(tos)
	h.jump.AddFailRange(from, to)
}

func (h *Handler) FailNode(w http.ResponseWriter, r *http.Request) {
	/*
		data, err := ioutil.ReadAll(r.Body)
		var a []int
		json.Unmarshal(data, &a)
		if err != nil {
			w.Write([]byte("ERROR"))
			return
		}
		h.jump.SetFailNode(a)
		node := h.jump.GetFailNode()
		fmt.Fprintf(w, "Fail Node : %#v\n", node)
	*/
}

func (h *Handler) RecvMessage(w http.ResponseWriter, r *http.Request) {
	h.insCount++
	if h.insCount%100 == 0 {
		fmt.Println("Read Message ", h.insCount)
	}
}

func (h *Handler) GetCount(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Count : %#v\n", h.count[:h.scale])
}

func (h *Handler) DelCount(w http.ResponseWriter, r *http.Request) {
	h.count = make([]int, h.scale+1)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	to := vars["to"]
	keyTo, err := strconv.ParseUint(to, 10, 64)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	node := h.jump.GetNodeMulti(keyTo)
	h.count[node]++
}
