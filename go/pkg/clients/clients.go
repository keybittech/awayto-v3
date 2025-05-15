package clients

import (
	"context"
	"errors"
	"hash/fnv"
	"sync"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const (
	routingTimeout = 100 * time.Millisecond
	cleanupTimeout = 5 * time.Second
)

var (
	globalWorkerPool              *WorkerPool
	globalWorkerPoolOnce          sync.Once
	channelClosedWithoutResponse  = errors.New("reply channel closed without response")
	channelTimedOutBeforeResponse = errors.New("timed out when receiving command")
	routingTimeoutError           = errors.New("timed out when routing command after" + routingTimeout.String())
)

// CommandHandler interface defines a type that can handle commands of a specific type
type CommandHandler[Command any] interface {
	RouteCommand(ctx context.Context, cmd Command) error
}

// ClientIdentifier is an interface for commands that have a client ID
type IdentifiedCommand interface {
	GetClientId() string
}

type ResponseCommand interface {
	GetReplyChannel() any
}

type CombinedCommand interface {
	IdentifiedCommand
	ResponseCommand
}

type ProcessFunction func(cmd CombinedCommand) bool

// WorkerPool manages a pool of workers to process commands
type WorkerPool struct {
	clientCleanup map[string]*time.Timer
	clientToQueue map[string]int
	processFuncs  sync.Map // map[string]ProcessFunction
	wg            sync.WaitGroup
	queueMutex    sync.RWMutex
	workerQueues  []chan CombinedCommand
	numWorkers    int
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
		clientCleanup: make(map[string]*time.Timer),
	}

	// Initialize worker queues
	for i := range numWorkers {
		pool.workerQueues[i] = make(chan CombinedCommand, bufferSize)
	}

	return pool
}

func InitGlobalWorkerPool(numWorkers, bufferSize int) {
	globalWorkerPoolOnce.Do(func() {
		globalWorkerPool = newWorkerPool(numWorkers, bufferSize)
		globalWorkerPool.Start()
	})
}

func GetGlobalWorkerPool() *WorkerPool {
	return globalWorkerPool
}

// SendCommand sends a command to a handler and waits for a response with timeout
func SendCommand[Command any, Response any](
	ctx context.Context,
	handler CommandHandler[Command],
	createCommand func(chan Response) Command,
) (Response, error) {
	var emptyResponse Response

	// Create a buffered channel to avoid leaks if no one reads from it
	replyChan := make(chan Response, 1)
	cmd := createCommand(replyChan)

	err := handler.RouteCommand(ctx, cmd)
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

func (p *WorkerPool) RegisterProcessFunction(id string, fn ProcessFunction) {
	p.processFuncs.Store(id, fn)
}

func (p *WorkerPool) UnregisterProcessFunction(id string) {
	p.processFuncs.Delete(id)
}

// worker processes commands from its assigned queue
func (p *WorkerPool) worker(queue <-chan CombinedCommand) {
	defer p.wg.Done()
	for cmd := range queue {
		// Try each process function until one succeeds
		p.processFuncs.Range(func(_, value any) bool {
			if processFunc, ok := value.(ProcessFunction); ok {
				// If the process function returns true, it means it handled the command
				if processFunc(cmd) {
					return false
				}
			}
			return true
		})
	}
}

func (p *WorkerPool) Start() {
	p.wg.Add(p.numWorkers)
	for i := range p.numWorkers {
		go p.worker(p.workerQueues[i])
	}
}

func (p *WorkerPool) Stop() {
	for i := range p.numWorkers {
		close(p.workerQueues[i])
	}
	p.wg.Wait()
}

func (p *WorkerPool) RouteCommand(ctx context.Context, cmd CombinedCommand) error {
	clientId := cmd.GetClientId()
	queueIdx := p.getQueueForClient(clientId)

	select {
	case p.workerQueues[queueIdx] <- cmd:
		return nil
	case <-ctx.Done():
		return routingTimeoutError
	}
}

func (p *WorkerPool) CleanUpClientMapping(clientId string) {
	p.queueMutex.Lock()
	defer p.queueMutex.Unlock()

	if timer, ok := p.clientCleanup[clientId]; ok {
		timer.Reset(cleanupTimeout)
		return
	}

	timer := time.AfterFunc(cleanupTimeout, func() {
		p.queueMutex.Lock()
		defer p.queueMutex.Unlock()

		delete(p.clientToQueue, clientId)
		delete(p.clientCleanup, clientId)
	})

	p.clientCleanup[clientId] = timer
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
	_, err := h.Write([]byte(clientId))
	if err != nil {
		util.ErrorLog.Println(util.ErrCheck(err))
	}
	queueIdx := int(h.Sum32()) % p.numWorkers
	p.clientToQueue[clientId] = queueIdx
	return queueIdx
}

func ChannelError(general, response error) error {
	if general != nil {
		return general
	} else if response != nil {
		return response
	}

	return nil
}
