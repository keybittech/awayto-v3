package main

import "testing"

func testPlaywrightCreateQuote(t *testing.T) {
	page := login(t, "user")

	page.ByTestId("CalendarIcon").MouseOver().Click()
	page.ByRole("gridCell", "23").MouseOver().Click()
	page.ById("select_time_picker_select").MouseOver().Click()
	page.Close(t)
}
