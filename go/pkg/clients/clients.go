package clients

import (
	"context"
	"errors"
	"hash/fnv"
	"sync"
	"time"
)

// CommandHandler interface defines a type that can handle commands of a specific type
type CommandHandler[Command any, Params any, Response any] interface {
	GetCommandChannel() chan<- Command
}

// ClientIdentifier is an interface for commands that have a client ID
type IdentifiedCommand interface {
	GetClientId() string
}

// SendCommand sends a command to a handler and waits for a response with timeout
func SendCommand[Command any, Params any, Response any](
	handler CommandHandler[Command, Params, Response],
	createCommand func(Params, chan Response) Command,
	params Params,
) (Response, error) {
	var emptyResponse Response
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	replyChan := make(chan Response)
	cmd := createCommand(params, replyChan)
	select {
	case handler.GetCommandChannel() <- cmd:
	case <-ctx.Done():
		return emptyResponse, errors.New("timed out when sending command")
	}
	select {
	case response := <-replyChan:
		return response, nil
	case <-ctx.Done():
		return emptyResponse, errors.New("timed out when receiving command")
	}
}

// WorkerPool manages a pool of workers to process commands
type WorkerPool[Command IdentifiedCommand, Params any, Response any] struct {
	numWorkers    int
	commandChan   chan Command
	workerQueues  []chan Command
	wg            sync.WaitGroup
	clientToQueue map[string]int
	queueMutex    sync.RWMutex
	processFunc   func(Command)
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool[Command IdentifiedCommand, Params any, Response any](
	numWorkers int,
	bufferSize int,
	processFunc func(Command),
) *WorkerPool[Command, Params, Response] {
	if numWorkers <= 0 {
		numWorkers = 1
	}

	pool := &WorkerPool[Command, Params, Response]{
		numWorkers:    numWorkers,
		commandChan:   make(chan Command, bufferSize),
		workerQueues:  make([]chan Command, numWorkers),
		clientToQueue: make(map[string]int),
		processFunc:   processFunc,
	}

	// Initialize worker queues
	for i := 0; i < numWorkers; i++ {
		pool.workerQueues[i] = make(chan Command, bufferSize)
	}

	return pool
}

// Start launches the worker pool
func (p *WorkerPool[Command, Params, Response]) Start() {
	// Start command router
	go p.routeCommands()

	// Start workers
	p.wg.Add(p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		go p.worker(p.workerQueues[i])
	}
}

// Stop gracefully shuts down the worker pool
func (p *WorkerPool[Command, Params, Response]) Stop() {
	close(p.commandChan)
	p.wg.Wait()
}

// GetCommandChannel returns the command channel for sending commands
func (p *WorkerPool[Command, Params, Response]) GetCommandChannel() chan<- Command {
	return p.commandChan
}

// routeCommands routes incoming commands to the appropriate worker queue
func (p *WorkerPool[Command, Params, Response]) routeCommands() {
	for cmd := range p.commandChan {
		clientId := cmd.GetClientId()
		queueIdx := p.getQueueForClient(clientId)
		p.workerQueues[queueIdx] <- cmd
	}

	// Close all worker queues when command channel is closed
	for i := 0; i < p.numWorkers; i++ {
		close(p.workerQueues[i])
	}
}

// getQueueForClient determines which queue to use for a given client
func (p *WorkerPool[Command, Params, Response]) getQueueForClient(clientId string) int {
	p.queueMutex.RLock()
	if queueIdx, exists := p.clientToQueue[clientId]; exists {
		p.queueMutex.RUnlock()
		return queueIdx
	}
	p.queueMutex.RUnlock()

	// Client not seen before, assign to a queue based on hash
	p.queueMutex.Lock()
	defer p.queueMutex.Unlock()

	// Check again in case another goroutine assigned it while we were waiting for the lock
	if queueIdx, exists := p.clientToQueue[clientId]; exists {
		return queueIdx
	}

	// Assign based on hash to ensure even distribution
	h := fnv.New32()
	h.Write([]byte(clientId))
	queueIdx := int(h.Sum32()) % p.numWorkers
	p.clientToQueue[clientId] = queueIdx
	return queueIdx
}

// worker processes commands from its assigned queue
func (p *WorkerPool[Command, Params, Response]) worker(queue <-chan Command) {
	defer p.wg.Done()
	for cmd := range queue {
		p.processFunc(cmd)
	}
}

func ChannelError(general, response error) error {
	if general != nil {
		return general
	} else if response != nil {
		return response
	}

	return nil
}
