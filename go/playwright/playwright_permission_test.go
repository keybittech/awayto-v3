package main_test

import (
	"testing"
)

func testPlaywrightPermission(t *testing.T) {
	t.Run("admin can update user roles", func(tt *testing.T) {
		page := login(t, "admin")

		page.ById("home_available_role_actions_edit_group_permissions").MouseOver().Click()

		// Staff can make schedules
		page.ById("manage_role_actions_advisor_app_group_schedules").MouseOver().Click()
		page.ById("manage_role_actions_advisor_app_group_bookings").MouseOver().Click()

		// Student can request appointments
		page.ById("manage_role_actions_student_app_group_bookings").MouseOver().Click()

		// Submit
		page.ById("manage_role_actions_update_assignments").MouseOver().Click()

		goHome(page)
	})
}
