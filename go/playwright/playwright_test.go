package main_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/playwright-community/playwright-go"
)

var (
	aiEnabled   = false
	useRandUser = false
	headless    = playwright.Bool(false)
	slowMo      = playwright.Float(100)
	browser     playwright.Browser
)

func TestMain(m *testing.M) {
	util.ParseEnv()

	cmd, err := testutil.StartTestServer()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Printf("Failed to close server: %v", util.ErrCheck(err))
		}
	}()

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
	defer testutil.TestPanic(t)

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
