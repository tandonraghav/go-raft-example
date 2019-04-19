package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/hashicorp/raft"
)

type fsm struct {
	mutex sync.Mutex
	name  map[string]string
}

type PostRequest struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (fsm *fsm) Apply(logEntry *raft.Log) interface{} {
	var e = PostRequest{}
	//Unmarshal the Log to get the Operation.
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		panic("Failed unmarshaling Raft log entry. This is a bug.")
	}

	switch e.Op {
	case "set":
		fsm.applyValue(e.Key, e.Value)
	default:
		panic(fmt.Sprintf("Unrecognized event type in Raft log entry: %s. This is a bug.", e.Op))
	}
	return nil
}

func (fsm *fsm) applyValue(key, value string) error {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	log.Println("setting key=  ", key)
	fsm.name[key] = value
	return nil
}

func (fsm *fsm) Snapshot() (raft.FSMSnapshot, error) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	// Clone the map.
	o := make(map[string]string)
	for k, v := range fsm.name {
		o[k] = v
	}
	return &fsmSnapshot{name: o}, nil
}

func (fsm *fsm) Restore(serialized io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(serialized).Decode(&o); err != nil {
		return err
	}

	fsm.name = o
	return nil
}
