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

	userUuid := fmt.Sprint(time.Now().UTC().UnixMilli())

	// Register user
	page.Locator(".header-btns > a.register-btn").Click()
	page.Locator("#email").Fill(userUuid + "@" + userUuid)
	page.Locator("#password").Fill(userUuid)
	page.Locator("#password-confirm").Fill(userUuid)
	page.Locator("#firstName").Fill(userUuid)
	page.Locator("#lastName").Fill(userUuid)
	page.GetByText("Register").Last().Click()

	// Fill out group details
	page.GetByText("Create Group").WaitFor()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Group Name"}).Fill("test group " + userUuid)
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Group Description"}).Click()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Group Description"}).Fill("a group for testing")
	page.GetByRole("checkbox", playwright.PageGetByRoleOptions{Name: "Use AI Suggestions"}).Check()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Save Group"}).Click()

	// Add group roles
	page.GetByText("Create Roles").WaitFor()

	page.GetByRole("listitem").WaitFor()
	page.GetByRole("listitem").First().Click()
	page.GetByRole("listitem").Last().Click()
	page.GetByRole("combobox", playwright.PageGetByRoleOptions{Name: "Default Role"}).Click()
	page.GetByRole("listbox").Locator("li").First().Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Save Roles"}).Click()

	// Create service
	page.GetByText("Create Service").WaitFor()
	serviceNameSelections := page.GetByText("Step 1. Provide details").Locator("..")
	serviceNameSelections.GetByRole("listitem").WaitFor()
	serviceNameSelections.GetByRole("listitem").First().Click()
	tierNameSelections := page.GetByText("Step 2. Add a tier").Locator("..").GetByText("AI:").First()
	tierNameSelections.GetByRole("listitem").WaitFor()
	tierNameSelections.GetByRole("listitem").First().Click()
	featureNameSelections := page.GetByText("Step 2. Add a tier").Locator("..").GetByText("AI:").Last()
	featureNameSelections.GetByRole("listitem").WaitFor()
	featureNameSelections.GetByRole("listitem").First().Click()
	featureNameSelections.GetByRole("listitem").Last().Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add Tier to Service"}).Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Save Service"}).Click()

	// Create schedule
	page.GetByText("Create Schedule").WaitFor()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Name"}).Fill("test schedule " + userUuid)
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Start Date"}).Fill("2025-02-05")
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Save Schedule"}).Click()

	// Review group creation
	page.GetByText("Review").WaitFor()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create Group"}).Nth(1).Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Click here to confirm."}).Click()

	// Logout
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "show mobile main menu"}).WaitFor()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "show mobile main menu"}).Click()
	page.GetByText("Logout").Click()

	// Login
	page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Login"}).Click()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Email"}).Fill(userUuid + "@" + userUuid)
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Password"}).Fill(userUuid)
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign In"}).Click()

	time.Sleep(5 * time.Second)

	browser.Close()
	pw.Stop()
}
