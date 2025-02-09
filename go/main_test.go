package main

import (
	"av3api/pkg/types"
	"flag"
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

	go main()

	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	originUrl := fmt.Sprintf("https://localhost:%d", httpsPort)
	workDir := fmt.Sprintf("%s/%s", pwd, os.Getenv("PLAYWRIGHT_CACHE_DIR"))

	browser, err := pw.Chromium.LaunchPersistentContext(workDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		BaseURL:           playwright.String(originUrl),
		IgnoreHttpsErrors: playwright.Bool(true),
		ColorScheme:       playwright.ColorSchemeDark,
		SlowMo:            playwright.Float(100),
		Headless:          playwright.Bool(false),
		Devtools:          playwright.Bool(true),
	})

	cookies, err := browser.Cookies()
	if err != nil {
		log.Fatal()
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	// goHome := func() {
	// 	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "show mobile main menu"}).Click()
	// 	page.GetByText("Home").Click()
	// }

	personA := &types.IUserProfile{
		Email:     "a@person",
		FirstName: "a",
		LastName:  "person",
	}
	pwA := personA.FirstName + personA.LastName

	randId := fmt.Sprint(time.Now().UTC().UnixMilli())

	if len(cookies) == 0 {
		if _, err = page.Goto(""); err != nil {
			log.Fatalf("could not goto: %v", err)
		}
		// Register user
		page.Locator(".header-btns > a.register-btn").Click()
		page.Locator("#email").Fill(personA.Email)
		page.Locator("#password").Fill(pwA)
		page.Locator("#password-confirm").Fill(pwA)
		page.Locator("#firstName").Fill(personA.FirstName)
		page.Locator("#lastName").Fill(personA.LastName)
		page.GetByText("Register").Last().Click()

		// Fill out group details
		page.GetByText("Create Group").WaitFor()
		page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Group Name"}).Fill("test group " + randId)
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
		page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Name"}).Fill("test schedule " + randId)
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

	} else {
		if _, err = page.Goto("/app"); err != nil {
			log.Fatalf("could not goto: %v", err)
		}
	}

	time.Sleep(2 * time.Second)

	onSignInPage, err := page.GetByText("Sign in to your account").IsVisible()
	if err != nil {
		log.Fatal(err)
	}

	onHomePage, err := page.GetByRole("menu").IsVisible()
	if err != nil {
		log.Fatal(err)
	}

	if !onSignInPage && !onHomePage {
		page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Login"}).Click()

		if !onHomePage {
			// Login
			page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Email"}).Fill(personA.Email)
			page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Password"}).Fill(pwA)
			page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign In"}).Click()
		}
	}

	// Create a form
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Admin"}).Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "forms"}).Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create"}).Click()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Name"}).Fill("test form bla")
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "add row"}).Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "add column"}).Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "add column"}).Click()
	page.GetByRole("row").GetByRole("gridcell").First().GetByRole("button").Click()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Label"}).Fill("field bla " + randId)
	page.GetByRole("row").GetByRole("gridcell").Nth(1).GetByRole("button").Click()
	page.GetByRole("combobox", playwright.PageGetByRoleOptions{Name: "Field Type Textfield"}).Click()
	page.GetByRole("option", playwright.PageGetByRoleOptions{Name: "Date"}).Click()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Label"}).Fill("date field " + randId)
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Default Value"}).Fill("2025-02-06")
	page.GetByRole("row").GetByRole("gridcell").Nth(2).GetByRole("button").Click()
	page.GetByRole("combobox", playwright.PageGetByRoleOptions{Name: "Field Type Textfield"}).Click()
	page.GetByRole("option", playwright.PageGetByRoleOptions{Name: "Time"}).Click()
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Label"}).Fill("time field")
	page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Default Value"}).Fill("17:55")
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
	page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit"}).Click()

	time.Sleep(6 * time.Second)

	browser.Close()
	pw.Stop()
}
