package testutil

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"golang.org/x/exp/mmap"
)

var (
	HandlerOptions  map[string]*util.HandlerOptions
	IntegrationTest = &IntegrationTestStruct{
		IntegrationTest: &types.IntegrationTest{},
	}
)

func init() {
	HandlerOptions = util.GenerateOptions()
}

func ResetB(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func LoadIntegrations() {
	jsonBytes, err := os.ReadFile(filepath.Join(util.E_PROJECT_DIR, "go", "integrations", "integration_results.json"))
	if err == nil {
		err = json.Unmarshal(jsonBytes, IntegrationTest)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func SaveIntegrations() {
	jsonBytes, _ := json.Marshal(IntegrationTest)
	integrationTestPath := filepath.Join(util.E_PROJECT_DIR, "go", "integrations", "integration_results.json")
	os.WriteFile(integrationTestPath, jsonBytes, 0600)
}

type IntegrationTestStruct struct {
	TestUsers map[int32]*TestUsersStruct `json:"testUsers"`
	*types.IntegrationTest
}

func (its *IntegrationTestStruct) GetTestUsers() map[int32]*TestUsersStruct {
	return its.TestUsers
}

func TestPanic(t *testing.T) func() {
	return func() {
		if r := recover(); r != nil {

			t.Log("PANIC RECOVERY ERROR: ", r)

			errLogPath := filepath.Join(util.E_PROJECT_DIR, "log", "errors.log")
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
	}
}
