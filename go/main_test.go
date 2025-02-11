package main

import (
	"av3api/pkg/types"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func MoveToBoundingBox(loc playwright.Locator) playwright.Locator {
	rect, _ := loc.BoundingBox()
	if rect != nil {
		centerX := rect.X + (rect.Width / 2)
		centerY := rect.Y + (rect.Height / 2)

		page, _ := loc.Page()
		page.Mouse().Move(centerX, centerY)
	}
	return loc
}

type Locator struct {
	playwright.Locator
}

func (p Locator) ByRole(role string, name ...string) Locator {
	opts := playwright.LocatorGetByRoleOptions{}
	if len(name) > 0 {
		opts.Name = name[0]
	}
	return Locator{p.GetByRole(playwright.AriaRole(role), opts)}
}

func (l Locator) MouseOver() Locator {
	err := l.Hover()
	if err != nil {
		log.Fatal(err)
	}
	return l
}

func (l Locator) Nth(num int) Locator {
	return Locator{l.Locator.Nth(num)}
}

func (l Locator) First() Locator {
	return Locator{l.Locator.First()}
}

func (l Locator) Last() Locator {
	return Locator{l.Locator.Last()}
}

func (l Locator) ByLocator(selector string) Locator {
	return Locator{l.Locator.Locator(selector)}
}

func (l Locator) ByText(text string) Locator {
	return Locator{l.Locator.GetByText(text)}
}

type Page struct {
	playwright.Page
}

func (p *Page) ByRole(role string, name ...string) Locator {
	opts := playwright.PageGetByRoleOptions{}
	if len(name) > 0 {
		opts.Name = name[0]
	}
	return Locator{p.GetByRole(playwright.AriaRole(role), opts)}
}

func (p *Page) ByLabel(label string) Locator {
	return Locator{p.GetByLabel(label)}
}

func (p *Page) ByText(text string) Locator {
	return Locator{p.GetByText(text)}
}

func (p *Page) ByLocator(selector string) Locator {
	return Locator{p.Locator(selector)}
}

func TestMain(t *testing.T) {

	go main()

	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	originUrl := fmt.Sprintf("https://localhost:%d", httpsPort)
	workDir := fmt.Sprintf("%s/%s", pwd, os.Getenv("PLAYWRIGHT_CACHE_DIR"))

	browser, err := pw.Chromium.LaunchPersistentContext(workDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		BaseURL:           playwright.String(originUrl),
		IgnoreHttpsErrors: playwright.Bool(true),
		ColorScheme:       playwright.ColorSchemeDark,
		SlowMo:            playwright.Float(1000),
		Headless:          playwright.Bool(false),
		Devtools:          playwright.Bool(false),
		NoViewport:        playwright.Bool(true),
		Screen:            &playwright.Size{Width: 1900, Height: 1050},
		Args: []string{
			"--start-maximized",
			"--window-position=1950,20",
			"--disable-session-crashed-bubble",
			"--disable-infobars",
			"--hide-crash-restore-bubble",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	cookies, err := browser.Cookies()
	if err != nil {
		log.Fatal(err)
	}

	p, err := browser.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	page := &Page{p}

	doEval := func() {
		page.Evaluate(`
			const style = document.createElement("style");
			style.innerHTML = '.cursor-trail { position: fixed; top: 0; left: 0;	width: 40px; height: 40px; border-radius: 50%; pointer-events: none; z-index: 10000; opacity: 0.5; transition: transform 0.3s ease; background: radial-gradient(circle, rgba(255, 0, 0, 1) 10px, rgba(0, 0, 0, 0) 12px); border: 2px solid white; }';
			document.head.appendChild(style);

			const trailDot = document.createElement('div');
			trailDot.classList.add('cursor-trail');
			document.body.appendChild(trailDot);

			document.addEventListener('mousemove', (event) => {
				trailDot.style.transform = 'translate(' + (event.clientX - 20) + 'px, ' + (event.clientY - 20) + 'px)';
			});
		`)
	}

	goHome := func() {
		page.ByRole("navigation").MouseOver().Click()
		page.ByText("Home").MouseOver().Click()
	}

	personA := &types.IUserProfile{
		Email:     "jsmith@myschool.edu",
		FirstName: "John",
		LastName:  "Smith",
	}

	pwA := strings.ToLower(personA.FirstName + personA.LastName)

	if len(cookies) == 0 {
		if _, err = page.Goto(""); err != nil {
			log.Fatalf("could not goto: %v", err)
		}
		doEval()
		// time.Sleep(30 * time.Second)
		// Register user
		page.ByRole("link", "Open Registration!").MouseOver().Click()
		// MoveToBoundingBox(page.Locator(".header-btns > a.register-btn"))
		// page.Locator(".header-btns > a.register-btn").Click()
		page.ByText("Register").WaitFor()
		doEval()
		page.ByLocator("#email").MouseOver().Fill(personA.Email)
		page.ByLocator("#password").MouseOver().Fill(pwA)
		page.ByLocator("#password-confirm").MouseOver().Fill(pwA)
		page.ByLocator("#firstName").MouseOver().Fill(personA.FirstName)
		page.ByLocator("#lastName").MouseOver().Fill(personA.LastName)
		page.ByRole("button", "Register").MouseOver().Click()

		// Fill out group details
		page.ByText("Group").WaitFor()
		doEval()
		page.ByRole("textbox", "Group Name").MouseOver().Fill("Downtown Writing Center")
		page.ByRole("textbox", "Group Description").MouseOver().Fill("Works with students and the public to teach writing")
		page.ByRole("checkbox", "Use AI Suggestions").MouseOver().SetChecked(true)
		page.ByRole("button", "Save Group").MouseOver().Click()

		// Add group roles
		page.ByText("Roles").WaitFor()

		page.ByRole("listitem").WaitFor()
		page.ByRole("listitem").First().MouseOver().Click()
		page.ByRole("listitem").Last().MouseOver().Click()
		page.ByRole("combobox", "Default Role").MouseOver().Click()
		page.ByRole("listbox").ByLocator("li").First().Click()
		page.ByRole("button", "Save Roles").Click()

		// Create service
		serviceNameSelections := page.ByText("Step 1. Provide details").ByLocator("..")
		serviceNameSelections.ByRole("listitem").WaitFor()
		serviceNameSelections.ByRole("listitem").First().MouseOver().Click()

		// Select a tier name
		tierNameSelections := page.ByText("Step 2. Add a tier").ByLocator("..").ByText("AI:").First()
		tierNameSelections.ByRole("listitem").First().WaitFor()
		tierNameSelections.ByRole("listitem").First().MouseOver().Click()

		// Add features to the tier
		featureNameSelections := page.ByText("Step 2. Add a tier").ByLocator("..").ByText("AI:").Last()
		featureNameSelections.ByRole("listitem").First().WaitFor()
		featureNameSelections.ByRole("listitem").First().MouseOver().Click()
		featureNameSelections.ByRole("listitem").Nth(1).MouseOver().Click()
		page.ByRole("button", "Add Tier to Service").MouseOver().Click()

		// Add a second tier
		tierNameSelections.ByRole("listitem").Nth(1).MouseOver().Click()
		featuresBox := page.ByLabel("Features")
		featuresBox.MouseOver().Click()
		featuresBox.ByLocator("li").Nth(1).MouseOver().Click()
		featuresBox.ByLocator("li").Nth(2).MouseOver().Click()
		featureNameSelections.ByRole("listitem").Nth(2).MouseOver().Click()
		featureNameSelections.ByRole("listitem").Nth(3).MouseOver().Click()
		page.ByRole("button", "Add Tier to Service").MouseOver().Click()

		// Review and save service
		page.ByRole("button", "Save Service").ScrollIntoViewIfNeeded()
		page.ByText("Step 3").MouseOver()
		time.Sleep(2 * time.Second)
		page.ByRole("button", "Save Service").MouseOver().Click()

		// Create schedule
		page.ByRole("textbox", "Start Date").WaitFor()
		page.ByRole("textbox", "Name").MouseOver().Fill("Fall 2025 Learning Center")
		page.ByRole("textbox", "Start Date").MouseOver().Fill("2025-02-05")
		page.ByRole("button", "Save Schedule").ScrollIntoViewIfNeeded()
		page.ByRole("button", "Save Schedule").MouseOver().Click()

		// Review group creation
		page.ByText("Review").WaitFor()
		time.Sleep(2 * time.Second)
		page.ByRole("button", "Create group").MouseOver().Click()
		page.ByRole("button", "Click here to confirm.").MouseOver().Click()

		// Logout
		page.ByRole("navigation").WaitFor()
		page.ByRole("navigation").MouseOver().Click()
		page.ByText("Logout").MouseOver().Click()
	} else {
		if _, err = page.Goto("/app"); err != nil {
			log.Fatalf("could not goto: %v", err)
		}
	}

	doEval()

	time.Sleep(2 * time.Second)

	onLandingPage, err := page.ByRole("link", "Login").IsVisible()
	if err != nil {
		log.Fatal(err)
	}

	if onLandingPage {
		page.ByRole("link", "Login").MouseOver().Click()
	}

	onSignInPage, err := page.GetByText("Sign in to your account").IsVisible()
	if err != nil {
		log.Fatal(err)
	}

	if onSignInPage {
		// Login
		page.ByRole("textbox", "Email").MouseOver().Fill(personA.Email)
		page.ByRole("textbox", "Password").MouseOver().Fill(pwA)
		page.ByRole("button", "Sign In").MouseOver().Click()
	}

	doEval()

	page.ByRole("navigation").WaitFor()

	// Create a form
	page.ByRole("button", "Admin").Click()
	page.ByRole("button", "forms").Click()

	formCount, err := page.ByRole("checkbox", "Select row").Count()
	if err != nil {
		log.Fatal(err)
	}

	if formCount == 0 {
		page.ByRole("button", "Create").MouseOver().Click()
		page.ByRole("textbox", "Name").MouseOver().Fill(fmt.Sprintf("Assignment %d Intake", formCount+1))

		page.ByRole("button", "add row").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").First().ByRole("button").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Assignment Name")

		time.Sleep(2 * time.Second)

		page.ByRole("button", "add column").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").Nth(1).ByRole("button").MouseOver().Click()
		page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
		page.ByRole("option", "Date").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Due Date")
		page.ByRole("textbox", "Default Value").MouseOver().Fill("2025-02-06")

		page.ByRole("button", "add column").MouseOver().Click()
		page.ByRole("row").ByRole("gridcell").Nth(2).ByRole("button").MouseOver().Click()
		page.ByRole("combobox", "Field Type Textfield").MouseOver().Click()
		page.ByRole("option", "Time").MouseOver().Click()
		page.ByRole("textbox", "Label").MouseOver().Fill("Class Start Time")
		page.ByRole("textbox", "Default Value").MouseOver().Fill("17:55")
		page.ByRole("button", "Close").MouseOver().Click()

		time.Sleep(3 * time.Second)

		page.ByRole("button", "Submit").MouseOver().Click()

	}

	goHome()

	time.Sleep(300 * time.Second)

	browser.Close()
	pw.Stop()
}
