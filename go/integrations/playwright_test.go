//go:build integration
// +build integration

package integrations

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"

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

var recording bool
var recStdout io.ReadCloser
var recs map[string]*exec.Cmd

func init() {
	recording = false
	recs = make(map[string]*exec.Cmd)
}

func killRec(title string) {
	rec := recs[title]
	output, _ := exec.Command("ps", "-o", "pid=", "--ppid", fmt.Sprintf("%d", rec.Process.Pid)).Output()
	fields := strings.Fields(string(output))
	for i := 0; i < len(fields); i++ {
		pid, err := strconv.Atoi(fields[i])
		if err != nil {
			log.Fatal(err)
		}
		println("killing spawn", pid)
		syscall.Kill(pid, syscall.SIGTERM)
	}

	println("killing", title)
	recs[title].Process.Kill()
	delete(recs, title)
}

func startRec(title string) {
	if !recording {
		return
	}

	recCmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ffmpeg -video_size 1920x890 -framerate 10 -f x11grab -i :1.0+3840,120 -vf \"crop=iw*0.6:ih:iw*0.2:0\" -y demos/%s.mp4", title))
	if err := recCmd.Start(); err != nil {
		log.Fatal(err)
	}

	recs[title] = recCmd
}

func stopRec(title string) {
	if recs[title] == nil {
		return
	}
	killRec(title)
	exec.Command("/bin/sh", "-c", fmt.Sprintf("ffmpeg -ss 30 -i demos/%s.mp4 -vf \"fps=10:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse\" demos/%s.gif", title, title)).Start()
}

func catchErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestPlaywright(t *testing.T) {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGSTOP, syscall.SIGTTOU, syscall.SIGSTOP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGABRT, syscall.SIGSEGV)

	go func() {
		<-c
		fmt.Println("Program killed!")
		for title := range recs {
			killRec(title)
		}
		os.Exit(1)
	}()

	defer func() {
		if r := recover(); r != nil {
			println("did recover")
			fmt.Println("Program killed!")
			for title := range recs {
				killRec(title)
			}
			os.Exit(1)
		}

		fmt.Println("Program killed!")
		for title := range recs {
			killRec(title)
		}
		os.Exit(1)
	}()

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

	time.Sleep(5 * time.Second)

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

	randId := rand.IntN(1000000)

	personA := &types.IUserProfile{
		Email:     fmt.Sprintf("jsmith%d@myschool.edu", randId),
		FirstName: "John",
		LastName:  "Smith",
	}

	pwA := strings.ToLower(personA.FirstName + personA.LastName)

	if len(cookies) == 0 {
		if _, err = page.Goto(""); err != nil {
			log.Fatalf("could not goto: %v", err)
		}

		startRec("registration")
		// Register user
		page.ByRole("link", "Open Registration!").MouseOver().Click()
		page.ByText("Register").WaitFor()
		doEval()
		catchErr(page.ByLocator("#email").MouseOver().Fill(personA.Email))
		page.ByLocator("#password").MouseOver().Fill(pwA)
		page.ByLocator("#password-confirm").MouseOver().Fill(pwA)
		page.ByLocator("#firstName").MouseOver().Fill(personA.FirstName)
		page.ByLocator("#lastName").MouseOver().Fill(personA.LastName)

		page.ByRole("button", "Register").MouseOver().Click()

		time.Sleep(2 * time.Second)
		stopRec("registration")
	} else {
		if _, err = page.Goto("/app"); err != nil {
			log.Fatalf("could not goto: %v", err)
		}
	}

	time.Sleep(2 * time.Second)

	registrationNext := page.ByRole("button", "I HAVE A GROUP CODE")
	onRegistration, err := registrationNext.IsVisible()
	if err != nil {
		log.Fatalf("registration select err %v", err)
	}

	if onRegistration {
		startRec("onboarding")

		// Fill out group details
		startRec("create_group")

		time.Sleep(3 * time.Second)

		doEval()
		page.ByRole("textbox", "Group Name").MouseOver().Fill(fmt.Sprintf("Downtown Writing Center %d", randId))
		page.ByRole("textbox", "Group Description").MouseOver().Fill("Works with students and the public to teach writing")
		page.ByRole("checkbox", "Use AI Suggestions").MouseOver().SetChecked(true)
		page.ByRole("button", "Next").MouseOver().Click()
		stopRec("create_group")

		// Add group roles
		startRec("create_roles")
		time.Sleep(2 * time.Second)
		page.ByText("Edit Roles").MouseOver().Click()
		page.ByLocator(`span[id^="suggestion-"]`).Nth(0).MouseOver().Click()
		page.ByLocator(`span[id^="suggestion-"]`).Nth(1).MouseOver().Click()
		page.ByRole("combobox", "Default Role").MouseOver().Click()
		page.ByRole("listbox").ByLocator("li").First().Click()
		page.ByRole("button", "Next").MouseOver().Click()
		stopRec("create_roles")

		// Create service
		startRec("create_service")
		page.ByLocator(`span[id^="suggestion-"]`).First().WaitFor() // service name suggestion
		page.ByLocator(`span[id^="suggestion-"]`).First().MouseOver().Click()

		// Select a tier name
		page.ByLocator(`span[id^="suggestion-"]`).Nth(5).WaitFor() // tier name suggestion
		page.ByLocator(`span[id^="suggestion-"]`).Nth(5).MouseOver().Click()

		// Add features to the tier
		page.ByLocator(`span[id^="suggestion-"]`).Nth(10).WaitFor() // feature name suggestion
		page.ByLocator(`span[id^="suggestion-"]`).Nth(10).MouseOver().Click()
		page.ByLocator(`span[id^="suggestion-"]`).Nth(11).MouseOver().Click()
		page.ByRole("button", "Add service tier").MouseOver().Click()

		// Add a second tier
		page.ByLocator(`span[id^="suggestion-"]`).Nth(6).MouseOver().Click() // tier name suggestion
		featuresBox := page.ByRole("combobox", "Features")
		featuresBox.MouseOver().Click()
		featuresList := page.ByRole("listbox", "Features")
		featuresList.ByLocator("li").Nth(1).MouseOver().Click()
		featuresList.ByLocator("li").Nth(2).MouseOver().Click()
		page.Mouse().Click(500, 500)
		page.ByLocator(`span[id^="suggestion-"]`).Nth(12).MouseOver().Click()
		page.ByLocator(`span[id^="suggestion-"]`).Nth(13).MouseOver().Click()
		page.ByRole("button", "Add service tier").MouseOver().Click()

		// Review and save service
		page.Mouse().Wheel(0, 500)
		time.Sleep(2 * time.Second)
		page.ByRole("button", "Next").ScrollIntoViewIfNeeded()
		page.ByRole("button", "Next").MouseOver().Click()
		stopRec("create_service")

		// Create schedule
		startRec("create_schedule")
		page.ByRole("textbox", "Start Date").WaitFor()
		page.ByRole("textbox", "Name").MouseOver().Fill("Fall 2025 Learning Center")
		page.ByRole("textbox", "Start Date").MouseOver().Fill("2025-02-05")
		page.ByRole("button", "Next").MouseOver().Click()
		stopRec("create_schedule")

		// Review group creation
		page.ByText("Review").WaitFor()
		time.Sleep(2 * time.Second)
		page.ByRole("button", "Create group").MouseOver().Click()
		page.ByRole("button", "Click here to confirm.").MouseOver().Click()

		time.Sleep(2 * time.Second)

		stopRec("onboarding")

		// Logout
		page.ByRole("navigation").WaitFor()
		page.ByRole("navigation").MouseOver().Click()
		page.ByText("Logout").MouseOver().Click()
	}

	doEval()

	time.Sleep(1 * time.Second)

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
