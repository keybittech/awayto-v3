package main_test

import (
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/testutil"
)

func testPlaywrightUser(t *testing.T) {
	t.Run("admin can update user", func(tt *testing.T) {
		member3 := testutil.IntegrationTest.TestUsers[6]
		page := login(t, "admin")

		page.ById("home_available_role_actions_edit_group_users").MouseOver().Click()

		staffMemberRow := page.ByText(member3.GetProfile().GetEmail()).Locator.Locator("xpath=..").Locator(`[type="checkbox"]`)
		staffMemberRow.Hover()
		staffMemberRow.Click()

		page.ById("manage_users_edit").MouseOver().Click()

		page.ById("manage_user_modal_role_selection").MouseOver().Click()
		page.ByRole("listbox", "Role").ByText("Staff").MouseOver().Click()
		page.ById("mange_users_modal_submit").MouseOver().Click()
	})
}
