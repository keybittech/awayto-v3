package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/playwright-community/playwright-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	pages      = make(map[string]*Page)
	responses  = make([]*PageResponse, 0)
	responseMu sync.Mutex
)

type PageResponse struct {
	body        []byte
	method, url string
}

type UserWithPass struct {
	Password            string
	UserId              string
	AuthorizationHeader map[string]string
	Profile             *types.IUserProfile
}

type Page struct {
	playwright.Page
	*testing.T
	*UserWithPass
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

func (p *Page) Close(t *testing.T) {
	err := p.Page.Close()
	if err != nil {
		t.Fatal(err)
	}
}

type Locator struct {
	playwright.Locator
	*testing.T
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

func (l Locator) WaitFor(options ...playwright.LocatorWaitForOptions) error {
	return l.Locator.WaitFor(options...)
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

func debugErr(err error) error {
	return errors.New(fmt.Sprintf("%s %s", err, debug.Stack()))
}

// return a page and if it already existed
func getBrowserPage(t *testing.T, userId string) (*Page, bool) {
	if p, ok := pages[userId]; ok {
		return p, true
	}

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

	p.OnResponse(func(response playwright.Response) {
		go func() {
			body, err := response.Body()
			if err != nil || body == nil || strings.Contains(response.URL(), "src") || strings.Contains(response.URL(), "node_modules") {
				return
			}
			responseMu.Lock()
			responses = append(responses, &PageResponse{
				body:   body,
				method: response.Request().Method(),
				url:    response.URL(),
			})
			responseMu.Unlock()
		}()
	})

	if useRandUser {
		userId += strconv.Itoa(rand.IntN(1000000) + 1)
	}

	user := &UserWithPass{
		UserId: userId,
		Profile: &types.IUserProfile{
			Email:     fmt.Sprintf("jsmith%s@myschool.edu", userId),
			FirstName: "John",
			LastName:  "Smith",
		},
	}

	user.Password = user.Profile.FirstName + user.Profile.LastName

	println("Testing with user:", user.Profile.GetEmail(), "pass:", user.Password)

	page := &Page{p, t, user}

	doEval(page)

	pages[userId] = page
	return pages[userId], false
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

func clearResponses() {
	responseMu.Lock()
	responses = make([]*PageResponse, 0)
	responseMu.Unlock()
}

func readResponse[T proto.Message](method, path string, action func()) (T, error) {
	var empty T
	clearResponses()

	action()

	time.Sleep(1 * time.Second)

	defer clearResponses()

	responseMu.Lock()
	defer responseMu.Unlock()

	checked := make([]string, 0)
	for _, r := range responses {
		if method != r.method {
			checked = append(checked, "\n"+r.method+" "+r.url)
			continue
		}

		matched, err := regexp.MatchString(path, r.url)
		if err != nil {
			return empty, fmt.Errorf("failed to match path %s to url %s", path, r.url)
		}

		if matched {
			pb := reflect.New(reflect.TypeOf(empty).Elem()).Interface().(T)
			err := protojson.Unmarshal(r.body, pb)
			if err != nil {
				return empty, fmt.Errorf("failed to unmarshal during %s parsing, %v", path, err)
			}
			return pb, nil
		}
	}

	return empty, fmt.Errorf("did not find any responses for path %s, checked %v", path, checked)
}

func readHandlerResponse[T proto.Message](action func()) (T, error) {
	var empty T
	emptyT := reflect.TypeOf(empty)
	emptyTName := emptyT.Elem().Name()
	var msg proto.Message
	var matched, ok bool
	var err error

	// T must be a pointer to a proto type
	msg, ok = reflect.Zero(emptyT).Interface().(proto.Message)
	if !ok {
		return empty, nil
	}

	var hops *util.HandlerOptions

	descriptorT := msg.ProtoReflect().Descriptor()
	for _, h := range handlerOpts {
		if descriptorT == h.ServiceMethod.Output() {
			hops = h
			break
		}
	}
	if hops == nil {
		return empty, fmt.Errorf("type did not resolve to a service output message %s", emptyTName)
	}

	clearResponses()

	action()

	time.Sleep(1 * time.Second)

	defer clearResponses()

	responseMu.Lock()
	defer responseMu.Unlock()

	checked := make([]string, 0)
	for _, r := range responses {
		// verify the right method
		if !strings.HasPrefix(hops.Pattern, r.method) {
			checked = append(checked, r.method+" "+r.url)
			continue
		}

		// verify a matching invalidation path
		for _, invalidation := range hops.Invalidations {
			matched, err = regexp.MatchString(invalidation, r.url)
			if err != nil {
				return empty, fmt.Errorf("failed to match invalidation %s to url %s", invalidation, r.url)
			}
			if matched {
				pb := reflect.New(reflect.TypeOf(empty).Elem()).Interface().(T)

				err = protojson.Unmarshal(r.body, pb)
				if err != nil {
					return empty, fmt.Errorf("failed to unmarshal into message: %s, data: %s, %v", emptyTName, r.body, err)
				}

				return pb, nil
			}
		}
		checked = append(checked, r.method+" "+r.url)
	}

	return empty, fmt.Errorf("did not find any responses for output message %s (%s), checked %s", emptyTName, hops.Pattern, checked)
}

func login(t *testing.T, userId string) *Page {
	page, existing := getBrowserPage(t, userId)
	if !strings.HasSuffix(strings.Trim(page.URL(), "/"), "app") {
		page.Goto("/app")
		doEval(page)
	}

	_, err := page.Page.Evaluate("() => window.localStorage.clear()")
	if err != nil {
		t.Fatalf("error cleaning local storage on delete login %v", err)
	}

	onSignInPage, err := page.GetByText("Sign in to your account").IsVisible()
	if err != nil {
		t.Fatalf("sign in error %v", err)
	}

	if onSignInPage {
		page.ByRole("textbox", "Email").MouseOver().Fill(page.UserWithPass.Profile.Email)
		page.ByRole("textbox", "Password").MouseOver().Fill(page.UserWithPass.Password)
		tokenResponse, err := readResponse[*types.OIDCToken](http.MethodPost, "/auth/realms/.*/protocol/openid-connect/token", func() {
			page.ByRole("button", "Sign In").MouseOver().Click()
		})
		if err == nil && tokenResponse != nil && tokenResponse.GetAccessToken() != "" {
			authHeader := make(map[string]string)
			authHeader["Authorization"] = "Bearer " + tokenResponse.GetAccessToken()
			page.UserWithPass.AuthorizationHeader = authHeader
		} else if existing {
			t.Fatalf("did not refresh auth token with a setup user, err: %v", err)
		}
	}

	return page
}

func register(page *Page, codeParam ...string) {
	page.ByRole("link", "Register").MouseOver().Click()

	// On registration page
	doEval(page)
	if len(codeParam) > 0 {
		page.ById("groupCode").MouseOver().Fill(codeParam[0])
	}
	page.ById("email").MouseOver().Fill(page.UserWithPass.Profile.Email)
	page.ById("password").MouseOver().Fill(page.UserWithPass.Password)
	page.ById("password-confirm").MouseOver().Fill(page.UserWithPass.Password)
	page.ById("firstName").MouseOver().Fill(page.UserWithPass.Profile.FirstName)
	page.ById("lastName").MouseOver().Fill(page.UserWithPass.Profile.LastName)
	page.ByRole("button", "Register").MouseOver().Click()

	println(fmt.Sprintf("Registered user %s with pass %s", page.UserWithPass.Profile.Email, page.UserWithPass.Password))
}

func goHome(page *Page) {
	page.ById("topbar_open_menu").MouseOver().Click()
	page.ByText("Home").MouseOver().Click()
}

func getFirstVisible(page *Page, ids []string, duration ...time.Duration) (string, error) {
	var d time.Duration = 6
	if len(duration) > 0 {
		d = duration[0]
	}
	ctx, cancel := context.WithTimeout(context.Background(), d*time.Second)
	defer cancel()

	resultChan := make(chan string, 1)
	defer close(resultChan)

	for _, id := range ids {
		go func() {
			err := page.ById(id).WaitFor()
			if err == nil {
				resultChan <- id
			}
		}()
	}

	select {
	case res := <-resultChan:
		return res, nil
	case <-ctx.Done():
		// No result
	}

	return "", fmt.Errorf("didn't find any of %s after %d seconds", ids, d)
}
