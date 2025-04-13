package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

type TestUser struct {
	TestUserId  int
	TestToken   string
	TestTicket  string
	TestConnId  string
	UserSession *types.UserSession
	Quote       *types.IQuote
}

type IntegrationTest struct {
	TestUsers      map[int]*TestUser
	Connections    map[string]net.Conn
	Roles          map[string]*types.IRole
	MemberRole     *types.IRole
	StaffRole      *types.IRole
	Group          *types.IGroup
	MasterService  *types.IService
	GroupService   *types.IGroupService
	MasterSchedule *types.ISchedule
	GroupSchedule  *types.IGroupSchedule
	UserSchedule   *types.ISchedule
	Quote          *types.IQuote
	Booking        *types.IBooking
	DateSlots      []*types.IGroupScheduleDateSlots
}

var integrationTest = &IntegrationTest{}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	// Create a unique process group ID for the server and its children
	cmd := exec.Command("go", "run", "../main.go", "../gc.go")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Set up pipes to capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error setting up stdout pipe:", err)
		os.Exit(1)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error setting up stderr pipe:", err)
		os.Exit(1)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	// Create a channel to signal when the main test function is done
	done := make(chan struct{})

	// Setup cleanup in a separate goroutine
	go func() {
		// Wait for the test to finish
		<-done

		fmt.Println("Test finished, cleaning up server...")

		// Kill the entire process group
		pgid := cmd.Process.Pid
		if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
			fmt.Printf("Failed to kill process group: %v\n", err)
		}

		// Try normal process kill as backup
		if err := cmd.Process.Kill(); err != nil {
			fmt.Printf("Failed to kill process: %v\n", err)
		}

		// Wait for process to exit
		cmd.Process.Wait()

		// Check for any remaining processes on the port
		checkAndKillPort(7443)
	}()

	// Copy command output to the test process output in separate goroutines
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// Give the server time to initialize
	time.Sleep(2 * time.Second)

	// Run the tests
	code := m.Run()

	// Signal that tests are done
	close(done)

	// Wait a moment for cleanup to complete
	time.Sleep(1 * time.Second)

	// Exit with the appropriate code
	os.Exit(code)
}

// Helper function to check for and kill processes on a specific port
func checkAndKillPort(port int) {
	// This requires the lsof command to be available
	cmd := exec.Command("lsof", "-t", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		// lsof might return non-zero if no processes found
		return
	}

	// Split the output into lines
	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, pid := range pids {
		if pid == "" {
			continue
		}

		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			fmt.Printf("Error converting PID %s: %v\n", pid, err)
			continue
		}

		fmt.Printf("Killing process %d on port %d\n", pidInt, port)

		// Kill the process
		process, err := os.FindProcess(pidInt)
		if err != nil {
			fmt.Printf("Error finding process %d: %v\n", pidInt, err)
			continue
		}

		if err := process.Kill(); err != nil {
			fmt.Printf("Error killing process %d: %v\n", pidInt, err)
		}
	}
}

func TestIntegrations(t *testing.T) {
	testIntegrationUser(t)
	testIntegrationGroup(t)
	testIntegrationRoles(t)
	testIntegrationService(t)
	testIntegrationSchedule(t)
	testIntegrationOnboarding(t)
	testIntegrationJoinGroup(t)
	testIntegrationPromoteUser(t)
	testIntegrationUserSchedule(t)
	testIntegrationQuotes(t)
	testIntegrationBookings(t)
}
