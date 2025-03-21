package main

import (
	"flag"

	bc "github.com/ignolardo/chat-grpc/src/client/client_lib"
)

func main() {
	flag.Parse()

	inCh := make(chan string)
	go bc.InitClient(inCh)

	waitc := make(chan struct{})
	<-waitc

}
