package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	c "github.com/ignolardo/chat-grpc/src/client/client_lib"
	pp "github.com/ignolardo/chat-grpc/src/prompt"
)

var (
	addr = flag.String("addr", ":41797", "Client address")
)

func main() {

	flag.Parse()

	conn := c.NewGRPCClient(*addr)
	defer conn.Close()

	c := pp.NewPromptClient(conn)

	stream, err := c.Connect(context.Background())

	if err != nil {
		log.Fatalf("Could not connect to prompt")
	}

	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for {
		fmt.Print("> ")
		if scanner.Scan() {
			input = scanner.Text()
		}

		if input == ":q" {
			return
		}

		stream.Send(&pp.Text{Content: input})
	}
}
