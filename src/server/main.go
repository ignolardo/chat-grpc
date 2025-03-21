package main

import (
	bs "github.com/ignolardo/chat-grpc/src/server/server_lib"
)

func main() {
	go bs.InitServer()

	waitc := make(chan struct{})
	<-waitc
}
