Azure Active Directory OAuther
------------------------------

Simplistic command line tool written in go that implements an Azure Active Directory [authorization code grant flow](https://msdn.microsoft.com/en-us/library/azure/dn645542.aspx)
to retrieve an OAuth token. The token can be used by an Azure Active Directory Web Application (or Web API)
to make requests to the Azure Service Management APIs on behalf of the user granting access. 

Usage (assuming the [go](http://golang.org) toolchain in properly setup):
```
export CLIENT_ID=<your application client id here>
export CLIENT_SECRET=<your application client secret here>
make run
```

A couple of things to note:

* The tool is designed to be run on Mac OS X. This is because it needs to open web pages and relies on the `open`
  shell command to do so (PRs welcome for other OSes ;) ). It will "work" on other OSes but the experience won't
  be great (read requires to copy paste into HTML file then open in browser).

* The application must be a multi-tenant application, refer to [Adding, Updating, and Removing an Application](https://msdn.microsoft.com/en-us/library/azure/dn132599.aspx)
  for more information.

