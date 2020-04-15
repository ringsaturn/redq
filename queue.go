package redq

import (
	"github.com/gomodule/redigo/redis"
)

type RedQueue struct {
	id          string
	waitingList string
	pendingList string
	pool        *redis.Pool
	closed      bool
}

func (rq *RedQueue) recover() (err error) {
	conn := rq.pool.Get()
	defer conn.Close()
	for {
		_, err := redis.Bytes(conn.Do("RPOPLPUSH", rq.pendingList, rq.waitingList))
		if err == redis.ErrNil {
			break
		}
		if err != nil {
			return err
		}
	}
	return
}

func (rq *RedQueue) Close() (err error) {
	rq.closed = true
	return
}

func NewQueue(pool *redis.Pool, id string) (*RedQueue, error) {
	rq := &RedQueue{
		id:          id,
		pool:        pool,
		waitingList: id,
		pendingList: id + ":pending",
	}
	if err := rq.recover(); err != nil {
		return nil, err
	}
	return rq, nil
}
