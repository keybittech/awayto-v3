package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

var (
	aiEnabled, useRand bool
	browser            playwright.Browser
)

func init() {
	useRand = false
	aiEnabled = false
}

func TestMain(m *testing.M) {
	cmd := exec.Command(filepath.Join(os.Getenv("PROJECT_DIR"), "go", os.Getenv("BINARY_NAME")))
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

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}

	browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		SlowMo:   playwright.Float(750),
		Headless: playwright.Bool(false),
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
		fmt.Printf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		fmt.Printf("could not stop Playwright: %v", err)
	}

	if err := cmd.Process.Kill(); err != nil {
		fmt.Printf("Failed to close server: %v", err)
	}

	os.Exit(code)
}

func TestPlaywright(t *testing.T) {
	page := getBrowserPage(t)
	user := getUiUser()

	if useRand {
		if _, err := page.Goto(""); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		// Register user
		page.ByRole("link", "Open Registration!").MouseOver().Click()
		page.ByText("Register").WaitFor()

		// On registration page
		doEval(page)
		page.ByLocator("#email").MouseOver().Fill(user.Profile.Email)
		page.ByLocator("#password").MouseOver().Fill(user.Password)
		page.ByLocator("#password-confirm").MouseOver().Fill(user.Password)
		page.ByLocator("#firstName").MouseOver().Fill(user.Profile.FirstName)
		page.ByLocator("#lastName").MouseOver().Fill(user.Profile.LastName)

		page.ByRole("button", "Register").MouseOver().Click()

		time.Sleep(2 * time.Second)

		registrationFailed, err := page.ById("input-error-email").IsVisible()
		if err != nil {
			t.Fatalf("registration failed err %v", err)
		}
		if registrationFailed {
			page.ByText("Back to Login").MouseOver().Click()

			// On login page
			login(t, page, user)
		}
	} else {
		if _, err := page.Goto("/app"); err != nil {
			t.Fatalf("could not goto: %v", err)
		}
	}

	time.Sleep(2 * time.Second)

	onRegistration, err := page.ByText("Watch the tutorial").IsVisible()
	if err != nil {
		t.Fatalf("registration select err %v", err)
	}

	if onRegistration {
		doEval(page)

		page.ByRole("textbox", "Group Name").MouseOver().Fill(fmt.Sprintf("Downtown Writing Center %d", user.UserId))
		page.ByRole("textbox", "Group Description").MouseOver().Fill("Works with students and the public to teach writing")
		if aiEnabled {
			page.ByLocator(`label[id="manage_group_modal_ai"]`).MouseOver().SetChecked(true)
		}
		page.ByRole("button", "Next").MouseOver().Click()

		// Add group roles
		time.Sleep(2 * time.Second)
		page.ByText("Edit Roles").MouseOver().Click()
		if aiEnabled {
			page.ByLocator(`span[id^="suggestion-"]`).Nth(0).MouseOver().Click()
			page.ByLocator(`span[id^="suggestion-"]`).Nth(1).MouseOver().Click()
		} else {
			page.ByRole("combobox", "Group Roles").MouseOver().Click()
			page.ByLocator(`button[id="lookup_creation_toggle_group_role"]`).MouseOver().Click()
			page.ByRole("textbox", "Group Role Name").MouseOver().Fill("Advisor")
			page.ByLocator(`button[id="select_lookup_input_submit_group_role"]`).MouseOver().Click()

			page.ByRole("combobox", "Group Roles").MouseOver().Click()
			page.ByLocator(`button[id="lookup_creation_toggle_group_role"]`).MouseOver().Click()
			page.ByRole("textbox", "Group Role Name").MouseOver().Fill("Tutee")
			page.ByLocator(`button[id="select_lookup_input_submit_group_role"]`).MouseOver().Click()
		}
		page.ByRole("combobox", "Default Role").MouseOver().Click()
		page.ByRole("listbox").ByLocator("li").Last().Click()
		page.ByRole("button", "Next").MouseOver().Click()

		// Create service
		if aiEnabled {
			page.ByLocator(`span[id^="suggestion-"]`).First().WaitFor() // service name suggestion
			page.ByLocator(`span[id^="suggestion-"]`).First().MouseOver().Click()

			// Select a tier name
			page.ByLocator(`span[id^="suggestion-"]`).Nth(5).WaitFor() // tier name suggestion
			page.ByLocator(`span[id^="suggestion-"]`).Nth(5).MouseOver().Click()
			for range 2 {
				page.Mouse().Wheel(0, 100)
			}

			// Add features to the tier
			page.ByLocator(`span[id^="suggestion-"]`).Nth(10).WaitFor() // feature name suggestion
			page.ByLocator(`span[id^="suggestion-"]`).Nth(10).MouseOver().Click()
			page.ByLocator(`span[id^="suggestion-"]`).Nth(11).MouseOver().Click()
			page.ByRole("button", "Add service tier").MouseOver().Click()

			// Add a second tier
			page.ByLocator(`span[id^="suggestion-"]`).Nth(6).MouseOver().Click() // tier name suggestion
			featuresBox := page.ByRole("combobox", "Features")
			featuresBox.MouseOver().Click()
			featuresList := page.ByRole("listbox", "Features")
			featuresList.ByLocator("li").Nth(1).MouseOver().Click()
			featuresList.ByLocator("li").Nth(2).MouseOver().Click()
			page.Mouse().Click(500, 500)
			page.ByLocator(`span[id^="suggestion-"]`).Nth(12).MouseOver().Click()
			page.ByLocator(`span[id^="suggestion-"]`).Nth(13).MouseOver().Click()
		} else {
			page.ByRole("textbox", "Service Name").MouseOver().Fill("Writing Tutoring")
			page.ByRole("textbox", "Tier Name").MouseOver().Fill("Basic")
			for range 2 {
				page.Mouse().Wheel(0, 100)
			}
			featuresBox := page.ByRole("combobox", "Features")
			featuresBox.MouseOver().Click()
			page.ByLocator(`button[id="lookup_creation_toggle_feature"]`).MouseOver().Click()
			page.ByRole("textbox", "Feature Name").MouseOver().Fill("Detailed Analysis")
			page.ByLocator(`button[id="select_lookup_input_submit_feature"]`).MouseOver().Click()
			featuresBox.MouseOver().Click()
			page.ByLocator(`button[id="lookup_creation_toggle_feature"]`).MouseOver().Click()
			page.ByRole("textbox", "Feature Name").MouseOver().Fill("Helpful Feedback")
			page.ByLocator(`button[id="select_lookup_input_submit_feature"]`).MouseOver().Click()
		}

		page.ByRole("button", "Add service tier").MouseOver().Click()

		// Review and save service
		for range 2 {
			page.Mouse().Wheel(0, 100)
		}
		time.Sleep(2 * time.Second)
		page.ByRole("button", "Next").ScrollIntoViewIfNeeded()
		page.ByRole("button", "Next").MouseOver().Click()

		// Create schedule
		page.ByRole("textbox", "Start Date").WaitFor()
		page.ByRole("textbox", "Name").MouseOver().Fill("Fall 2025 Learning Center")
		page.ByTestId("CalendarIcon").First().MouseOver().Click()
		page.ByLocator(`button[role="gridcell"]`).First().MouseOver().Click()
		for range 3 {
			page.Mouse().Wheel(0, 300)
		}
		time.Sleep(2 * time.Second)
		page.ByRole("button", "Next").MouseOver().Click()

		// Review group creation
		page.ByText("Review Submission").WaitFor()
		page.ByText("Group Name").MouseOver()
		page.ByText("Service Name").MouseOver()
		page.ByText("Schedule Name").MouseOver()
		page.ByText("Review Submission").MouseOver()
		for range 2 {
			page.Mouse().Wheel(0, 200)
		}
		page.ByRole("button", "Next").MouseOver().Click()
		page.ByLocator(`button[id="confirmation_approval"]`).MouseOver().Click()

		time.Sleep(2 * time.Second)

		// Logout
		page.ById("topbar_open_menu").WaitFor()
		page.ById("topbar_open_menu").MouseOver().Click()
		page.ByText("Logout").MouseOver().Click()
	}

	time.Sleep(1 * time.Second)

	login(t, page, user)

	if err := page.Close(); err != nil {
		log.Fatalf("failed to close page: %v", err)
	}

}

