package scheduler

import (
	"log"
	"net/url"

	"github.com/garyburd/redigo/redis"
	"github.com/oklog/ulid"
)

// TODO: Add retry logic and only panic if connection is unrecoverable.

var (
	scripts map[string]*redis.Script

	scriptsSources = map[string]string{
		"pop": `
			local result_set = redis.call('ZRANGE', 'pq:ids', 0, 0)
			if not result_set or #result_set == 0 then
				return false
			end

			redis.call('ZREMRANGEBYRANK', 'pq:ids', 0, 0)

			return result_set[1]
		`,
		"push": `
			local timestamp = ARGV[1]
			local id = ARGV[2]

			redis.call('ZADD', 'pq:ids', timestamp, id)

			return true
		`,
		"peek": `
			local result_set = redis.call('ZRANGE', 'pq:ids', 0, 0)
			if not result_set or #result_set == 0 then
				return false
			end
			return result_set[1]
		`,
	}
)

func init() {
	scripts = make(map[string]*redis.Script)

	for k, v := range scriptsSources {
		scripts[k] = redis.NewScript(0, v)
	}
}

type priorityQueue struct {
	pool interface {
		Get() redis.Conn
	}
}

func newPriorityQueue(config StorageConfig) *priorityQueue {
	log.Println(config)
	pool := &redis.Pool{
		Dial:        dial(config),
		MaxIdle:     config.RedisMaxIdle,
		IdleTimeout: config.RedisIdleTimeout,
	}

	conn := pool.Get()
	if err := conn.Err(); err != nil {
		panic(err)
	}
	conn.Close()

	return &priorityQueue{pool}
}

func (pq *priorityQueue) Push(id ulid.ULID) {
	conn := pq.pool.Get()
	defer conn.Close()

	_, err := scripts["push"].Do(conn, id.Time(), id.String())
	if err != nil {
		panic(err)
	}
}

func (pq *priorityQueue) Peek() *ulid.ULID {
	conn := pq.pool.Get()
	defer conn.Close()

	idStr, err := redis.String(scripts["peek"].Do(conn))
	if err != nil {
		if err == redis.ErrNil {
			return nil
		}
		panic(err)
	}

	id, err := ulid.Parse(idStr)
	if err != nil {
		panic(err)
	}
	return &id
}

func (pq *priorityQueue) Pop() *ulid.ULID {
	conn := pq.pool.Get()
	defer conn.Close()

	idStr, err := redis.String(scripts["pop"].Do(conn))
	if err != nil {
		log.Println("BBBB")
		panic(err)
	}
	id, err := ulid.Parse(idStr)
	if err != nil {
		log.Println("CCCC")
		panic(err)
	}
	return &id
}

func dial(config StorageConfig) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		u, err := url.Parse(config.RedisURL)
		if err != nil {
			return nil, err
		}

		conn, err := redis.Dial("tcp", u.Host)
		if err != nil {
			log.Printf("reymonkey")
			return nil, err
		}
		if _, err = conn.Do("SELECT", config.RedisDatabase); err != nil {
			conn.Close()
			return nil, err
		}

		if u.User == nil {
			return conn, nil
		}
		if pw, ok := u.User.Password(); ok {
			if _, err = conn.Do("AUTH", pw); err != nil {
				conn.Close()
				return nil, err
			}
		}
		return conn, nil
	}
}
