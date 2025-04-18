package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

var (
	integrationTest = &types.IntegrationTest{}
	connections     map[string]net.Conn
)

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	cmd := exec.Command(filepath.Join("..", os.Getenv("BINARY_NAME")))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	startupTicker := time.NewTicker(1 * time.Second)
	started := false

	for {
		select {
		case <-startupTicker.C:
			err := checkServer()
			if err != nil {
				continue
			}
			started = true
		default:
		}
		if started {
			break
		}
	}

	startupTicker.Stop()

	getPublicKey()

	code := m.Run()

	jsonBytes, _ := protojson.Marshal(integrationTest)
	os.WriteFile("integration_results.json", jsonBytes, 0600)

	if err := cmd.Process.Kill(); err != nil {
		fmt.Println("Failed to close server:", err)
	}

	os.Exit(code)
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
