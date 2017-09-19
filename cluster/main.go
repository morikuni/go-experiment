package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
)

type State struct {
	kvs       map[string]string
	broadcast *memberlist.TransmitLimitedQueue
	raft      *raft.Raft
	raftPort  string
	peerStore raft.PeerStore
	ml        *memberlist.Memberlist
}

func (s *State) NotifyJoin(node *memberlist.Node) {
	fmt.Println("NotifyJoin", node.Addr.String(), string(node.Meta))
	if s.IsLeader() {
		f := s.raft.AddPeer(node.Addr.String() + ":" + string(node.Meta))
		if err := f.Error(); err != raft.ErrKnownPeer && err != nil {
			panic(err)
		}
	}
}

func (s *State) NotifyLeave(node *memberlist.Node) {
	fmt.Println("NotifyLeave")
	// if s.IsLeader() {
	// 	f := s.raft.RemovePeer(node.Addr.String() + ":" + string(node.Meta))
	// 	if err := f.Error(); err != raft.ErrUnknownPeer && err != nil {
	// 		panic(err)
	// 	}
	// }
}

func (s *State) Peers() []string {
	p, _ := s.peerStore.Peers()
	return p
}

func (s *State) Members() []*memberlist.Node {
	return s.ml.Members()
}

func (s *State) NotifyUpdate(node *memberlist.Node) {
	fmt.Println("NotifyUpdate")
}

func (s *State) NodeMeta(limit int) []byte {
	fmt.Println("NodeMeta", limit)
	return []byte(s.raftPort)
}

func (s *State) NotifyMsg(msg []byte) {
	fmt.Println("NotifyMsg", string(msg))
	if s.IsLeader() {
		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			panic(err)
		}
		s.Set(message.Key, message.Value)
	}
}

func (s *State) GetBroadcasts(overhead, limit int) [][]byte {
	return s.broadcast.GetBroadcasts(overhead, limit)
}

func (s *State) LocalState(join bool) []byte {
	fmt.Println("LocalState", join)
	return []byte("")
}

func (s *State) MergeRemoteState(buf []byte, join bool) {
	fmt.Println("MergeRemoteState", string(buf), join)
}

func (s *State) Apply(l *raft.Log) interface{} {
	fmt.Println("Apply", l.Type)
	var msg Message
	if err := json.Unmarshal(l.Data, &msg); err != nil {
		panic(err)
	}
	s.kvs[msg.Key] = msg.Value
	return nil
}

func (s *State) Snapshot() (raft.FSMSnapshot, error) {
	fmt.Println("SnapShot")
	return &SnapShot{s.kvs}, nil
}

func (s *State) Restore(rc io.ReadCloser) error {
	fmt.Println("Restore")
	defer rc.Close()
	var kvs map[string]string
	if err := json.NewDecoder(rc).Decode(kvs); err != nil {
		return err
	}
	s.kvs = kvs
	return nil
}

func (s *State) StartRaft(addr string, port int, bootstrap bool) {
	s.raftPort = strconv.Itoa(port)
	bind := addr + ":" + s.raftPort
	a, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		panic(err)
	}
	transport, err := raft.NewTCPTransport(bind, a, 3, time.Second, os.Stderr)
	if err != nil {
		panic(err)
	}
	peerStore := raft.NewJSONPeers("/", transport)
	s.peerStore = peerStore
	logStore := raft.NewInmemStore()
	snapshotStore := raft.NewDiscardSnapshotStore()
	config := raft.DefaultConfig()
	if bootstrap {
		config.EnableSingleNode = true
	}
	rft, err := raft.NewRaft(config, s, logStore, logStore, snapshotStore, peerStore, transport)
	if err != nil {
		panic(err)
	}
	s.raft = rft
}

func (s *State) StartMemberlist(addr string, port int, joinAddr string) {
	mlConfig := memberlist.DefaultLANConfig()
	mlConfig.BindAddr = addr
	mlConfig.BindPort = port
	mlConfig.Name = addr + ":" + strconv.Itoa(port)
	mlConfig.Delegate = s
	mlConfig.Events = s
	ml, err := memberlist.Create(mlConfig)
	if err != nil {
		panic(err)
	}
	s.broadcast.NumNodes = ml.NumMembers

	if joinAddr != "" {
		_, err := ml.Join([]string{joinAddr})
		if err != nil {
			panic(err)
		}
	}

	s.ml = ml
	fmt.Println("start on", mlConfig.Name)
}

func (s *State) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

func (s *State) Set(key, value string) {
	msg := Message{key, value}
	if s.IsLeader() {
		if err := s.raft.Apply(msg.Message(), time.Second).Error(); err != nil {
			panic(err)
		}
		return
	}
	s.broadcast.QueueBroadcast(msg)
}

func (s *State) Get(key string) string {
	return s.kvs[key]
}

func (s *State) ForEach(f func(k, v string)) {
	for k, v := range s.kvs {
		f(k, v)
	}
}

type SnapShot struct {
	kvs map[string]string
}

func (ss *SnapShot) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	if err := json.NewEncoder(sink).Encode(ss.kvs); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (ss *SnapShot) Release() {}

type Message struct {
	Key   string
	Value string
}

func (m Message) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (m Message) Message() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func (m Message) Finished() {}

type InMemoryPeerStore struct {
	peers []string
}

func (s *InMemoryPeerStore) Peers() ([]string, error) {
	return s.peers, nil
}

func (s *InMemoryPeerStore) SetPeers(peers []string) error {
	s.peers = peers
	return nil
}

func main() {
	mlPort := flag.Int("ml-port", 12345, "memberlist port")
	raftPort := flag.Int("raft-port", 54321, "raft port")
	join := flag.String("join", "", "cluster address")
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addrs, err := net.DefaultResolver.LookupHost(context.Background(), hostname)
	if err != nil {
		panic(err)
	}
	var addr string
	for _, a := range addrs {
		if net.ParseIP(a).To4() != nil {
			addr = a
			break
		}
	}

	state := &State{
		kvs:       make(map[string]string),
		broadcast: &memberlist.TransmitLimitedQueue{RetransmitMult: 3},
	}
	state.StartRaft(addr, *raftPort, *join == "")
	state.StartMemberlist(addr, *mlPort, *join)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)

	setRe := regexp.MustCompile(`set (\w+) (\w+)`)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			match := setRe.FindStringSubmatch(text)
			switch {
			case len(match) != 0:
				fmt.Println("trying set")
				state.Set(match[1], match[2])
			case text == "show":
				state.ForEach(func(k, v string) {
					fmt.Println(k, "=", v)
				})
			case text == "leader":
				fmt.Println(state.IsLeader())
			case text == "peers":
				fmt.Println(state.Peers())
			case text == "members":
				fmt.Println(state.Members())
			}
		}
	}()

	<-sig
}
