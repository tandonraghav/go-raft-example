package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Command line defaults
const (
	DefaultHTTPAddr = ":11000"
	DefaultRaftAddr = ":12000"
)

type httpService struct {
	node *Node
}

func NewHttpService(n *Node) *httpService {
	return &httpService{
		node: n,
	}
}

func (s *httpService) SetKeyHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	//log.Println(decoder)
	e := PostRequest{}
	err := decoder.Decode(&e)
	if err != nil {
		log.Println(err.Error())
	}
	e.Op = "set"
	log.Println("**** " + e.Key + " " + e.Value + " " + e.Op)
	b, _ := json.Marshal(e)
	s.node.raft.Apply(b, 10*time.Second)
	log.Println("Done!!!")
}

func (s *httpService) getKey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	e := PostRequest{}
	decoder.Decode(&e)
	s.node.fsm.mutex.Lock()
	defer s.node.fsm.mutex.Unlock()
	log.Println("Key == ", e.Key)
	log.Println("Value == ", s.node.fsm.name[e.Key])
}

func (s *httpService) joinHandler(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.node.Join(nodeID, remoteAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
