package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/playwright-community/playwright-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Locator struct {
	playwright.Locator
	*testing.T
}

func debugErr(err error) error {
	return errors.New(fmt.Sprintf("%s %s", err, debug.Stack()))
}

// Actionable
func (l Locator) MouseOver() Locator {
	err := l.Hover()
	if err != nil {
		l.T.Fatal(debugErr(err))
	}
	return l
}

func (l Locator) Click() {
	if err := l.Locator.Click(); err != nil {
		l.T.Fatal(debugErr(err))
	}
}

func (l Locator) WaitFor() {
	if err := l.Locator.WaitFor(); err != nil {
		l.T.Fatal(debugErr(err))
	}
}

func (l Locator) Fill(value string, opts ...playwright.LocatorFillOptions) {
	if err := l.Locator.Fill(value, opts...); err != nil {
		l.T.Fatal(debugErr(err))
	}
}

func (l Locator) SetChecked(checked bool, opts ...playwright.LocatorSetCheckedOptions) {
	if err := l.Locator.SetChecked(checked, opts...); err != nil {
		l.T.Fatal(debugErr(err))
	}
}

func (l Locator) IsVisible() bool {
	visible, err := l.Locator.IsVisible()
	if err != nil {
		l.T.Fatal(debugErr(err))
	}
	return visible
}

// Selectors
func (l Locator) ByRole(role string, name ...string) Locator {
	opts := playwright.LocatorGetByRoleOptions{}
	if len(name) > 0 {
		opts.Name = name[0]
	}
	return Locator{l.GetByRole(playwright.AriaRole(role), opts), l.T}
}

func (l Locator) Nth(num int) Locator {
	return Locator{l.Locator.Nth(num), l.T}
}

func (l Locator) First() Locator {
	return Locator{l.Locator.First(), l.T}
}

func (l Locator) Last() Locator {
	return Locator{l.Locator.Last(), l.T}
}

func (l Locator) ByLocator(selector string) Locator {
	return Locator{l.Locator.Locator(selector), l.T}
}

func (l Locator) ByText(text string) Locator {
	return Locator{l.Locator.GetByText(text), l.T}
}

type Page struct {
	playwright.Page
	*testing.T
}

func (p *Page) ByRole(role string, name ...string) Locator {
	opts := playwright.PageGetByRoleOptions{}
	if len(name) > 0 {
		opts.Name = name[0]
	}
	return Locator{p.GetByRole(playwright.AriaRole(role), opts), p.T}
}

func (p *Page) ByLabel(label string) Locator {
	return Locator{p.GetByLabel(label), p.T}
}

func (p *Page) ByText(text string) Locator {
	return Locator{p.GetByText(text), p.T}
}

func (p *Page) ByLocator(selector string) Locator {
	return Locator{p.Locator(selector), p.T}
}

func (p *Page) ByTestId(selector string) Locator {
	return Locator{p.GetByTestId(selector), p.T}
}

func (p *Page) ById(selector string) Locator {
	return Locator{p.Locator(`[id="` + selector + `"]`), p.T}
}

func (p *Page) ByIdStartsWith(selector string) Locator {
	return Locator{p.Locator(`[id^="` + selector + `"]`), p.T}
}

func getBrowserPage(t *testing.T) *Page {
	defaultOptions := playwright.BrowserNewPageOptions{
		BaseURL:           playwright.String(fmt.Sprintf("https://localhost:%d", util.E_GO_HTTPS_PORT)),
		IgnoreHttpsErrors: playwright.Bool(true),
		ColorScheme:       playwright.ColorSchemeDark,
		NoViewport:        playwright.Bool(true),
		RecordVideo: &playwright.RecordVideo{
			Dir:  filepath.Join(util.E_PROJECT_DIR, "demos/"),
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

	page := &Page{p, t}

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
	Password string
	Profile  *types.IUserProfile
	UserId   int
}

var testUserWithPass *UserWithPass

func getUiUser() *UserWithPass {
	defer func() {
		println("Testing with user:", testUserWithPass.Profile.GetEmail(), "pass:", testUserWithPass.Password)
	}()

	if testUserWithPass != nil {
		return testUserWithPass
	}

	var userId int
	if useRandUser {
		userId = rand.IntN(1000000) + 1
	} else {
		userId = 1
	}

	testUserWithPass = &UserWithPass{
		UserId: userId,
		Profile: &types.IUserProfile{
			Email:     fmt.Sprintf("jsmith%d@myschool.edu", userId),
			FirstName: "John",
			LastName:  "Smith",
		},
	}

	testUserWithPass.Password = strings.ToLower(testUserWithPass.Profile.FirstName + testUserWithPass.Profile.LastName)

	return testUserWithPass
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

type ResponseError struct {
	URL        string
	StatusCode int
	Body       string
	Err        error
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("response error from %s (status: %d): %v\nBody: %s",
		e.URL, e.StatusCode, e.Err, e.Body)
}

// ExpectResponseWrapper creates a function that, when called, will wait for a response matching
// the specified URL substring after performing an action (like clicking a button).
// It returns the response data unmarshaled into the specified protobuf message type T.
//
// Usage:
//
//	getResponse := ExpectResponseWrapper[*myproto.ResponseType](page, "api/endpoint")
//
//	// Call the function, passing a callback that performs the action
//	response, err := getResponse(func() error {
//	    // Click button or perform action that triggers the request
//	    return page.GetByRole("button", playwright.PageGetByRoleOptions{
//	        Name: playwright.String("Submit"),
//	    }).Click()
//	})
//
//	if err != nil {
//	    // Handle error
//	}
//	// Use the typed response
//	fmt.Println(response.SomeField)
func expectResponseWrapper[T proto.Message](page playwright.Page, urlSubstring string) func(action func() error) (T, error) {
	var zero T

	// Return a function that executes the action and waits for the response
	return func(action func() error) (T, error) {
		// Set up and wait for the response
		response, err := page.ExpectResponse(
			// URL matcher function
			func(r playwright.Response) bool {
				return contains(r.URL(), urlSubstring)
			},
			// Action to trigger the request
			action,
		)

		if err != nil {
			return zero, fmt.Errorf("failed to receive response: %w", err)
		}

		// Get the response URL for error reporting
		responseURL := response.URL()

		// Check the status code
		statusCode := response.Status()
		if statusCode < 200 || statusCode >= 300 {
			// Get body for error details
			body, bodyErr := response.Text()
			if bodyErr != nil {
				body = "[failed to extract response body]"
			}
			return zero, ResponseError{
				URL:        responseURL,
				StatusCode: statusCode,
				Body:       body,
				Err:        errors.New("non-success status code"),
			}
		}

		// Get the body
		body, err := response.Text()
		if err != nil {
			return zero, fmt.Errorf("failed to get response body: %w", err)
		}

		// Try to unmarshal as protobuf
		result := proto.Clone(zero).(T)
		err = protojson.Unmarshal([]byte(body), result)
		if err != nil {
			// For better error messages, check if the body is valid JSON
			var jsonCheck any
			jsonErr := json.Unmarshal([]byte(body), &jsonCheck)
			if jsonErr != nil {
				return zero, fmt.Errorf("response is not valid JSON: %w\nBody: %s", jsonErr, body)
			}
			return zero, fmt.Errorf("failed to unmarshal response as proto: %w", err)
		}

		return result, nil
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && substr != "" && (s == substr ||
		len(s) >= len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			(len(s) > len(substr)*2 && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)-len(substr)/2] == substr)))
}
