package testutil

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func StartTestServer() (*exec.Cmd, error) {
	_, err := net.DialTimeout("tcp", fmt.Sprintf("[::]:%d", util.E_GO_HTTPS_PORT), 2*time.Second)
	if err != nil {
		cmd := exec.Command(filepath.Join(util.E_PROJECT_DIR, "go", util.E_BINARY_NAME), "-rateLimit=500", "-rateLimitBurst=500")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGKILL,
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Start()
		if err != nil {
			return nil, fmt.Errorf("Error starting server internal: %v", util.ErrCheck(err))
		}

		time.Sleep(5 * time.Second)

		return cmd, nil
	}

	return nil, fmt.Errorf("Error starting server: %v", util.ErrCheck(err))
}
