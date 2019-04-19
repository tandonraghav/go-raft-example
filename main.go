package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var httpAddr string
var raftAddr string
var joinAddr string
var nodeID string

func init() {
	//flag.BoolVar(&inmem, "inmem", false, "Use in-memory storage for Raft")
	flag.StringVar(&httpAddr, "hport", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&nodeID, "id", "", "Node ID")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	// Ensure Raft storage exists.
	raftDir := flag.Arg(0)
	if raftDir == "" {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}
	os.MkdirAll(raftDir, 0700)

	node := Node{
		RaftBind: raftAddr,
		RaftDir:  raftDir,
	}

	httpService := NewHttpService(&node)

	if err := node.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	if joinAddr != "" {
		join(joinAddr, raftAddr, nodeID)
	}

	r := mux.NewRouter()
	r.HandleFunc("/set", httpService.SetKeyHandler).Methods("POST")
	r.HandleFunc("/join", httpService.joinHandler).Methods("POST")
	r.HandleFunc("/get", httpService.getKey).Methods("POST")
	http.ListenAndServe(httpAddr, r)

}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	log.Println("Going to post @", joinAddr)
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application/json", bytes.NewReader(b))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	return nil
}
