package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testPlaywrightCreateForm(t *testing.T) {
	page := login(t, "admin")

	// Create a form
	page.ById("available_role_actions_manage_group").MouseOver().Click()
	page.ByRole("button", "forms").MouseOver().Click()

	formCount, err := page.ByRole("checkbox", "Select row").Count()
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	if formCount == 0 {
		page.ByRole("button", "Create").MouseOver().Click()
		page.ByRole("textbox", "Name").MouseOver().Fill(fmt.Sprintf("Assignment %d Intake", formCount+1))

		page.ById("form_build_add_row").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").First().ByRole("button").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Assignment Name")

		page.ById("form_build_add_column_row_1").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").Nth(1).ByRole("button").MouseOver().Click()
		page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
		page.ByRole("option", "Date").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Due Date")
		page.ByRole("textbox", "Default Value").MouseOver().Fill("2025-02-06")

		page.ById("form_build_add_column_row_1").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").Nth(2).ByRole("button").MouseOver().Click()
		page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
		page.ByRole("option", "Time").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Class Start Time")
		page.ByRole("textbox", "Default Value").MouseOver().Fill("17:55")
		page.ByRole("button", "Close").MouseOver().Click()

		page.ByRole("button", "Submit").MouseOver().Click()

	}

	goHome(page)

	page.Close(t)
}
