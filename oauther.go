// Example usage:
//   ./oauther --client=$CLIENT_ID --secret=$CLIENT_SECRET --redirect=https://ad.test.rightscale.com
//
// TBD: handle errors
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/alecthomas/kingpin"
)

const (
	authHost      = "https://login.windows.net"
	codeEndpoint  = "/oauth2/authorize"
	tokenEndpoint = "/oauth2/token"
)

var (
	dump         = kingpin.Flag("dump", "Dump request/response").Bool()
	clientId     = kingpin.Flag("client", "The client id of the application that is registered in Azure Active Directory.").Required().String()
	clientSecret = kingpin.Flag("secret", "The client key of the application that is registered in Azure Active Directory.").Required().String()
	tenantId     = kingpin.Flag("tenant", "Azure AD application tenant.").Required().String()
	multiTenant  = kingpin.Flag("multi", "Whether application is multi-tenant.").Default("false").Bool()
	domainHint   = kingpin.Flag("domain", "Provides a hint about the tenant or domain that the user should use to sign in.").String()
	loginHint    = kingpin.Flag("hint", "Provides a hint to the user on the sign-in page.").String()
	prompt       = kingpin.Flag("prompt", "Indicate the type of user interaction that is required.").Enum("login", "consent", "admin_consent")
	redirectUri  = kingpin.Flag("redirect", "Specifies the reply URL of the application.").Required().String()
	resource     = kingpin.Flag("resource", "The App ID URI of the web API (secured resource).").Default("https://management.core.windows.net/").String()
	state        = kingpin.Flag("state", "A randomly generated non-reused value that is sent in the request and returned in the response. This parameter is used as a mitigation against cross-site request forgery (CSRF) attacks.").String()

	// Signin URL returned by auth code request
	signinUrl string
)

func main() {
	kingpin.Parse()
	req, err := buildAuthCodeRequest()
	if err != nil {
		kingpin.Fatalf("failed to build auth code request: %s", err)
	}
	signinUrl = req.URL.String()
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open")
		cmd.Args = append(cmd.Args, signinUrl)
		fmt.Printf(">%s\nPlease paste the updated (redirected) URL below from the opened browser window here:\n", signinUrl)
		kingpin.FatalIfError(cmd.Run(), "open browser window")
	} else {
		fmt.Printf(">%s\nPlease go to the URL above in a browser and paste the updated (redirected) URL below:\n", signinUrl)
	}
	bio := bufio.NewReader(os.Stdin)
	loc, _, err := bio.ReadLine()
	if err != nil {
		kingpin.Fatalf("failed to read redirect: %s", err)
	}
	u, err := url.Parse(string(loc))
	if err != nil {
		kingpin.Fatalf("failed to parse pasted URL: %s", err)
	}
	code := u.Query().Get("code")
	if code == "" {
		kingpin.Fatalf("failed to extract code from location header in auth code response, header was '%s'", loc)
	}
	req, err = buildAccessTokenRequest(code)
	if err != nil {
		kingpin.Fatalf("failed to build code redeem request: %s", err)
	}
	resp, err := performRequest(req, nil)
	if err != nil {
		kingpin.Fatalf("failed to make code redeem request: %s", err)
	}
	if runtime.GOOS == "darwin" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			kingpin.Fatalf("failed to load response body: %s", err)
		}
		go serveHtml(string(body), 8088)
		cmd := exec.Command("open")
		cmd.Args = append(cmd.Args, "http://localhost:8088/index.html")
		kingpin.FatalIfError(cmd.Run(), "open browser window")
		fmt.Println("Press enter when done")
		loc, _, err = bio.ReadLine()
	} else {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			kingpin.Fatalf("failed to serialize code redeem response: %s", err)
		}
		fmt.Println(string(b))
	}
}

// Server html on localhost on given port, super hyper simple - not reentrant - doesn't cleanup etc. (call once)
func serveHtml(html string, port int) {
	http.HandleFunc("/index.html", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, html) })
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		panic(err) // Crash the whole thing
	}
}

// Build request to retrieve authorization code
func buildAuthCodeRequest() (*http.Request, error) {
	u, _ := url.Parse(endpoint(false, codeEndpoint))
	values := u.Query()
	values.Set("response_type", "code")
	values.Set("client_id", *clientId)
	values.Set("resource", *resource)
	setOptionalParam(values, "domain_hint", *domainHint)
	setOptionalParam(values, "login_hint", *loginHint)
	setOptionalParam(values, "prompt", *prompt)
	setOptionalParam(values, "redirect_uri", *redirectUri)
	setOptionalParam(values, "state", *state)
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Build request to redeem authorization code and get access token
func buildAccessTokenRequest(code string) (*http.Request, error) {
	*tenantId = ""
	u, _ := url.Parse(endpoint(*multiTenant, tokenEndpoint))
	data := url.Values{}
	data.Set("client_id", *clientId)
	data.Set("client_secret", *clientSecret)
	data.Set("code", code)
	data.Set("resource", *resource)
	data.Set("grant_type", "authorization_code")
	setOptionalParam(data, "redirect_uri", *redirectUri)
	encoded := data.Encode()
	req, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(encoded))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(encoded)))
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Compute endpoint taking into account tenantId if any.
// No tenantId means the application is multi-tenant and uses the "common" namespace to auth.
func endpoint(useCommon bool, suffix string) string {
	prefix := authHost
	if useCommon {
		prefix += "/common"
	} else {
		prefix += "/" + *tenantId
	}
	return prefix + suffix
}

// Dump request content if required then make request and dump response if needed.
func performRequest(req *http.Request, redirectHandler RedirectHandler) (*http.Response, error) {
	req.URL.Scheme = "https"
	req.Header.Set("User-Agent", "RightScale")
	if *dump {
		b, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", string(b))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to dump request content - %s\n", err)
		}
	}
	c := &http.Client{CheckRedirect: redirectHandler}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if *dump {
		b, err := httputil.DumpResponse(resp, false)
		if err == nil {
			fmt.Fprintf(os.Stderr, "--------\n%s", b)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to dump response content - %s\n", err)
		}
	}

	return resp, err
}

// Function that can be given to the "CheckRedirect" field of http.Client
type RedirectHandler func(req *http.Request, via []*http.Request) error

// Update signinUrl global
// Return nil so redirect is followed in case of intermediary redirects before final signin page
func captureSigninUrl(req *http.Request, via []*http.Request) error {
	signinUrl = req.URL.String()
	return nil
}

// Helper function that sets optional query string param
func setOptionalParam(values url.Values, name, val string) {
	if val != "" {
		values.Set(name, val)
	}
}
