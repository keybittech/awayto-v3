package clients

import (
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

type MockRequest struct {
	Data string
}

type MockResponse struct {
	Result string
	Error  error
}

type MockHandler struct {
	pool *WorkerPool[MockCommand, MockRequest, MockResponse]
}

// NewMockHandler creates a new mock handler with a worker pool
func NewMockHandler(numWorkers, bufferSize int) *MockHandler {
	processFunc := func(cmd MockCommand) {
		time.Sleep(20 * time.Millisecond)
		response := MockResponse{
			Result: "Processed: " + cmd.Data + " for client: " + cmd.ClientId,
			Error:  nil,
		}
		cmd.ReplyChan <- response
	}

	pool := NewWorkerPool[MockCommand, MockRequest, MockResponse](numWorkers, bufferSize, processFunc)
	pool.Start()

	return &MockHandler{pool: pool}
}

// GetCommandChannel returns the channel for sending commands
func (h *MockHandler) GetCommandChannel() chan<- MockCommand {
	return h.pool.GetCommandChannel()
}

// Close gracefully shuts down the mock handler worker pool
func (h *MockHandler) Close() {
	h.pool.Stop()
}

// Test the worker pool implementation
func TestWorkerPool(t *testing.T) {
	// Create a mock handler with 5 workers and a buffer of 10 commands
	handler := NewMockHandler(5, 10)
	defer handler.Close()

	// Simulate 3 clients sending 10 commands each
	var wg sync.WaitGroup
	clientCount := 30
	commandsPerClient := 30
	results := make([][]MockResponse, clientCount)

	for i := 0; i < clientCount; i++ {
		clientId := strconv.Itoa(i)
		results[i] = make([]MockResponse, commandsPerClient)

		wg.Add(1)
		go func(clientIdx int, cId string) {
			defer wg.Done()
			for j := 0; j < commandsPerClient; j++ {
				createMockCommand := func(replyChan chan MockResponse) MockCommand {
					return MockCommand{
						Data:      cId + "-command-" + strconv.Itoa(j),
						ReplyChan: replyChan,
						ClientId:  cId, // Use first character as client Id for testing
					}
				}
				response, err := SendCommand(handler, createMockCommand)
				if err != nil {
					t.Errorf("SendCommand failed for client %s, command %d: %v", cId, j, err)
				} else {
					results[clientIdx][j] = response
				}
			}
		}(i, clientId)
	}

	wg.Wait()

	// Verify that commands for each client were processed in order
	for i := 0; i < clientCount; i++ {
		clientId := strconv.Itoa(i)
		for j := 0; j < commandsPerClient; j++ {
			expected := "Processed: " + clientId + "-command-" + strconv.Itoa(j) + " for client: " + clientId
			if results[i][j].Result != expected {
				t.Errorf("Unexpected result for client %s, command %d: got %s, want %s",
					clientId, j, results[i][j].Result, expected)
			}
		}
	}
}

// // Example of using the worker pool with websocket clients
// func ExampleSocketWithWebsockets() {
// 	// Create a socket with 10 workers and a buffer for 1000 commands
// 	socket := NewSocket(10, 1000)
// 	defer socket.Close()
//
// 	// Example of command creation function for a websocket client
// 	createSocketCommand := func(params SocketRequest, replyChan chan SocketResponse) SocketCommand {
// 		return SocketCommand{
// 			Ty:        CreateSocketTicketSocketCommand, // Assume Ty is added to SocketRequest
// 			Request:    params,
// 			ReplyChan: replyChan,
// 			ClientId:  params.UserSub, // Use UserSub as client Id
// 		}
// 	}
//
// 	// Simulate a websocket client handler
// 	handleWebsocketClient := func(clientId string, messageCount int) {
// 		for i := 0; i < messageCount; i++ {
// 			// Simulate receiving a websocket message every 150ms
// 			time.Sleep(150 * time.Millisecond)
//
// 			// Process the message by sending a command
// 			params := SocketRequest{
// 				UserSub:      clientId,
// 				Targets:      "target-" + strconv.Itoa(i%10),
// 				MessageBytes: []byte("message-" + strconv.Itoa(i%10)),
// 			}
//
// 			response, err := SendCommand(socket, createSocketCommand, params)
// 			if err != nil {
// 				// Handle error
// 				continue
// 			}
//
// 			// Use the response to decide on next command
// 			if response.HasSub {
// 				// Do something with subscription status
// 			} else {
// 				// Request more specific data based on the ticket
// 			}
// 		}
// 	}
//
// 	// Simulate 3 concurrent websocket clients
// 	var wg sync.WaitGroup
// 	for i := 0; i < 3; i++ {
// 		wg.Add(1)
// 		clientId := "client-" + strconv.Itoa(i)
// 		go func(id string) {
// 			defer wg.Done()
// 			handleWebsocketClient(id, 5) // Each client sends 5 messages
// 		}(clientId)
// 	}
//
// 	wg.Wait()
// }
