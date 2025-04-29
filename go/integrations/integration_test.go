package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/exp/mmap"
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
	defer func() {
		if r := recover(); r != nil {

			t.Log("PANIC RECOVERY ERROR: ", r)

			errLogPath := filepath.Join(os.Getenv("PROJECT_DIR"), "log", "errors.log")
			reader, err := mmap.Open(errLogPath)
			if err != nil {
				t.Log(err)
				return
			}
			defer reader.Close()

			fileSize := reader.Len()
			if fileSize == 0 {
				t.Log("File is empty")
				return
			}

			data := make([]byte, fileSize)
			n, err := reader.ReadAt(data, 0)
			if err != nil && err != io.EOF {
				t.Log("Error reading file:", err)
				return
			}

			content := string(data[:n])
			lines := strings.Split(content, "\n")

			lastLineIndex := len(lines) - 1
			if lines[lastLineIndex] == "" && lastLineIndex > 0 {
				lastLineIndex--
			}

			lastLine := lines[lastLineIndex]
			t.Log("LAST LINE OF ERROR FILE: ", lastLine)
		}
	}()

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

func BenchmarkProfileDetails(b *testing.B) {
	user := integrationTest.TestUsers[0]

	client, profileDetailsRequest, err := apiBenchRequest(user.TestToken, http.MethodGet, "/api/v1/profile/details", nil, nil)
	if err != nil {
		b.Fatalf("error preparing bench request: %v", err)
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.Do(profileDetailsRequest)
		b.StopTimer()
		if err != nil {
			b.Fatalf("error doing bench request: %v", err)
		}
		b.StartTimer()
	}
}

func BenchmarkProfileDetailsParallel(b *testing.B) {
	numUsers := int32(len(integrationTest.TestUsers))
	tokens := make([]string, numUsers)
	for i := int32(0); i < numUsers; i++ {
		tokens[i] = integrationTest.TestUsers[i].TestToken
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := int32(0)

		for pb.Next() {
			tokenIndex := counter % numUsers
			counter++

			profileResponse := &types.GetUserProfileDetailsResponse{}
			err := apiRequest(tokens[tokenIndex], http.MethodGet, "/api/v1/profile/details", nil, nil, profileResponse)

			if err != nil {
				b.Fatalf("Failed to get profile details in parallel benchmark: %v", err)
			}

			if profileResponse.UserProfile == nil || profileResponse.UserProfile.Id == "" {
				b.Fatalf("Invalid profile response in parallel benchmark: empty or missing ID")
			}
		}
	})
}
