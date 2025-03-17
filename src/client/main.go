package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/ignolardo/chat-grpc/src/broadcast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	id    = flag.String("id", "uknown", "User ID")
	name  = flag.String("name", "uknown", "User Name")
	email = flag.String("email", "uknown", "User Email")
	addr  = "localhost:50067"
)

func main() {
	flag.Parse()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not create gRPC client : %v", err)
	}
	defer conn.Close()

	c := pb.NewBroadcastClient(conn)

	user := &pb.User{
		Id:          *id,
		Name:        *name,
		Email:       *email,
		LastUpdated: timestamppb.Now(),
	}

	connection := &pb.Connection{
		User:   user,
		Active: true,
	}

	ctx := context.Background()

	stream, err := c.Connect(ctx, connection)

	if err != nil {
		log.Fatalf("Could not connect to server : %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				//close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed at receive new messages : %v", err)

			}
			fmt.Printf("%v > %v\n", msg.GetUserId(), msg.GetContent())
		}
	}()

	go func() {
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
	}()

	<-waitc

}
