package main_test

import (
	"testing"
)

func testPlaywrightCreateQuote(t *testing.T) {
	page := login(t, "member2")

	doEval(page)

	page.ByTestId("CalendarIcon").MouseOver().Click()
	page.ByLocator("[role='gridcell'][tabindex='0']").MouseOver().Click()
	page.ById("select_time_picker_select").MouseOver().Click()
	page.ByRole("listbox", "Time").ByLocator("li").Nth(2).MouseOver().Click()
	page.ByText("Submit Request").MouseOver().Click()
	page.ByText("Click here to confirm.").MouseOver().Click()

	page.Close(t)
}
