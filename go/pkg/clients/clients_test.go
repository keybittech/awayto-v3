package clients

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
)

var integrationTest = &types.IntegrationTest{}

func init() {
	jsonBytes, err := os.ReadFile(filepath.Join(os.Getenv("PROJECT_DIR"), "go", "integrations", "integration_results.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = protojson.Unmarshal(jsonBytes, integrationTest)
	if err != nil {
		log.Fatal(err)
	}
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

// Mock implementation for testing
type MockCommand struct {
	Data      string
	ReplyChan chan MockResponse
	ClientId  string
}

// GetClientId implements the ClientIdentifier interface
func (cmd MockCommand) GetClientId() string {
	return cmd.ClientId
}

func (cmd MockCommand) GetReplyChannel() interface{} {
	return cmd.ReplyChan
}

type MockRequest struct {
	Data string
}

type MockResponse struct {
	Result string
	Error  error
}

type MockHandler struct {
	handlerId string
}

// NewMockHandler creates a new mock handler
func NewMockHandler(handlerId string) *MockHandler {
	// Ensure global worker pool is initialized
	InitGlobalWorkerPool(5, 10)

	// Register process function
	GetGlobalWorkerPool().RegisterProcessFunction(handlerId, func(cmd CombinedCommand) bool {
		// Check if this is our command type
		mockCmd, ok := cmd.(MockCommand)
		if !ok {
			return false
		}

		// Process the command
		time.Sleep(20 * time.Millisecond)
		response := MockResponse{
			Result: "Processed: " + mockCmd.Data + " for client: " + mockCmd.ClientId,
			Error:  nil,
		}

		mockCmd.ReplyChan <- response
		return true
	})

	return &MockHandler{handlerId: handlerId}
}

// RouteCommand implements the CommandHandler interface
func (h *MockHandler) RouteCommand(ctx context.Context, cmd MockCommand) error {
	// Cast to CombinedCommand and route through global worker pool
	return GetGlobalWorkerPool().RouteCommand(ctx, cmd)
}

// Close unregisters the handler's process function
func (h *MockHandler) Close() {
	GetGlobalWorkerPool().UnregisterProcessFunction(h.handlerId)
}

// Test the worker pool implementation
func TestGlobalWorkerPool(t *testing.T) {
	// Initialize the global worker pool
	InitGlobalWorkerPool(5, 10)

	// Create multiple handlers that will share the worker pool
	handler1 := NewMockHandler("handler1")
	handler2 := NewMockHandler("handler2")
	defer handler1.Close()
	defer handler2.Close()

	// Simulate 3 clients sending commands to each handler
	var wg sync.WaitGroup
	clientCount := 4
	commandsPerClient := 8
	results1 := make([][]MockResponse, clientCount)
	results2 := make([][]MockResponse, clientCount)

	// Send commands to first handler
	for i := 0; i < clientCount; i++ {
		clientId := fmt.Sprintf("h1-client-%d", i)
		results1[i] = make([]MockResponse, commandsPerClient)

		wg.Add(1)
		go func(clientIdx int, cId string) {
			defer wg.Done()
			for j := 0; j < commandsPerClient; j++ {
				createMockCommand := func(replyChan chan MockResponse) MockCommand {
					return MockCommand{
						Data:      cId + "-command-" + strconv.Itoa(j),
						ReplyChan: replyChan,
						ClientId:  cId,
					}
				}
				response, err := SendCommand(context.Background(), handler1, createMockCommand)
				if err != nil {
					t.Errorf("SendCommand failed for client %s, command %d: %v", cId, j, err)
				} else {
					results1[clientIdx][j] = response
				}
			}
		}(i, clientId)
	}

	// Send commands to second handler
	for i := 0; i < clientCount; i++ {
		clientId := fmt.Sprintf("h2-client-%d", i)
		results2[i] = make([]MockResponse, commandsPerClient)

		wg.Add(1)
		go func(clientIdx int, cId string) {
			defer wg.Done()
			for j := 0; j < commandsPerClient; j++ {
				createMockCommand := func(replyChan chan MockResponse) MockCommand {
					return MockCommand{
						Data:      cId + "-command-" + strconv.Itoa(j),
						ReplyChan: replyChan,
						ClientId:  cId,
					}
				}
				response, err := SendCommand(context.Background(), handler2, createMockCommand)
				if err != nil {
					t.Errorf("SendCommand failed for client %s, command %d: %v", cId, j, err)
				} else {
					results2[clientIdx][j] = response
				}
			}
		}(i, clientId)
	}

	wg.Wait()

	// Verify that commands for each client were processed in order for handler1
	for i := 0; i < clientCount; i++ {
		clientId := fmt.Sprintf("h1-client-%d", i)
		for j := 0; j < commandsPerClient; j++ {
			expected := "Processed: " + clientId + "-command-" + strconv.Itoa(j) + " for client: " + clientId
			if results1[i][j].Result != expected {
				t.Errorf("Unexpected result for handler1 client %s, command %d: got %s, want %s",
					clientId, j, results1[i][j].Result, expected)
			}
		}
	}

	// Verify that commands for each client were processed in order for handler2
	for i := 0; i < clientCount; i++ {
		clientId := fmt.Sprintf("h2-client-%d", i)
		for j := 0; j < commandsPerClient; j++ {
			expected := "Processed: " + clientId + "-command-" + strconv.Itoa(j) + " for client: " + clientId
			if results2[i][j].Result != expected {
				t.Errorf("Unexpected result for handler2 client %s, command %d: got %s, want %s",
					clientId, j, results2[i][j].Result, expected)
			}
		}
	}
}

func TestInitGlobalWorkerPool(t *testing.T) {
	type args struct {
		numWorkers int
		bufferSize int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitGlobalWorkerPool(tt.args.numWorkers, tt.args.bufferSize)
		})
	}
}

func TestGetGlobalWorkerPool(t *testing.T) {
	tests := []struct {
		name string
		want *WorkerPool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGlobalWorkerPool(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGlobalWorkerPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendCommand(t *testing.T) {
	type args struct {
		handler       CommandHandler[MockCommand]
		createCommand func(chan MockResponse) MockCommand
	}
	tests := []struct {
		name    string
		args    args
		want    MockResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, err := SendCommand(tt.args.handler, tt.args.createCommand)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("SendCommand(%v, %v) error = %v, wantErr %v", tt.args.handler, tt.args.createCommand, err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("SendCommand(%v, %v) = %v, want %v", tt.args.handler, tt.args.createCommand, got, tt.want)
			// }
		})
	}
}

func Test_newWorkerPool(t *testing.T) {
	type args struct {
		numWorkers int
		bufferSize int
	}
	tests := []struct {
		name string
		args args
		want *WorkerPool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newWorkerPool(tt.args.numWorkers, tt.args.bufferSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newWorkerPool(%v, %v) = %v, want %v", tt.args.numWorkers, tt.args.bufferSize, got, tt.want)
			}
		})
	}
}

func TestWorkerPool_RegisterProcessFunction(t *testing.T) {
	type args struct {
		id string
		fn ProcessFunction
	}
	tests := []struct {
		name string
		p    *WorkerPool
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.RegisterProcessFunction(tt.args.id, tt.args.fn)
		})
	}
}

