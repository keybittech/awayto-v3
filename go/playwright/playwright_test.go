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

	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/playwright-community/playwright-go"
)

var (
	aiEnabled, useRandUser bool
	browser                playwright.Browser
)

func TestMain(m *testing.M) {
	util.ParseEnv()

	useRandUser = true
	aiEnabled = false

	cmd := exec.Command(filepath.Join(util.E_PROJECT_DIR, "go", util.E_BINARY_NAME))
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

	time.Sleep(2 * time.Second)

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(util.ErrCheck(err))
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
		fmt.Printf("could not close browser: %v", util.ErrCheck(err))
	}
	if err = pw.Stop(); err != nil {
		fmt.Printf("could not stop Playwright: %v", util.ErrCheck(err))
	}

	if err := cmd.Process.Kill(); err != nil {
		fmt.Printf("Failed to close server: %v", util.ErrCheck(err))
	}

	os.Exit(code)
}

func TestPlaywright(t *testing.T) {
	testPlaywrightRegistration(t)
	testPlaywrightCreatePersonalSchedule(t)
	testPlaywrightCreateForm(t)
}

func testPlaywrightRegistration(t *testing.T) {
	page := getBrowserPage(t)
	user := getUiUser()

	if useRandUser {
		if _, err := page.Goto(""); err != nil {
			t.Fatalf("could not goto: %v", util.ErrCheck(err))
		}

		// Register user
		page.ByRole("link", "Open Registration!").MouseOver().Click()
		page.ByRole("button", "Register").WaitFor()

		// On registration page
		doEval(page)
		page.ByLocator("#email").MouseOver().Fill(user.Profile.Email)
		page.ByLocator("#password").MouseOver().Fill(user.Password)
		page.ByLocator("#password-confirm").MouseOver().Fill(user.Password)
		page.ByLocator("#firstName").MouseOver().Fill(user.Profile.FirstName)
		page.ByLocator("#lastName").MouseOver().Fill(user.Profile.LastName)

		page.ByRole("button", "Register").MouseOver().Click()

		if page.ById("input-error-email").IsVisible() {
			page.ByText("Back to Login").MouseOver().Click()

			// On login page
			login(t, page, user)
		}
	} else {
		if _, err := page.Goto("/app"); err != nil {
			t.Fatalf("could not goto: %v", util.ErrCheck(err))
		}
	}

	if page.ByText("Watch the tutorial").IsVisible() {
		doEval(page)

		println(fmt.Sprintf("Registered user %s with pass %s", user.Profile.Email, user.Password))

		page.ByRole("textbox", "Group Name").MouseOver().Fill(fmt.Sprintf("Downtown Writing Center %d", user.UserId))
		page.ByRole("textbox", "Group Description").MouseOver().Fill("Works with students and the public to teach writing")
		if aiEnabled {
			page.ByLocator(`label[id="manage_group_modal_ai"]`).MouseOver().SetChecked(true)
		}
		page.ByRole("button", "Next").MouseOver().Click()

		// Add group roles
		if aiEnabled {
			page.ByLocator(`span[id^="suggestion-"]`).Nth(0).MouseOver().Click()
			page.ByLocator(`span[id^="suggestion-"]`).Nth(1).MouseOver().Click()
		} else {
			page.ById("group_role_entry").MouseOver().Click()
			page.ById("group_role_entry").Fill("Advisor")
			page.ById("manage_group_roles_modal_add_role").MouseOver().Click()

			page.ById("group_role_entry").MouseOver().Click()
			page.ById("group_role_entry").Fill("Student")
			page.ById("manage_group_roles_modal_add_role").MouseOver().Click()
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
				err := page.Mouse().Wheel(0, 100)
				if err != nil {
					panic(util.ErrCheck(err))
				}
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
		page.ByRole("button", "Next").ScrollIntoViewIfNeeded()
		page.ByRole("button", "Next").MouseOver().Click()

		// Create schedule
		page.ByRole("textbox", "Start Date").WaitFor()
		page.ByRole("textbox", "Name").MouseOver().Fill("Fall 2025 Learning Center")
		page.ByTestId("CalendarIcon").First().MouseOver().Click()
		page.ByLocator(`button[role="gridcell"]`).First().MouseOver().Click()
		for range 2 {
			page.Mouse().Wheel(0, 300)
		}
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

		// Logout
		page.ById("topbar_open_menu").WaitFor()
		page.ById("topbar_open_menu").MouseOver().Click()
		page.ByText("Logout").MouseOver().Click()
	}

	login(t, page, user)

	if err := page.Close(); err != nil {
		log.Fatalf("failed to close page: %v", util.ErrCheck(err))
	}
}

func testPlaywrightCreatePersonalSchedule(t *testing.T) {
	page := getBrowserPage(t)
	user := getUiUser()
	login(t, page, user)

	page.ById("topbar_open_menu").WaitFor()

	createPersonalScheduleButton := page.ById("request_quote_create_personal_schedule")
	if createPersonalScheduleButton.IsVisible() {
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

func testPlaywrightCreateForm(t *testing.T) {
	page := getBrowserPage(t)
	user := getUiUser()
	login(t, page, user)

	// Create a form
	page.ByRole("button", "Admin").Click()
	page.ByRole("button", "forms").Click()

	formCount, err := page.ByRole("checkbox", "Select row").Count()
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	if formCount == 0 {
		page.ByRole("button", "Create").MouseOver().Click()
		page.ByRole("textbox", "Name").MouseOver().Fill(fmt.Sprintf("Assignment %d Intake", formCount+1))

		page.ByRole("button", "add row").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").First().ByRole("button").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Assignment Name")

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

		page.ByRole("button", "Submit").MouseOver().Click()

	}

	goHome(page)
}
