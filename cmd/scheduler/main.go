package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/WiseGrowth/pigeon/pigeon/proto"
	"github.com/WiseGrowth/pigeon/rpc/scheduler"
	"github.com/WiseGrowth/pigeon/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	yaml "gopkg.in/yaml.v2"
)

// Config define all env values used for service execution
type Config struct {
	Port             string `yaml:"port"`
	Host             string `yaml:"host"`
	DbFile           string `yaml:"db_file"`
	RedisURL         string `yaml:"redis_url"`
	RedisIdleTimeout string `yaml:"redis_idle_timeout"`
	RedisDatabase    string `yaml:"redis_database"`
	RedisMaxIdle     string `yaml:"redis_max_idle"`
}

func main() {
	var config Config

	configPath := flag.String("path", "config.yml", "path with config file")
	flag.Parse()

	file, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Print(err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		fmt.Println(err)
	}

	redisIdleTimeout, err := strconv.Atoi(config.RedisIdleTimeout)
	if err != nil {
		fmt.Println(err)
	}

	redisDatabase, err := strconv.Atoi(config.RedisDatabase)
	if err != nil {
		fmt.Println(err)
	}

	redisMaxIdle, err := strconv.Atoi(config.RedisMaxIdle)
	if err != nil {
		fmt.Println(err)
	}

	addr := fmt.Sprintf("%s:%d", config.Host, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	log.Printf("Starting server at %s redis_url: %s redis_db: %d database: %s\n", addr, config.RedisURL, redisDatabase, config.DbFile)

	proto.RegisterSchedulerServiceServer(s, schedulersvc.New(scheduler.StorageConfig{
		BoltDatabase:     config.DbFile,
		RedisURL:         config.RedisURL,
		RedisIdleTimeout: time.Duration(redisIdleTimeout) * time.Second,
		RedisDatabase:    redisDatabase,
		RedisMaxIdle:     redisMaxIdle,
	}))
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
