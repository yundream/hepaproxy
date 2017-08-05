package handler

import (
	"bitbucket.org/dream_yun/hepaProxy/hash"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Handler struct {
	jump     *hash.Hash
	router   *mux.Router
	insCount int
	count    []int
	offset   int32
}

func New() *Handler {
	return &Handler{jump: hash.New(10)}
}

func (h Handler) Router() *mux.Router {
	return h.router
}
func (h *Handler) Regist() {
	h.router = mux.NewRouter()
	h.router.HandleFunc("/node/scale/{scale}/{offset}", h.ScaleNode).Methods("PUT", "POST")
	h.router.HandleFunc("/node/fail", h.FailNode).Methods("PUT", "POST")
	h.router.HandleFunc("/message/{from}/{to}", h.SendMessage).Methods("POST")
	h.router.HandleFunc("/message", h.RecvMessage).Methods("POST")
	h.router.HandleFunc("/count", h.GetCount).Methods("GET")
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
	h.count = make([]int, scale+1)
	h.offset = int32(offset)
	fmt.Fprintf(w, "%d ~ %d", offset, offset+int(scale)-1)
}

func (h *Handler) FailNode(w http.ResponseWriter, r *http.Request) {
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
}

func (h *Handler) RecvMessage(w http.ResponseWriter, r *http.Request) {
	h.insCount++
	if h.insCount%100 == 0 {
		fmt.Println("Read Message ", h.count)
	}
}

func (h *Handler) GetCount(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Count : %#v\n", h.count)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	to := vars["to"]
	fmt.Println("To : ", to)
	keyTo, err := strconv.ParseUint(to, 10, 64)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	node := h.jump.GetNode(keyTo)
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/message", node), "text/plain", nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "SendMessageFail : %d node %d\n", node, resp.StatusCode)
		return
	}
	h.count[node-h.offset]++
	fmt.Fprintf(w, "SendMessage : %d node", node)
}
