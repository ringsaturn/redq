# RedQ

RedQ provides a basic Golang abstraction for a Redis based `BRPOPLPUSH` queue.

## Tutorial

First, you need a Redis pool:

    import "github.com/gomodule/redigo/redis"

    pool := &redis.Pool{
    	MaxIdle:     3,
    	IdleTimeout: 240 * time.Second,
    	Dial: func() (redis.Conn, error) {
    		return redis.DialURL(url)
    	},
    }

Then, you can create a queue:

    import "gitlab.com/pennersr/redq"

    q, err := redq.NewQueue(pool, "my-queue")

The above results in a Redis key `my-queue` being used to store queued
messages. A key named `my-queue:pending` will be used for messages that are
pending to be processed. Specifically, `BRPOPLPUSH` pops from `my-queue` onto
`my-queue:pending`.

The producer typically queues arbitrary bytes:

    message := []byte("Hello world!")
    err := q.Queue(msg)

The consumer could look like this:

    ctx := context.Background()
    for ctx.Err() == nil {
        queuedMessage, err := q.Get(ctx)
        if err != nil {
            break
        }
        // The message has been moved onto the pending list.
        // Hence, if the consumer crashes at this point, it won't be lost.

        // Get ahold of the original bytes that were queued...
        bytes := queuedMessage.Message()

        err = processMessage(bytes)  // Your own message processor
        if err == nil {
            // Successfully processed -- remove the message from the queue.
            err = q.Remove(queuedMessage)
        } else {
            // Something went wrong. Put it back on the queue.
            err = q.Requeue(queuedMessage)
        }
    }
