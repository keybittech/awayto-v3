package main

import (
	"testing"
)

func testPlaywrightUser(t *testing.T) {
	t.Run("admin can update user", func(tt *testing.T) {
		page := login(t, "admin")

		page.ById("available_role_actions_edit_group_users").MouseOver().Click()

		staffMemberRow := page.ByText("jsmithstaff@myschool.edu").Locator.Locator("xpath=..").Locator(`[type="checkbox"]`)
		staffMemberRow.Hover()
		staffMemberRow.Click()

		page.ById("manage_users_edit").MouseOver().Click()

		page.ById("manage_user_modal_role_selection").MouseOver().Click()
		page.ByRole("listbox", "Role").ByText("Advisor").MouseOver().Click()
		page.ById("mange_users_modal_submit").MouseOver().Click()
	})
}
