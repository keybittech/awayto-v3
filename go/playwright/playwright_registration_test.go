package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testPlaywrightRegistration(t *testing.T) {
	var groupCode string

	t.Run("admin can register and create a group", func(tt *testing.T) {
		page := login(t, "admin")
		t.Cleanup(func() {
			page.Close(t)
		})

		// Login as the admin
		// If we haven't registered before, go through the full process of user and group registration
		// If we're on the inner app screens, then delete the group and go back to group registration

		var flow string

		// On inner screens
		if page.ById("topbar_open_menu").IsVisible() {
			flow = "group deletion"
			request := page.Request()
			request.Delete("/api/v1/groups")
			page.Reload()

			// On the login screen
		} else if page.ByText("Register").IsVisible() {
			flow = "initial registration"

			// Register user
			page.ByRole("link", "Register").MouseOver().Click()
			page.ByRole("button", "Register").WaitFor()

			// On registration page
			doEval(page)
			page.ByLocator("#email").MouseOver().Fill(page.UserWithPass.Profile.Email)
			page.ByLocator("#password").MouseOver().Fill(page.UserWithPass.Password)
			page.ByLocator("#password-confirm").MouseOver().Fill(page.UserWithPass.Password)
			page.ByLocator("#firstName").MouseOver().Fill(page.UserWithPass.Profile.FirstName)
			page.ByLocator("#lastName").MouseOver().Fill(page.UserWithPass.Profile.LastName)
			page.ByRole("button", "Register").MouseOver().Click()

			println(fmt.Sprintf("Registered user %s with pass %s", page.UserWithPass.Profile.Email, page.UserWithPass.Password))
		}

		time.Sleep(2 * time.Second)

		onRegistrationPage := page.ByText("Watch the tutorial").IsVisible()
		if !onRegistrationPage {
			t.Fatalf("admin didn't land on registration page after %s flow", flow)
		}

		// On the onboarding page
		doEval(page)

		// Verify group name check
		checkNameResponse, err := readResponse[*types.CheckGroupNameResponse](func() {
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
		postGroupResponse, err := readResponse[*types.PostGroupResponse](func() {
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

		// Logout
		page.ById("topbar_open_menu").WaitFor()
		page.ById("topbar_open_menu").MouseOver().Click()
		page.ByText("Logout").MouseOver().Click()

		login(t, "admin")
	})

	t.Run("staff member joins with the group code", func(tt *testing.T) {
		// page := getBrowserPage(t, "admin")

	})
}
