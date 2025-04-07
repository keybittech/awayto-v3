package clients

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

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
func (h *MockHandler) RouteCommand(cmd MockCommand) error {
	// Cast to CombinedCommand and route through global worker pool
	return GetGlobalWorkerPool().RouteCommand(cmd)
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
				response, err := SendCommand(handler1, createMockCommand)
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
				response, err := SendCommand(handler2, createMockCommand)
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
