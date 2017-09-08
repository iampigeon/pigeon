package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/WiseGrowth/pigeon/pigeon/proto"
	"github.com/WiseGrowth/pigeon/rpc/scheduler"
	"github.com/WiseGrowth/pigeon/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := flag.Int("port", 5050, "port of the service")
	host := flag.String("host", "", "host of the service")
	dbfile := flag.String("db", "massages.db", "file to store messages")

	redisURL := flag.String("redis_url", "redis://127.0.0.1:6379", "URL of the redis server.")
	redisIdleTimeout := flag.Duration("redis_idle_timeout", 5*time.Second, "Timeout for redis idle connections.")
	redisDatabase := flag.Int("redis_db", 1, "Redis database to use")
	redisMaxIdle := flag.Int("redis_max_idle", 10, "Maximum number of idle connections in the pool")

	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	log.Printf("Starting server at %s redis_url: %s redis_db: %d database: %s\n", addr, *redisURL, *redisDatabase, *dbfile)

	proto.RegisterSchedulerServiceServer(s, schedulersvc.New(scheduler.StorageConfig{
		BoltDatabase:     *dbfile,
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
