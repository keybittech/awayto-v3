package clients

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

// type MockCommand struct {
// 	Data      string
// 	ReplyChan chan MockResponse
// }
//
// type MockRequest struct {
// 	Data string
// }
//
// type MockResponse struct {
// 	Result string
// 	Error  error
// }
//
// type MockHandler struct {
// 	CommandChan chan MockCommand
// }
//
// func (h *MockHandler) GetCommandChannel() chan<- MockCommand {
// 	return h.CommandChan
// }
//
// func MockHandlerProcessor(h *MockHandler, quit chan struct{}) {
// 	for {
// 		select {
// 		case cmd := <-h.CommandChan:
// 			go func(cmd MockCommand) {
// 				time.Sleep(10 * time.Millisecond)
// 				response := MockResponse{
// 					Result: "Processed: " + cmd.Data,
// 					Error:  nil,
// 				}
// 				cmd.ReplyChan <- response
// 			}(cmd)
// 		case <-quit:
// 			return
// 		}
// 	}
// }
//
// func TestSendCommand(t *testing.T) {
// 	handler := &MockHandler{
// 		CommandChan: make(chan MockCommand, 5),
// 	}
// 	quit := make(chan struct{})
//
// 	go MockHandlerProcessor(handler, quit)
// 	defer close(quit)
//
// 	createMockCommand := func(params MockRequest, replyChan chan MockResponse) MockCommand {
// 		return MockCommand{
// 			Data:      params.Data,
// 			ReplyChan: replyChan,
// 		}
// 	}
//
// 	tests := []struct {
// 		name    string
// 		params  MockRequest
// 		want    MockResponse
// 		wantErr bool
// 	}{
// 		{
// 			name:    "Valid command",
// 			params:  MockRequest{Data: "test-data"},
// 			want:    MockResponse{Result: "Processed: test-data", Error: nil},
// 			wantErr: false,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := SendCommand(handler, createMockCommand, tt.params)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("SendCommand() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("SendCommand() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestSendCommandTimeout(t *testing.T) {
// 	handler := &MockHandler{
// 		CommandChan: make(chan MockCommand),
// 	}
//
// 	createMockCommand := func(params MockRequest, replyChan chan MockResponse) MockCommand {
// 		return MockCommand{
// 			Data:      params.Data,
// 			ReplyChan: replyChan,
// 		}
// 	}
//
// 	originalTimeout := 5 * time.Second
//
// 	sendWithShortTimeout := func(
// 		handler CommandHandler[MockCommand, MockRequest, MockResponse],
// 		createCommand func(MockRequest, chan MockResponse) MockCommand,
// 		params MockRequest,
// 	) (MockResponse, error) {
// 		var emptyResponse MockResponse
// 		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond) // Short timeout for testing
// 		defer cancel()
// 		replyChan := make(chan MockResponse)
// 		cmd := createCommand(params, replyChan)
// 		select {
// 		case handler.GetCommandChannel() <- cmd:
// 		case <-ctx.Done():
// 			return emptyResponse, errors.New("timed out when sending command")
// 		}
// 		select {
// 		case response := <-replyChan:
// 			return response, nil
// 		case <-ctx.Done():
// 			return emptyResponse, errors.New("timed out when receiving command")
// 		}
// 	}
//
// 	_, err := sendWithShortTimeout(handler, createMockCommand, MockRequest{Data: "timeout-test"})
// 	if err == nil {
// 		t.Error("Expected timeout error, got nil")
// 	}
//
// 	// Restore the original timeout value (though it doesn't affect existing functions)
// 	_ = originalTimeout
// }
//
// func TestChannelError(t *testing.T) {
// 	type args struct {
// 		general  error
// 		response error
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := ChannelError(tt.args.general, tt.args.response); (err != nil) != tt.wantErr {
// 				t.Errorf("ChannelError(%v, %v) error = %v, wantErr %v", tt.args.general, tt.args.response, err, tt.wantErr)
// 			}
// 		})
// 	}
// }

/////////////////////////
/////////////////////////

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
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
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
	clientCount := 3
	commandsPerClient := 10
	results := make([][]MockResponse, clientCount)

	for i := 0; i < clientCount; i++ {
		clientId := strconv.Itoa(i)
		results[i] = make([]MockResponse, commandsPerClient)

		wg.Add(1)
		go func(clientIdx int, cId string) {
			defer wg.Done()
			for j := 0; j < commandsPerClient; j++ {
				params := MockRequest{
					Data: cId + "-command-" + strconv.Itoa(j),
				}
				createMockCommand := func(replyChan chan MockResponse) MockCommand {
					return MockCommand{
						Data:      params.Data,
						ReplyChan: replyChan,
						ClientId:  params.Data[:1], // Use first character as client Id for testing
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

// Benchmark the worker pool
func BenchmarkSocketWorkerPool(b *testing.B) {
	// numWorkers := 10
	clientCount := 10
	commandsPerClient := 100

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create a socket handler with worker pool
		socket := InitSocket()

		var wg sync.WaitGroup
		b.StartTimer()

		// Simulate multiple clients sending commands concurrently
		for c := 0; c < clientCount; c++ {
			clientId := "client-" + strconv.Itoa(c)
			wg.Add(1)
			go func(cId string) {
				defer wg.Done()
				for j := 0; j < commandsPerClient; j++ {
					params := interfaces.SocketRequest{
						SocketRequestParams: &types.SocketRequestParams{
							UserSub: cId,
							Targets: "target-" + strconv.Itoa(j%10),
						},
					}

					// Create command generator function
					createSocketCommand := func(replyChan chan interfaces.SocketResponse) interfaces.SocketCommand {
						return interfaces.SocketCommand{
							SocketCommandParams: &types.SocketCommandParams{
								Ty: CreateSocketTicketSocketCommand,
							},
							Request:   params,
							ReplyChan: replyChan,
						}
					}
					_, _ = SendCommand(socket, createSocketCommand)
				}
			}(clientId)
		}

		wg.Wait()
		b.StopTimer()
		socket.Close()
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
