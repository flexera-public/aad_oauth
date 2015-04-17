Azure Active Directory OAuther
------------------------------

Simplistic command line tool written in go that implements an Azure Active Directory [authorization code grant flow](https://msdn.microsoft.com/en-us/library/azure/dn645542.aspx)
to retrieve an OAuth token. The token can be used by an Azure Active Directory Web Application (or Web API)
to make requests to the Azure Service Management APIs on behalf of the user granting access for example. 

Usage (assuming the [go](http://golang.org) toolchain is properly setup):
```
export CLIENT_ID=<your application client id here>
export CLIENT_SECRET=<your application client secret here>
make run
```
For additional control:
```
$ ./oauther --help
usage: oauther --client=CLIENT --secret=SECRET --tenant=TENANT --redirect=REDIRECT [<flags>]

Flags:
  --help               Show help.
  --dump               Dump request/response
  --client=CLIENT      The client id of the application that is registered in Azure Active Directory.
  --secret=SECRET      The client key of the application that is registered in Azure Active Directory.
  --tenant=TENANT      Azure AD application tenant.
  --multi              Whether application is multi-tenant.
  --domain=DOMAIN      Provides a hint about the tenant or domain that the user should use to sign in.
  --hint=HINT          Provides a hint to the user on the sign-in page.
  --prompt=PROMPT      Indicate the type of user interaction that is required.
  --redirect=REDIRECT  Specifies the reply URL of the application.
  --resource="https://management.core.windows.net/"
                       The App ID URI of the web API (secured resource).
  --state=STATE        A randomly generated non-reused value that is sent in the request and returned in the response.
                       This parameter is used as a mitigation against cross-site request forgery (CSRF) attacks.
```
*Note:*
The tool is designed to be run on Mac OS X. This is because it needs to open web pages and relies on the `open`
shell command to do so (PRs welcome for other OSes ;) ). It will "work" on other OSes but the experience won't
be great (read requires to copy paste into HTML file then open in browser).

