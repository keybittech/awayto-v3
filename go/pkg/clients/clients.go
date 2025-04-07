package clients

import (
	"context"
	"errors"
	"hash/fnv"
	"sync"
	"time"
)

// CommandHandler interface defines a type that can handle commands of a specific type
type CommandHandler[Command any] interface {
	RouteCommand(cmd Command) error
}

// ClientIdentifier is an interface for commands that have a client ID
type IdentifiedCommand interface {
	GetClientId() string
}

type ResponseCommand interface {
	GetReplyChannel() interface{}
}

type CombinedCommand interface {
	IdentifiedCommand
	ResponseCommand
}

// GlobalWorkerPool is a singleton instance of worker pool
var (
	globalWorkerPool     *WorkerPool
	globalWorkerPoolOnce sync.Once
)

// ProcessFunction is a type for command processing functions
type ProcessFunction func(cmd CombinedCommand) bool

// InitGlobalWorkerPool initializes the global worker pool
func InitGlobalWorkerPool(numWorkers, bufferSize int) {
	globalWorkerPoolOnce.Do(func() {
		globalWorkerPool = newWorkerPool(numWorkers, bufferSize)
		globalWorkerPool.Start()
	})
}

// GetGlobalWorkerPool returns the global worker pool instance
func GetGlobalWorkerPool() *WorkerPool {
	return globalWorkerPool
}

var channelClosedWithoutResponse = errors.New("reply channel closed without response")
var channelTimedOutBeforeResponse = errors.New("timed out when receiving command")

// SendCommand sends a command to a handler and waits for a response with timeout
func SendCommand[Command any, Response any](
	handler CommandHandler[Command],
	createCommand func(chan Response) Command,
) (Response, error) {
	var emptyResponse Response
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a buffered channel to avoid leaks if no one reads from it
	replyChan := make(chan Response, 1)
	cmd := createCommand(replyChan)

	err := handler.RouteCommand(cmd)
	if err != nil {
		close(replyChan)
		return emptyResponse, err
	}

	// Wait for response
	select {
	case response, ok := <-replyChan:
		if !ok {
			// Channel was closed without a value
			return emptyResponse, channelClosedWithoutResponse
		}
		return response, nil
	case <-ctx.Done():
		// We don't close the channel here because the worker might still try to use it
		// It will be garbage collected eventually
		return emptyResponse, channelTimedOutBeforeResponse
	}
}

// WorkerPool manages a pool of workers to process commands
type WorkerPool struct {
	numWorkers    int
	workerQueues  []chan CombinedCommand
	wg            sync.WaitGroup
	clientToQueue map[string]int
	queueMutex    sync.RWMutex
	processFuncs  sync.Map // map[string]ProcessFunction
}

// newWorkerPool creates a new worker pool with the specified number of workers
func newWorkerPool(numWorkers int, bufferSize int) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = 1
	}

	pool := &WorkerPool{
		numWorkers:    numWorkers,
		workerQueues:  make([]chan CombinedCommand, numWorkers),
		clientToQueue: make(map[string]int),
	}

	// Initialize worker queues
	for i := 0; i < numWorkers; i++ {
		pool.workerQueues[i] = make(chan CombinedCommand, bufferSize)
	}

	return pool
}

// RegisterProcessFunction registers a process function with a given ID
func (p *WorkerPool) RegisterProcessFunction(id string, fn ProcessFunction) {
	p.processFuncs.Store(id, fn)
}

// UnregisterProcessFunction removes a process function
func (p *WorkerPool) UnregisterProcessFunction(id string) {
	p.processFuncs.Delete(id)
}

// Start launches the worker pool
func (p *WorkerPool) Start() {
	// Start workers
	p.wg.Add(p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		go p.worker(p.workerQueues[i])
	}
}

// Stop gracefully shuts down the worker pool
func (p *WorkerPool) Stop() {
	for i := 0; i < p.numWorkers; i++ {
		close(p.workerQueues[i])
	}
	p.wg.Wait()
}

// RouteCommand routes a command to the appropriate worker queue
func (p *WorkerPool) RouteCommand(cmd CombinedCommand) error {
	clientId := cmd.GetClientId()
	queueIdx := p.getQueueForClient(clientId)

	// Try to send with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case p.workerQueues[queueIdx] <- cmd:
		return nil
	case <-ctx.Done():
		return errors.New("timed out when routing command")
	}
}

func (p *WorkerPool) QueueLength() int {
	p.queueMutex.Lock()
	defer p.queueMutex.Unlock()
	return len(p.clientToQueue)
}

func (p *WorkerPool) CleanUpClientMapping(clientId string) {
	p.queueMutex.Lock()
	defer p.queueMutex.Unlock()
	for k, v := range p.clientToQueue {
		println("IN CHECK", k, v)
	}

	delete(p.clientToQueue, clientId)
}

// getQueueForClient determines which queue to use for a given client
func (p *WorkerPool) getQueueForClient(clientId string) int {
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
func (p *WorkerPool) worker(queue <-chan CombinedCommand) {
	defer p.wg.Done()
	for cmd := range queue {
		// Try each process function until one succeeds
		var handled bool
		p.processFuncs.Range(func(_, value interface{}) bool {
			if processFunc, ok := value.(ProcessFunction); ok {
				// If the process function returns true, it means it handled the command
				if processFunc(cmd) {
					handled = true
					return false // Stop iterating through process functions
				}
			}
			return true // Continue to next process function
		})

		if !handled {
			// No process function handled this command
			// We could log this or handle it in some way
			// For now, we'll just continue to the next command
		}
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
