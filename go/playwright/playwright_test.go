package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/playwright-community/playwright-go"
)

var (
	aiEnabled   = false
	useRandUser = false
	headless    = playwright.Bool(false)
	slowMo      = playwright.Float(100)
	browser     playwright.Browser
	handlerOpts map[string]*util.HandlerOptions
)

func TestMain(m *testing.M) {
	util.ParseEnv()

	handlerOpts = util.GenerateOptions()

	_, err := net.DialTimeout("tcp", fmt.Sprintf("[::]:%d", util.E_GO_HTTPS_PORT), 2*time.Second)
	if err != nil {
		cmd := exec.Command(filepath.Join(util.E_PROJECT_DIR, "go", util.E_BINARY_NAME), "-rateLimit=500", "-rateLimitBurst=500")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGKILL,
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Start(); err != nil {
			fmt.Println("Error starting server:", util.ErrCheck(err))
			os.Exit(1)
		}

		defer func() {
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("Failed to close server: %v", util.ErrCheck(err))
			}
		}()

		time.Sleep(2 * time.Second)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: headless,
		SlowMo:   slowMo,
		Devtools: playwright.Bool(false),
		Args: []string{
			// "--start-maximized",
			"--window-position=1921,50",
			"--window-size=1600,1031",
			"--disable-session-crashed-bubble",
			"--disable-infobars",
			"--hide-crash-restore-bubble",
		},
	})

	code := m.Run()

	if err = browser.Close(); err != nil {
		fmt.Printf("could not close browser: %v", util.ErrCheck(err))
	}
	if err = pw.Stop(); err != nil {
		fmt.Printf("could not stop Playwright: %v", util.ErrCheck(err))
	}

	os.Exit(code)
}

func TestPlaywright(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			println("final recovery", fmt.Sprint(r))
		}
	}()
	for range 1 { // Create -> Delete/Create
		testPlaywrightRegistration(t)
		testPlaywrightPermission(t)
		testPlaywrightUser(t)
		// testPlaywrightRole(t)
		testPlaywrightCreatePersonalSchedule(t)
		// testPlaywrightCreateQuote(t)
		// testPlaywrightCreateForm(t)
	}
}
