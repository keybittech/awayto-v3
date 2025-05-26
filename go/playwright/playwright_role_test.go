package main

import (
	"testing"
)

func testPlaywrightRole(t *testing.T) {
	t.Run("admin can update user roles", func(tt *testing.T) {
		page := login(t, "admin")

		page.ById("home_available_role_actions_edit_group_roles").MouseOver().Click()

	})
}
