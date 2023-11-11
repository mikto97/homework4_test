package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	proto "grpc/grpc"

	"google.golang.org/grpc"
)

type Server struct {
	processID   string
	port        int
	peers       []string
	peerClients map[string]proto.NodeClient
	mu          sync.Mutex
	replyCount  int
	requested   bool
	timestamp   int64

	proto.UnimplementedNodeServer
}

func NewServer(processID string, port int, peers []string) *Server {
	return &Server{
		processID:   processID,
		port:        port,
		peers:       peers,
		peerClients: make(map[string]proto.NodeClient),
	}
}

func (s *Server) RequestAccess(ctx context.Context, in *proto.Request) (*proto.Reply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Node %s requesting access to the critical section at Lamport timestamp %d", s.processID, s.timestamp)

	// Update the timestamp based on Lamport clock
	s.timestamp = max(s.timestamp, in.Timestamp) + 1

	s.requested = true
	s.replyCount = 0

	for _, peer := range s.peers {
		log.Printf("Trying to connect to peer: %s", peer)

		client, ok := s.peerClients[peer]
		if !ok {
			conn, err := grpc.Dial(peer, grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to peer %s: %v", peer, err)
				continue
			}
			s.peerClients[peer] = proto.NewNodeClient(conn)
			client = s.peerClients[peer]
		}

		go func(peer string, client proto.NodeClient) {
			// Include Lamport timestamp in the request
			reply, err := client.ReplyToRequest(context.Background(), &proto.Reply{
				Timestamp: s.timestamp,
			})
			if err != nil {
				log.Printf("Error sending reply to peer %s: %v", peer, err)
				return
			}
			s.mu.Lock()
			defer s.mu.Unlock()
			s.replyCount++
			if s.replyCount == len(s.peers)-1 {
				s.EnterCriticalSection()
			}
			log.Printf("Received reply from peer %s: %+v", peer, reply)
		}(peer, client)

	}

	return &proto.Reply{Timestamp: s.timestamp}, nil
}

func (s *Server) ReplyToRequest(ctx context.Context, in *proto.Reply) (*proto.Reply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Node %s received request from Node %s at Lamport timestamp %d", s.processID, in.GetProcessId(), in.Timestamp)

	// Update the timestamp based on Lamport clock
	s.timestamp = max(s.timestamp, in.Timestamp) + 1

	if s.requested && !s.inCriticalSection() {
		go func() {
			s.EnterCriticalSection()
		}()
	}

	return &proto.Reply{
		Timestamp: s.timestamp,
	}, nil
}

func (s *Server) EnterCriticalSection() {

	// Increment the timestamp when entering the critical section
	s.timestamp++
	fmt.Printf("Node %s entering the critical section at Lamport timestamp %d\n", s.processID, s.timestamp)
	time.Sleep(2 * time.Second) // Simulate critical section work
	fmt.Printf("Node %s leaving the critical section\n", s.processID)
	s.requested = false
}

func (s *Server) inCriticalSection() bool {
	return s.requested && s.replyCount == len(s.peers)-1
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	flag.Parse()

	processID := flag.Arg(0)
	port, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		log.Fatal("Invalid port number:", err)
	}

	peers := flag.Args()[2:]

	server := NewServer(processID, port, peers)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}
	log.Printf("Started Node %s at port: %d\n", processID, port)

	grpcServer := grpc.NewServer()
	proto.RegisterNodeServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Could not serve listener: %v", err)
		}
	}()
	log.Print("Node is now listening")

	// Simulate Node 1 initiating a request after some delay
	go func() {
		time.Sleep(5 * time.Second)
		request := &proto.Request{
			//	Timestamp: time.Now().UnixNano(),
			Timestamp: server.timestamp,
		}
		reply, err := server.RequestAccess(context.Background(), request)
		if err != nil {
			log.Printf("Error requesting access: %v", err)
			return
		}
		log.Printf("Received reply: %+v", reply)
	}()

	select {}
}