func TestPlaywrightCreatePersonalSchedule(t *testing.T) {
	page := getBrowserPage(t)
	user := getUiUser()
	login(t, page, user)

	time.Sleep(time.Second)
	page.ById("topbar_open_menu").WaitFor()

	createPersonalScheduleButton := page.ById("request_quote_create_personal_schedule")
	createPersonalScheduleButtonVisible, err := createPersonalScheduleButton.IsVisible()
	if err != nil {
		t.Fatalf("create personal schedule button visible err %v", err)
	}

	if createPersonalScheduleButtonVisible {
		createPersonalScheduleButton.MouseOver().Click()
	} else {
		page.ById("available_role_actions_edit_personal_schedule").MouseOver().Click()
		page.ByLocator(`input[type="checkbox"]`).Last().MouseOver().Click()
		page.ById("manage_schedule_brackets_delete").MouseOver().Click()
		page.ByLocator(`button[id="confirmation_approval"]`).MouseOver().Click()
	}

	page.ById("manage_schedule_brackets_create").MouseOver().Click()
	page.ById("manage_schedule_brackets_modal_remaining_time").Fill("40")
	page.ByIdStartsWith("manage_personal_schedule_modal_toggle_service_").First().MouseOver().Click()
	page.ById("manage_personal_schedule_modal_next").MouseOver().Click()
	page.ByIdStartsWith("schedule_display_bracket_selection_").First().MouseOver().Click()
	page.ById("grid_cell_box_3_3").MouseOver()
	for range 4 {
		page.Mouse().Wheel(0, 150)
	}

	for _, day := range []string{"Wed", "Thu", "Fri"} {
		page.ByText(day + ", 10:00 AM").MouseOver()
		page.Mouse().Down()
		page.ByText(day + ", 10:30 AM").MouseOver()
		page.ByText(day + ", 11:00 AM").MouseOver()
		page.ByText(day + ", 11:30 AM").MouseOver()
		page.Mouse().Up()
	}

	page.ById("manage_personal_schedule_modal_submit").MouseOver().Click()

	goHome(page)
}

