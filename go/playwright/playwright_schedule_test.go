package main

import (
	"testing"
	"time"
)

func testPlaywrightCreatePersonalSchedule(t *testing.T) {
	page := login(t, "staff")

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
		time.Sleep(100 * time.Millisecond)
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
