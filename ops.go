package redq

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
	"sync/atomic"
)

var ErrClosed = errors.New("queue closed")

func (rq *RedQueue) Queue(msg []byte) (err error) {
	conn := rq.pool.Get()
	defer conn.Close()
	_, err = conn.Do("RPUSH", rq.waitingList, msg)
	return
}

func (rq *RedQueue) Remove(qm QueuedMessage) (err error) {
	conn := rq.pool.Get()
	defer conn.Close()
	_, err = redis.Int(conn.Do("LREM", rq.pendingList, 1, qm.Message()))
	return
}

func (rq *RedQueue) Requeue(qm QueuedMessage) (err error) {
	conn := rq.pool.Get()
	defer conn.Close()
	if err = conn.Send("MULTI"); err != nil {
		return
	}
	if err = conn.Send("LREM", rq.pendingList, 1, qm.Message()); err != nil {
		return
	}
	if err = conn.Send("RPUSH", rq.waitingList, qm.Message()); err != nil {
		return
	}
	_, err = conn.Do("EXEC")
	return
}

func (rq *RedQueue) Get(ctx context.Context) (qm QueuedMessage, err error) {
	conn := rq.pool.Get()
	defer conn.Close()

	var raw []byte
	for ctx.Err() == nil && atomic.LoadInt32(&rq.closed) == 0 {
		raw, err = redis.Bytes(conn.Do("BRPOPLPUSH", rq.waitingList, rq.pendingList, 2))
		if err == redis.ErrNil {
			err = nil
			continue
		}
		if err != nil {
			return
		}
		qm = QueuedMessage(raw)
		return
	}
	err = ErrClosed
	return
}
