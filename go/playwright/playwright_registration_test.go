package main_test

import (
	"testing"
)

func testPlaywrightRegistration(t *testing.T) {
	var groupCode string
	var firstRun bool

	t.Run("admin can register and create a group", func(tt *testing.T) {
		page := login(t, "videotest_admin")

		// Login as the admin
		// If we haven't registered before, go through the full process of user and group registration
		// If we're on the inner app screens, then delete the group and go back to group registration

		if page.ById("input-error").IsVisible() {
			firstRun = true
			register(page)
		} else {
			// deleteResponse, err := request.Delete("/api/v1/group")
			// if err != nil {
			// 	time.Sleep(30 * time.Second)
			// 	t.Fatalf("failed to delete group %v", err)
			// }

			page.Page.Reload()

			// if deleteResponse.Ok() {
			// 	_, err := page.Page.Evaluate("() => window.localStorage.clear()")
			// 	if err != nil {
			// 		t.Fatalf("error cleaning local storage on delete login %v", err)
			// 	}
			// 	page.Page.Reload()
			// } else {
			// 	txt, err := deleteResponse.Text()
			// 	if err != nil {
			// 		t.Fatalf("failed to read delete group response, %v", err)
			// 	}
			// 	t.Fatalf("failed to delete group on later pass, %s", txt)
			// }
		}

		err := page.ByText("Watch the tutorial").WaitFor()
		if err != nil {
			t.Fatalf("admin didn't land on registration page, firstRun: %v", firstRun)
		}

		// On the onboarding page
		doEval(page)

		// Verify group name check
		// checkNameResponse, err := readHandlerResponse[*types.CheckGroupNameResponse](func() {
		page.ByRole("textbox", "Group Name").MouseOver().Fill("Downtown Writing Center")

		// Fill out other group fields
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
					t.Fatalf("error scrolling services, err: %v", err)
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

		page.ById("home_available_role_actions_manage_group").MouseOver().Click()

		groupCodeText, err := page.ById("manage_group_home_code_input").InputValue()
		if err != nil {
			t.Fatalf("failed to read group code text, %v", err)
		}

		groupCode = groupCodeText
	})

	t.Run("staff joins on the registration page, with the group code", func(tt *testing.T) {
		page := login(t, "videotest_staff1")

		if !firstRun {

			// On logged in dashboard
			request := page.Request()
			deleteResponse, err := request.Delete("/api/v1/profile")
			if err != nil {
				t.Fatalf("failed to call delete staff %v", err)
			}

			if deleteResponse.Ok() {
				_, err := page.Page.Evaluate("() => window.localStorage.clear()")
				if err != nil {
					t.Fatalf("error cleaning local storage on delete login %v", err)
				}
				page.Page.Reload()
			} else {
				t.Fatal("failed to resolve delete staff")
			}
		}

		register(page, groupCode)
	})

	t.Run("user joins internally, with the group code", func(tt *testing.T) {
		page := login(t, "videotest_member1")

		if !firstRun {
			// On logged in dashboard
			request := page.Request()
			deleteResponse, err := request.Delete("/api/v1/profile")
			if err != nil {
				t.Fatalf("failed to call delete user %v", err)
			}

			if deleteResponse.Ok() {
				_, err := page.Page.Evaluate("() => window.localStorage.clear()")
				if err != nil {
					t.Fatalf("error cleaning local storage on delete login %v", err)
				}
				page.Page.Reload()
			} else {
				t.Fatal("failed to resolve delete user")
			}
		}

		register(page)

		page.ById("use_group_code").MouseOver().Click()
		page.ById("join_group_input_code").MouseOver().Fill(groupCode)
		page.ById("join_group_modal_submit").MouseOver().Click()
	})
}
