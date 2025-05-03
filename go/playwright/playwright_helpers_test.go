package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
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

func (p *Page) ByTestId(selector string) Locator {
	return Locator{p.GetByTestId(selector)}
}

func (p *Page) ById(selector string) Locator {
	return Locator{p.Locator(`[id="` + selector + `"]`)}
}

func (p *Page) ByIdStartsWith(selector string) Locator {
	return Locator{p.Locator(`[id^="` + selector + `"]`)}
}

func getBrowserPage(t *testing.T) *Page {
	defaultOptions := playwright.BrowserNewPageOptions{
		BaseURL:           playwright.String("https://localhost:" + os.Getenv("GO_HTTPS_PORT")),
		IgnoreHttpsErrors: playwright.Bool(true),
		ColorScheme:       playwright.ColorSchemeDark,
		NoViewport:        playwright.Bool(true),
		RecordVideo: &playwright.RecordVideo{
			Dir:  filepath.Join(os.Getenv("PROJECT_DIR"), "demos/"),
			Size: &playwright.Size{Width: 1556, Height: 900},
		},
	}

	p, err := browser.NewPage(defaultOptions)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := p.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	page := &Page{p}

	doEval(page)

	return page
}

func doEval(page *Page) {
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

type UserWithPass struct {
	UserId   int
	Password string
	Profile  *types.IUserProfile
}

func getUiUser() *UserWithPass {
	var userId int
	if useRand {
		userId = rand.IntN(1000000) + 1
	} else {
		userId = 1
	}

	user := &UserWithPass{
		UserId: userId,
		Profile: &types.IUserProfile{
			Email:     fmt.Sprintf("jsmith%d@myschool.edu", userId),
			FirstName: "John",
			LastName:  "Smith",
		},
	}

	user.Password = strings.ToLower(user.Profile.FirstName + user.Profile.LastName)

	return user
}

func login(t *testing.T, page *Page, user *UserWithPass) {
	if !strings.HasSuffix(page.URL(), "app") {
		page.Goto("/app")
		doEval(page)
	}

	onSignInPage, err := page.GetByText("Sign in to your account").IsVisible()
	if err != nil {
		t.Fatalf("sign in error %v", err)
	}

	if onSignInPage {
		page.ByRole("textbox", "Email").MouseOver().Fill(user.Profile.Email)
		page.ByRole("textbox", "Password").MouseOver().Fill(user.Password)
		page.ByRole("button", "Sign In").MouseOver().Click()
		time.Sleep(time.Second)
		doEval(page)
	}
}

func goHome(page *Page) {
	page.ById("topbar_open_menu").MouseOver().Click()
	page.ByText("Home").MouseOver().Click()
}
