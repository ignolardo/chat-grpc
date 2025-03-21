package clientlib

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/ignolardo/chat-grpc/src/broadcast"
	pp "github.com/ignolardo/chat-grpc/src/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	id          = flag.String("id", "no id", "User ID")
	name        = flag.String("name", "no name", "User Name")
	email       = flag.String("email", "no email", "User Email")
	server_addr = "localhost:41796"
	prompt_addr = flag.String("prompt_addr", "localhost:41797", "User prompt address")
)

var (
	user *pb.User
)

func InitClient(input_channel chan string) {
	ParseUser()

	conn := NewGRPCClient(server_addr)
	defer conn.Close()
	c := pb.NewBroadcastClient(conn)

	stream := ConnectToBroadcastServer(c)
	go ListenIncomingMessages(stream)

	go InitPromptServer(input_channel)

	go func() {
		for {
			message := &pb.Message{
				Id:        "_",
				UserId:    *id,
				ChatId:    "_",
				Content:   <-input_channel,
				Timestamp: timestamppb.Now(),
			}
			_, err := c.BroadcastMessage(context.Background(), message)
			if err != nil {
				log.Fatalf("Error at sending message : %v", err)
			}
		}
	}()

	waitc := make(chan struct{})
	<-waitc

	/* go func() {
		count := 1
		for {
			message := &pb.Message{
				Id:        "_",
				UserId:    *id,
				ChatId:    "_",
				Content:   fmt.Sprint(count),
				Timestamp: timestamppb.Now(),
			}
			_, err := c.BroadcastMessage(context.Background(), message)
			if err != nil {
				log.Fatalf("Error at sending message : %v", err)
			}
			count++
			time.Sleep(2 * time.Second)
		}
	}() */
}

func NewGRPCClient(address string) *grpc.ClientConn {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not create gRPC client : %v", err)
	}
	return conn
}

func InitPromptServer(input_channel chan string) {
	prompt_server := &Prompt{ch: input_channel}
	listener, err := net.Listen("tcp", *prompt_addr)
	if err != nil {
		log.Fatalf("Could not listen tcp port : %v", err)

	}

	log.Printf("Server listening at address: %v", listener.Addr())

	s := grpc.NewServer()

	pp.RegisterPromptServer(s, prompt_server)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Could not create gRPC server : %v", err)
	}
}

type Prompt struct {
	pp.UnimplementedPromptServer
	ch chan<- string
}

func (p *Prompt) Connect(stream pp.Prompt_ConnectServer) error {
	for {
		input, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pp.Close{})
		}
		if err != nil {
			return err
		}
		p.ch <- input.GetContent()
	}
}

func ConnectToBroadcastServer(c pb.BroadcastClient) grpc.ServerStreamingClient[pb.Message] {
	connection := &pb.Connection{
		User:   user,
		Active: true,
	}
	stream, err := c.Connect(context.Background(), connection)
	if err != nil {
		log.Fatalf("Could not connect to server : %v", err)
	}
	return stream
}

func ListenIncomingMessages(stream grpc.ServerStreamingClient[pb.Message]) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("Failed at receive new messages : %v", err)

		}
		fmt.Printf("%v > %v\n", msg.GetUserId(), msg.GetContent())
	}
}

func ParseUser() {
	user = &pb.User{
		Id:          *id,
		Name:        *name,
		Email:       *email,
		LastUpdated: timestamppb.Now(),
	}
}
