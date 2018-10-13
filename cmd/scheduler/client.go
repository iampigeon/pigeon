//+build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/iampigeon/pigeon/proto"
	"github.com/oklog/ulid"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()

	addr := "localhost:5050"

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := proto.NewSchedulerServiceClient(conn)

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	subcmd := flag.Arg(0)
	switch subcmd {
	case "get":
		get(c, flag.Args()[1:])
	case "put":
		put(c, flag.Args()[1:])
	default:
		flag.Usage()
		os.Exit(1)
	}

}

func get(sc proto.SchedulerServiceClient, args []string) {
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	resp, err := sc.Get(context.Background(), &proto.GetRequest{
		Id: args[0],
	})
	if err != nil {
		panic(err)
	}

	id, err := ulid.Parse(resp.Message.Id)
	if err != nil {
		panic(err)
	}

	t := time.Unix(
		int64(id.Time()/1000),
		int64((id.Time()%1000)*1000),
	)

	fmt.Printf(
		"id:%s\ncontent:%s\nendpoint:%s\ndeliver time:%s\n",
		resp.Message.Id,
		resp.Message.Content,
		resp.Message.Endpoint,
		t.Format(time.RFC3339),
	)
}

func put(sc proto.SchedulerServiceClient, args []string) {
	fs := flag.NewFlagSet("put", flag.ExitOnError)

	content := fs.String("c", "", "content of the message")
	delay := fs.Duration("d", 1*time.Second, "delay when to send the message")
	endpoint := fs.String("a", "http://localhost:5000", "address of the backend service to handle message")

	fs.Parse(args)

	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	id, err := ulid.New(
		ulid.Timestamp(time.Now().Add(*delay)),
		entropy,
	)
	if err != nil {
		log.Fatalf("Failed to create message id, %v", err)
	}

	_, err = sc.Put(context.Background(), &proto.PutRequest{
		Id:       id.String(),
		Content:  []byte(*content),
		Endpoint: *endpoint,
	})
	if err != nil {
		log.Fatalf("Put message failed, %v", err)
	}

	fmt.Printf("id:%s\n", id.String())
}
