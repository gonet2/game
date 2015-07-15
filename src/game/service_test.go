package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	pb "proto"
	"testing"
)

const (
	address = "localhost:51000"
)

func TestGamePing(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address)
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGameServiceClient(conn)

	stream, err := c.Stream(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	const N = 10

	waitc := make(chan struct{})
	go func() {
		i := 0
		for {
			if i == N {
				close(waitc)
				return
			}
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("ping reply : %v", string(in.Message))
			i++
		}
	}()

	for i := 0; i < N; i++ {
		t.Logf("ping %v", i)
		if err := stream.Send(&pb.Game_Frame{Type: pb.Game_Ping, Message: []byte(fmt.Sprintf("%v ping", i))}); err != nil {
			t.Fatal(err)
			return
		}
	}
	<-waitc
}