func TestWorkerPool_UnregisterProcessFunction(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		p    *WorkerPool
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.UnregisterProcessFunction(tt.args.id)
		})
	}
}

func TestWorkerPool_Start(t *testing.T) {
	tests := []struct {
		name string
		p    *WorkerPool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.Start()
		})
	}
}

func TestWorkerPool_Stop(t *testing.T) {
	tests := []struct {
		name string
		p    *WorkerPool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.Stop()
		})
	}
}

func TestWorkerPool_RouteCommand(t *testing.T) {
	type args struct {
		cmd CombinedCommand
	}
	tests := []struct {
		name    string
		p       *WorkerPool
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.RouteCommand(context.Background(), tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("WorkerPool.RouteCommand(%v) error = %v, wantErr %v", tt.args.cmd, err, tt.wantErr)
			}
		})
	}
}

func TestWorkerPool_CleanUpClientMapping(t *testing.T) {
	type args struct {
		clientId string
	}
	tests := []struct {
		name string
		p    *WorkerPool
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.CleanUpClientMapping(tt.args.clientId)
		})
	}
}

func TestWorkerPool_getQueueForClient(t *testing.T) {
	type args struct {
		clientId string
	}
	tests := []struct {
		name string
		p    *WorkerPool
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.getQueueForClient(tt.args.clientId); got != tt.want {
				t.Errorf("WorkerPool.getQueueForClient(%v) = %v, want %v", tt.args.clientId, got, tt.want)
			}
		})
	}
}

func TestWorkerPool_worker(t *testing.T) {
	type args struct {
		queue <-chan CombinedCommand
	}
	tests := []struct {
		name string
		p    *WorkerPool
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.worker(tt.args.queue)
		})
	}
}

func TestChannelError(t *testing.T) {
	type args struct {
		general  error
		response error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ChannelError(tt.args.general, tt.args.response); (err != nil) != tt.wantErr {
				t.Errorf("ChannelError(%v, %v) error = %v, wantErr %v", tt.args.general, tt.args.response, err, tt.wantErr)
			}
		})
	}
}
