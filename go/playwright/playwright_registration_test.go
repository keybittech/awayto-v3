package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/playwright-community/playwright-go"
)

func testPlaywrightRegistration(t *testing.T) {
	var groupCode string

	t.Run("admin can register and create a group", func(tt *testing.T) {
		page := login(t, "admin")

		// Login as the admin
		// If we haven't registered before, go through the full process of user and group registration
		// If we're on the inner app screens, then delete the group and go back to group registration

		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()

		resultChan := make(chan string, 1)

		// Check if we're logged in normally
		go func() {
			err := page.ById("topbar_open_menu").WaitFor()
			if err == nil {
				resultChan <- "Internal"
			}
		}()

		// Check if login failed and we need to register
		go func() {
			err := page.ById("kc-page-title").WaitFor()
			if err == nil {
				resultChan <- "Registration"
			}
		}()

		var flow string
		// Wait for the first successful result
		select {
		case res := <-resultChan:
			flow = res
			close(resultChan)
		case <-ctx.Done():
			t.Fatalf("did not find internal page or registration page after waiting")
		}

		// On inner screens
		if flow == "Internal" {
			request := page.Request()
			deleteResponse, err := request.Delete("/api/v1/group", playwright.APIRequestContextDeleteOptions{
				Headers: page.UserWithPass.AuthorizationHeader,
			})
			if err != nil {
				t.Fatalf("failed to delete group %v", err)
			}

			if deleteResponse.Ok() {
				_, err := page.Page.Evaluate("() => window.localStorage.clear()")
				if err != nil {
					t.Fatalf("error cleaning local storage on delete login %v", err)
				}
				page.Page.Reload()
			} else {
				t.Fatal("failed to delete group on later pass")
			}
		}

		// On the login screen
		if flow == "Registration" {
			// Register user
			page.ByRole("link", "Register").MouseOver().Click()

			// On registration page
			doEval(page)
			page.ById("email").MouseOver().Fill(page.UserWithPass.Profile.Email)
			page.ById("password").MouseOver().Fill(page.UserWithPass.Password)
			page.ById("password-confirm").MouseOver().Fill(page.UserWithPass.Password)
			page.ById("firstName").MouseOver().Fill(page.UserWithPass.Profile.FirstName)
			page.ById("lastName").MouseOver().Fill(page.UserWithPass.Profile.LastName)
			page.ByRole("button", "Register").MouseOver().Click()

			println(fmt.Sprintf("Registered user %s with pass %s", page.UserWithPass.Profile.Email, page.UserWithPass.Password))
		}

		err := page.ByText("Watch the tutorial").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			t.Fatalf("admin didn't land on registration page after %s flow", flow)
		}

		// On the onboarding page
		doEval(page)

		// Verify group name check
		checkNameResponse, err := readHandlerResponse[*types.CheckGroupNameResponse](func() {
			page.ByRole("textbox", "Group Name").MouseOver().Fill(fmt.Sprintf("Downtown Writing Center %s", page.UserWithPass.UserId))
		})
		if err != nil {
			page.T.Fatalf("error when getting response for check name %v", err)
		}
		if !checkNameResponse.GetIsValid() {
			page.T.Fatal("check name returned false")
		}

		// Fill out other group fields
		page.ByRole("textbox", "Group Description").MouseOver().Fill("Works with students and the public to teach writing")
		if aiEnabled {
			page.ByLocator(`label[id="manage_group_modal_ai"]`).MouseOver().SetChecked(true)
		}

		// Verify post group response and set groupCode
		postGroupResponse, err := readHandlerResponse[*types.PostGroupResponse](func() {
			page.ByRole("button", "Next").MouseOver().Click()
		})
		if err != nil {
			page.T.Fatalf("error when getting response for posting group %v", err)
		}

		groupCode = postGroupResponse.GetCode()
		if groupCode == "" {
			page.T.Fatal("a group code was not created")
		}

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

		// time.Sleep(1 * time.Hour)
	})

	t.Run("staff joins on the registration page, with the group code", func(tt *testing.T) {
		page := login(t, "staff")
		// Register user
		page.ByRole("link", "Register").MouseOver().Click()

		// On registration page
		doEval(page)
		page.ById("groupCode").MouseOver().Fill(groupCode)
		page.ById("email").MouseOver().Fill(page.UserWithPass.Profile.Email)
		page.ById("password").MouseOver().Fill(page.UserWithPass.Password)
		page.ById("password-confirm").MouseOver().Fill(page.UserWithPass.Password)
		page.ById("firstName").MouseOver().Fill(page.UserWithPass.Profile.FirstName)
		page.ById("lastName").MouseOver().Fill(page.UserWithPass.Profile.LastName)
		page.ByRole("button", "Register").MouseOver().Click()

		println(fmt.Sprintf("Registered user %s with pass %s", page.UserWithPass.Profile.Email, page.UserWithPass.Password))
	})

	t.Run("user joins internally, with the group code", func(tt *testing.T) {
		page := login(t, "user")
		// Register user
		page.ByRole("link", "Register").MouseOver().Click()

		// On registration page
		doEval(page)
		page.ById("email").MouseOver().Fill(page.UserWithPass.Profile.Email)
		page.ById("password").MouseOver().Fill(page.UserWithPass.Password)
		page.ById("password-confirm").MouseOver().Fill(page.UserWithPass.Password)
		page.ById("firstName").MouseOver().Fill(page.UserWithPass.Profile.FirstName)
		page.ById("lastName").MouseOver().Fill(page.UserWithPass.Profile.LastName)
		page.ByRole("button", "Register").MouseOver().Click()

		println(fmt.Sprintf("Registered user %s with pass %s", page.UserWithPass.Profile.Email, page.UserWithPass.Password))

		page.ById("use_group_code").MouseOver().Click()
		page.ById("join_group_input_code").MouseOver().Fill(groupCode)
		page.ById("join_group_modal_submit").MouseOver().Click()
	})
}
