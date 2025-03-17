package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "github.com/ignolardo/chat-grpc/src/broadcast"
	"google.golang.org/grpc"
)

func main() {
	var connections []*Connection
	broadcast := &Broadcast{
		connections: connections,
	}

	listener, err := net.Listen("tcp", ":50067")

	if err != nil {
		log.Fatalf("Could not listen tcp port : %v", err)

	}

	log.Printf("Server listening at address: %v", listener.Addr())

	s := grpc.NewServer()

	pb.RegisterBroadcastServer(s, broadcast)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Could not create gRPC server : %v", err)
	}
}

type Connection struct {
	stream  pb.Broadcast_ConnectServer
	user_id string
	active  bool
	error   chan error
}

type Broadcast struct {
	pb.UnimplementedBroadcastServer
	connections []*Connection
}

func (s *Broadcast) Connect(conn *pb.Connection, stream pb.Broadcast_ConnectServer) error {
	_conn := &Connection{
		stream:  stream,
		user_id: conn.GetUser().GetId(),
		active:  conn.GetActive(),
		error:   make(chan error),
	}

	s.connections = append(s.connections, _conn)

	log.Printf("User %v connected", conn.GetUser().GetId())

	return <-_conn.error
}

func (s *Broadcast) BroadcastMessage(ctx context.Context, message *pb.Message) (*pb.Close, error) {

	wait := sync.WaitGroup{}

	//chat_id := message.GetChatId();

	for _, conn := range s.connections {
		wait.Add(1)

		go func(message *pb.Message, conn *Connection) {

			defer wait.Done()

			if conn.active {
				log.Printf("Sending message from %v to %v", message.GetUserId(), conn.user_id)
				if err := conn.stream.Send(message); err != nil {
					log.Printf("Could not send message on stream: %v. %v", conn.stream, err)
					conn.active = false
					conn.error <- err
				}

			}
		}(message, conn)
	}

	wait.Wait()

	return &pb.Close{}, nil
}
