package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

const (
	baseURLCoop  = "https://www.coop.se"
	baseURLLogin = "https://login.coop.se"
)

var (
	formActionRe = regexp.MustCompile(`action=['"]([^'"]+)['"]`)
	formInputRe  = regexp.MustCompile(`<input type=['"]hidden['"] name=['"]([^'"]+)['"] value=['"]([^'"]*)['"]`)
)

type Session struct {
	Token          string
	ShoppingUserID string
	Client         *http.Client
}

type spaToken struct {
	Token          string `json:"token"`
	ShoppingUserID string `json:"shoppingUserId"`
	UserID         string `json:"userId"`
	Expires        string `json:"expires"`
	IsBankID       bool   `json:"isBankId"`
	IsPunchout     bool   `json:"isPunchout"`
	AT             string `json:"at"`
}

type loginState struct {
	LoginRequest struct {
		IsValid   bool   `json:"isValid"`
		ClientID  string `json:"clientId"`
		ReturnURL string `json:"returnUrl"`
	} `json:"loginRequest"`
	RedirectURL string `json:"redirectUrl"`
}

// Login performs the full OIDC login flow and returns an authenticated Session.
func Login(email, password string) (*Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	noRedirect := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	followRedirect := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 15 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Step 1: GET /default-login to get the OIDC authorize form.
	resp, err := followRedirect.Get(baseURLCoop + "/default-login")
	if err != nil {
		return nil, fmt.Errorf("fetching default-login: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("reading default-login: %w", err)
	}

	// Step 2: POST to login.coop.se/connect/authorize, which redirects to the login page.
	formAction, formData, err := parseHiddenForm(string(body))
	if err != nil {
		return nil, fmt.Errorf("parsing authorize form: %w", err)
	}

	resp, err = followRedirect.PostForm(formAction, formData)
	if err != nil {
		return nil, fmt.Errorf("posting authorize form: %w", err)
	}
	resp.Body.Close()

	// Extract the ReturnUrl from the login page URL we landed on.
	returnURL := resp.Request.URL.Query().Get("ReturnUrl")
	if returnURL == "" {
		return nil, fmt.Errorf("no ReturnUrl found in redirect URL: %s", resp.Request.URL)
	}

	// Step 3: Get XSRF token.
	xsrfToken, err := getXSRFToken(noRedirect)
	if err != nil {
		return nil, fmt.Errorf("getting XSRF token: %w", err)
	}

	// Step 4: Get login state (passing returnUrl) to get clientId.
	state, err := getLoginState(noRedirect, returnURL, xsrfToken)
	if err != nil {
		return nil, fmt.Errorf("getting login state: %w", err)
	}

	// Step 5: POST credentials.
	err = postLogin(noRedirect, email, password, xsrfToken, state.LoginRequest.ClientID)
	if err != nil {
		return nil, fmt.Errorf("posting login: %w", err)
	}

	// Step 6: Complete OIDC flow by following the authorize callback.
	err = completeOIDCFlow(noRedirect, followRedirect, returnURL)
	if err != nil {
		return nil, fmt.Errorf("completing OIDC flow: %w", err)
	}

	// Step 7: Get the SPA token.
	token, err := getSpaToken(followRedirect)
	if err != nil {
		return nil, fmt.Errorf("getting SPA token: %w", err)
	}

	return &Session{
		Token:          token.Token,
		ShoppingUserID: token.ShoppingUserID,
		Client:         followRedirect,
	}, nil
}

func parseHiddenForm(html string) (string, url.Values, error) {
	actionMatch := formActionRe.FindStringSubmatch(html)
	if actionMatch == nil {
		return "", nil, fmt.Errorf("no form action found")
	}
	action := strings.ReplaceAll(actionMatch[1], "&amp;", "&")

	matches := formInputRe.FindAllStringSubmatch(html, -1)
	values := url.Values{}
	for _, m := range matches {
		values.Set(m[1], m[2])
	}

	return action, values, nil
}

func getXSRFToken(client *http.Client) (string, error) {
	resp, err := client.Get(baseURLLogin + "/local/xsrf")
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	loginURL, _ := url.Parse(baseURLLogin)
	for _, c := range client.Jar.Cookies(loginURL) {
		if c.Name == "XSRF-TOKEN" {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("XSRF-TOKEN cookie not found")
}

func getLoginState(client *http.Client, returnURL, xsrfToken string) (*loginState, error) {
	stateURL := baseURLLogin + "/local/state?returnUrl=" + url.QueryEscape(returnURL)

	req, err := http.NewRequest("GET", stateURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-XSRF-TOKEN", xsrfToken)
	req.Header.Set("RequestVerificationToken", xsrfToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var state loginState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("decoding state: %w", err)
	}

	return &state, nil
}

func postLogin(client *http.Client, email, password, xsrfToken, clientID string) error {
	payload := map[string]interface{}{
		"email":       email,
		"password":    password,
		"accountType": "Private",
		"rememberMe":  true,
		"clientId":    clientID,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURLLogin+"/local/signin/application-schema/email-password", strings.NewReader(string(jsonBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-XSRF-TOKEN", xsrfToken)
	req.Header.Set("RequestVerificationToken", xsrfToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func completeOIDCFlow(noRedirect, followRedirect *http.Client, returnURL string) error {
	if returnURL == "" || returnURL == "/" {
		return fmt.Errorf("no valid returnUrl for OIDC callback")
	}

	fullURL := returnURL
	if !strings.HasPrefix(returnURL, "http") {
		fullURL = baseURLLogin + returnURL
	}

	resp, err := noRedirect.Get(fullURL)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	// Handle form_post response directly.
	if resp.StatusCode == http.StatusOK {
		return submitSigninOIDCForm(followRedirect, string(body))
	}

	// Handle redirect, then check for form_post.
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if location == "" {
			return fmt.Errorf("redirect with no Location header (status %d)", resp.StatusCode)
		}
		if !strings.HasPrefix(location, "http") {
			location = baseURLLogin + location
		}

		resp, err = noRedirect.Get(location)
		if err != nil {
			return err
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		return submitSigninOIDCForm(followRedirect, string(body))
	}

	return fmt.Errorf("unexpected response from authorize callback (status %d)", resp.StatusCode)
}

func submitSigninOIDCForm(client *http.Client, html string) error {
	action, formData, err := parseHiddenForm(html)
	if err != nil {
		return fmt.Errorf("parsing signin-oidc form: %w", err)
	}
	resp, err := client.PostForm(action, formData)
	if err != nil {
		return fmt.Errorf("posting signin-oidc: %w", err)
	}
	resp.Body.Close()
	return nil
}

func getSpaToken(client *http.Client) (*spaToken, error) {
	resp, err := client.Get(baseURLCoop + "/api/spa/token?_=0")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get SPA token (status %d): %s", resp.StatusCode, string(body))
	}

	var token spaToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decoding SPA token: %w", err)
	}

	if token.Token == "" {
		return nil, fmt.Errorf("empty token in SPA response")
	}

	return &token, nil
}
