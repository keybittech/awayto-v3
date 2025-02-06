package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func Asset(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get cwd: %v", err)
	}
	return filepath.Join(cwd, path)
}
func TestMain(t *testing.T) {
	println("main hello" + os.Getenv("PWD"))

	go main()

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		SlowMo:   playwright.Float(500),
		Headless: playwright.Bool(false),
	})

	originUrl := fmt.Sprintf("https://localhost:%d", httpsPort)

	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		BaseURL:           playwright.String(originUrl),
		IgnoreHttpsErrors: playwright.Bool(true),
		ColorScheme:       playwright.ColorSchemeDark,
	})

	if _, err = page.Goto(""); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	page.Locator(".header-btns > a.register-btn").Click()

	userUuid := fmt.Sprint(time.Now().UTC().UnixMilli())

	page.Locator("#email").Fill(userUuid + "@" + userUuid)
	page.Locator("#password").Fill(userUuid)
	page.Locator("#password-confirm").Fill(userUuid)
	page.Locator("#firstName").Fill(userUuid)
	page.Locator("#lastName").Fill(userUuid)

	page.GetByText("Register").Last().Click()

	browser.Close()
	pw.Stop()
}