func TestPlaywrightCreateForm(t *testing.T) {
	page := getBrowserPage(t)
	user := getUiUser()
	login(t, page, user)

	// Create a form
	page.ByRole("button", "Admin").Click()
	page.ByRole("button", "forms").Click()

	formCount, err := page.ByRole("checkbox", "Select row").Count()
	if err != nil {
		log.Fatal(err)
	}

	if formCount == 0 {
		page.ByRole("button", "Create").MouseOver().Click()
		page.ByRole("textbox", "Name").MouseOver().Fill(fmt.Sprintf("Assignment %d Intake", formCount+1))

		page.ByRole("button", "add row").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").First().ByRole("button").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Assignment Name")

		time.Sleep(2 * time.Second)

		page.ByRole("button", "add column").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").Nth(1).ByRole("button").MouseOver().Click()
		page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
		page.ByRole("option", "Date").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Due Date")
		page.ByRole("textbox", "Default Value").MouseOver().Fill("2025-02-06")

		page.ByRole("button", "add column").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").Nth(2).ByRole("button").MouseOver().Click()
		page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
		page.ByRole("option", "Time").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Class Start Time")
		page.ByRole("textbox", "Default Value").MouseOver().Fill("17:55")
		page.ByRole("button", "Close").MouseOver().Click()

		time.Sleep(3 * time.Second)

		page.ByRole("button", "Submit").MouseOver().Click()

	}

	goHome(page)
}
