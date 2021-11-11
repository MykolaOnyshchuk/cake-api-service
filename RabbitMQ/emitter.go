package RabbitMQ

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/streadway/amqp"
)

type emitterQueue struct {
	queueArr []int
}

type Message struct {
	message string
}

func NewQueue() *emitterQueue {
	queue := &emitterQueue{
		queueArr: make([]int, 0),
	}

	return queue
}

func (q *emitterQueue) Push(item int) {
	q.queueArr = append(q.queueArr, item)
}

func (queue *emitterQueue) Pop() int {
	item := queue.queueArr[0]
	queue.queueArr = queue.queueArr[1:]
	return item
}

type WorkerPool struct {
	sync.RWMutex
	id   *emitterQueue
	size int
}

func NewWorkerPool(poolSize int) *WorkerPool {
	workerPool := &WorkerPool{
		id:   NewQueue(),
		size: poolSize,
	}

	for i := 1; i <= poolSize; i++ {
		workerPool.id.Push(i)
	}

	return workerPool
}

func (workerPool *WorkerPool) getFromWorkerPool() int {
	workerPool.Lock()
	defer workerPool.Unlock()
	return workerPool.id.Pop()
}

func (workerPool *WorkerPool) pushToWorkerPool(id int) {
	workerPool.Lock()
	defer workerPool.Unlock()
	workerPool.id.Push(id)
}

func (workerPool *WorkerPool) canStart() bool {
	workerPool.RLock()
	defer workerPool.RUnlock()
	return len(workerPool.id.queueArr) != 0
}

func (workerPool *WorkerPool) startWorker(jobs <-chan Message, waitGroup *sync.WaitGroup, ch *amqp.Channel, queueName string) {
	id := workerPool.getFromWorkerPool()
	defer func() {
		workerPool.pushToWorkerPool(id)
		waitGroup.Done()
	}()

	for {
		select {
		case j := <-jobs:
			err := ch.Publish(
				"",        // exchange
				queueName, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(j.message),
				},
			)
			if err != nil {
				failOnError(err, "Failed to publish a message")
			}

			log.Printf(" [x] Sent %s", j.message)
		default:
			return
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func RunEmitter() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}
