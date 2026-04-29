package main_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func testPlaywrightCreateForm(t *testing.T) {
	page := login(t, "admin")

	doEval(page)

	// Create a form
	page.ById("home_available_role_actions_manage_group").MouseOver().Click()
	page.ById("manage_group_home_nav_forms").MouseOver().Click()

	formCount, err := page.ByRole("checkbox", "Select row").Count()
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}

	// if formCount == 0 {
	// create new form
	page.ById("manage_forms_create").MouseOver().Click()
	page.ById("manage_form_modal_name").MouseOver().Fill(fmt.Sprintf("Assignment %d Intake", formCount+1))

	// select roles
	page.ById("manage_form_modal_group_roles_selection").MouseOver().Click()
	page.ByRole("option", "Staff").MouseOver().Click()
	page.ByRole("option", "Member").MouseOver().Click()
	page.Mouse().Click(100, 200)

	// add a row
	page.ById("form_build_add_row").MouseOver().Click()
	page.ByRole("row").ByRole("gridcell").First().ByRole("button").MouseOver().Click()
	page.ByRole("textbox", "Label").MouseOver().Fill("Assignment Name")

	// add columns
	page.ById("form_build_add_column_row_1").MouseOver().Click()
	page.ByRole("row").ByRole("gridcell").Nth(1).ByRole("button").MouseOver().Click()
	page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
	page.ByRole("option", "Date").MouseOver().Click()
	page.ByRole("textbox", "Label").MouseOver().Fill("Due Date")

	page.ById("form_build_add_column_row_1").MouseOver().Click()
	page.ByRole("row").ByRole("gridcell").Nth(2).ByRole("button").MouseOver().Click()
	page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
	page.ByRole("option", "Time").MouseOver().Click()
	page.ByRole("textbox", "Label").MouseOver().Fill("Class Start Time")
	page.ById("form_build_close_editing").MouseOver().Click()

	page.ById("manage_form_modal_submit_new_form").MouseOver().Click()

	// }

	goHome(page)

	page.Close(t)
}
