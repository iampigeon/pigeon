package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	adbHttp "github.com/arangodb/go-driver/http"
	"github.com/iampigeon/pigeon/db"
	"github.com/iampigeon/pigeon/httpsvc"
	"github.com/iampigeon/pigeon/proto"
	"github.com/iampigeon/pigeon/rpc/scheduler"
	"github.com/iampigeon/pigeon/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := flag.Int("port", 9001, "port of the service")
	host := flag.String("host", "", "host of the service")
	dbfile := flag.String("db", "messages.db", "file to store messages")
	endpoint := flag.String("endpoint", "http://arango:8529", "arangodb network address")

	redisURL := flag.String("redis_url", "redis://redis:6379/0", "URL of the redis server.")
	redisIdleTimeout := flag.Duration("redis_idle_timeout", 5*time.Second, "Timeout for redis idle connections.")
	redisDatabase := flag.Int("redis_db", 1, "Redis database to use")
	redisMaxIdle := flag.Int("redis_max_idle", 10, "Maximum number of idle connections in the pool")

	flag.Parse()

	// ----- Init DB
	conn, err := adbHttp.NewConnection(adbHttp.ConnectionConfig{
		Endpoints: []string{*endpoint},
	})
	if err != nil {
		fmt.Print("FUCK")
		log.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	dst, err := db.NewDatastore(*dbfile, conn)
	if err != nil {
		log.Fatal(err)
	}

	// ----- Init HTTP
	// TODO: implements recover
	httpServer := httpsvc.NewHTTPServer(dst)
	log.Printf("Running server on: " + httpServer.Addr)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
			return
		}
	}()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	//stores
	ms, err := db.NewMessageStore(dst)
	if err != nil {
		log.Fatal(err)
	}

	// ----- Init grpc
	s := grpc.NewServer()
	log.Printf("Starting server at %s redis_url: %s redis_db: %d database: %s\n", addr, *redisURL, *redisDatabase, *dbfile)
	proto.RegisterSchedulerServiceServer(s, schedulersvc.New(scheduler.StorageConfig{
		// BoltDatabase:     *dbfile,
		MessageStore:     ms,
		RedisURL:         *redisURL,
		RedisIdleTimeout: *redisIdleTimeout,
		RedisDatabase:    *redisDatabase,
		RedisMaxIdle:     *redisMaxIdle,
	}))

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
