<h1 align="center">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/ddflare/ddflare/refs/heads/main/assets/logo/ddflare-logotype.svg#gh-light-mode-only" width=130 alt="logo">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/ddflare/ddflare/refs/heads/main/assets/logo/ddflare-logotype-dark.svg#gh-dark-mode-only" width=130 alt="logo">
  <br>

  [![release workflow](https://github.com/ddflare/ddflare/actions/workflows/build.yaml/badge.svg)](https://github.com/ddflare/ddflare/actions/workflows/build.yaml)
  [![container workflow](https://github.com/ddflare/ddflare/actions/workflows/container-image.yaml/badge.svg)](https://github.com/ddflare/ddflare/actions/workflows/container-image.yaml)
</h1>

ddflare is a [DDNS (Dynamic DNS)](https://en.wikipedia.org/wiki/Dynamic_DNS) library that allows DNS record
updates via either the [Cloudflare API](https://developers.cloudflare.com/api/) or
the [DynDNS update prococol v3](https://help.dyn.com/remote-access-api/perform-update/).
<br>
It comes with a CLI tool built on top of the library and released for different architectures.

>[!NOTE]
>The [DynDNS update prococol v3](https://help.dyn.com/remote-access-api/) is an HTTP API introduced
>back in the day by dyndns.org (now [Dyn](https://account.dyn.com/), part of Oracle) which is used by
>many of the available DDNS service providers available nowadays.

>[!TIP]
>In order to update your DNS records with ddflare you need a DNS domain registered with Cloudflare
>or a registratrion to a Dynamic DNS service provider
>(e.g., [No-IP](https://www.noip.com), [Dyn](https://account.dyn.com/), ...).

ddflare allows to:
* update a target domain name (FQDN, recorded as a type A record) to point to the current public address
or a custom IP
* retrieve and display the current public IP address
* resolve any domain name (acting as a simple DNS client)

<br>

Quickstart CLI
====
Get ddflare
----
ddflare CLI is released as statically compiled binaries for different OS/architetures that you can grab
from the [release page](https://github.com/ddflare/ddflare/releases/latest).

Get a x86_64 linux binary example:
```bash
wget https://github.com/ddflare/ddflare/releases/download/v0.6.0/ddflare-linux-amd64
sudo install ddflare-linux-amd64 /usr/local/bin/ddflare
```

Container images are availble as well on the github repository. Run ddflare
via docker with:

```bash
docker run -ti --rm ghcr.io/ddflare/ddflare:0.6.0
```

Available commands
----
ddflare has two main commands:
* `set` - to update the type A record of the target FQDN
to the current public ip (or a custom IP address)
* `get` - to retrieve the current public IP address
(or resolve the FQDN passed as argument)

Run `ddflare help` to display all the available commands and options.

Update domain name via DDNS services
----
>[!NOTE]
>You should register to a DDNS service first, configure the desired FQDN there and retrieve the
>_username_ and _password_ required to authenticate to the service.

To update a domain name (FQDN, type A record) to the current public IP address of the host run:

```bash
ddflare set -s <DDNS_URL> -t <AUTH_TOKEN> <FQDN>
```
where `<DDNS_URL>` is the http endpoint of the DDNS service,
`<FQDN>` is the domain to be updated and `<AUTH_TOKEN>` is the API authentication token.

**Example:**
|||
|-|-|
|DDNS provider| [No-IP](https://www.noip.com/)|
|FQDN to update| `myhost.ddns.net`|
|user |`68816xj`|
|password| `v4UMHzugodpE`|
```bash
ddflare set -s noip -t 68816xj:v4UMHzugodpE  myhost.ddns.net
```

Update domain name via [Cloudflare](https://www.cloudflare.com/)
----
To update a domain name (FQDN, type A record) to the current public IP address of the host run:

```bash
ddflare set -t <CLOUDFLARE_TOKEN> <FQDN>
```
where `<FQDN>` is the domain to be updated and `<CLOUDFLARE_TOKEN>` is the Cloudflare API token.

>[!TIP]
>You should get a Cloudflare API token with `Zone.DNS` `Edit` permission.
>You can create one following the
>[Cloudflare docs](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/).

**Example:**
|||
|-|-|
|FQDN to update| `myhost.example.com`|
|API token|`gB1fOLbl2d5XynhfIOJvzX8Y4rZnU5RLPW1hg7cM`|
```bash
ddflare set -t gB1fOLbl2d5XynhfIOJvzX8Y4rZnU5RLPW1hg7cM  myhost.example.com
```

>[!NOTE]
>ddflare can <ins>update</ins> existing type A DNS records but cannot create new ones yet, so the
>record should be created in advance.
>
>To create a type A record see Cloudflare's
>[Manage DNS Records](https://developers.cloudflare.com/dns/manage-dns-records/how-to/create-dns-records/)
>docs.

>[!TIP]
>When creating a type A DNS record pay attention to the value of the `TTL` field:
>it tracks the number of seconds DNS clients and DNS resolvers are allowed to
>cache the resolved IP address.
>You may want to keep the TTL low (the allowed minimum is 60 secs) if you plan to use the record
>to track the (dynamic) IP address of a host in a DDNS scenario.

Get the current public IP address
----
Retrieving the current public IP address is as easy as running:
```bash
ddflare get
```
ddflare queries the `ipify.org` service under the hood, which detects the public IP address used to reach the service.

Quickstart Library
====
The ddflare go library allows updates of DNS type A records with the actual Public IP address in 4 steps:
1. create a new _DNSManager_  (`NewDNSManager()`)
2. initialize the DNSManager with the authentication credentials (`.Init()`)
3. retrieve the current public IP address (`GetPublicIP()`)
4. update the DNS record via the DNSManager (`.UpdateFQDN`)


**Example**:
```go
import "github.com/ddflare/ddflare"

func main() {
  // fqdn to be updated
  fqdn := "www.myddns.host"
  // auth Token from the DDNS service if available or contatenation of user:password
  authToken := "user:password"

  // create a new DNSManager targeting the desired service (ddflare.Cloudflare,
  // ddflare.NoIP or ddflare.DDNS). For a DDNS provider using the DynDNS API v3
  // but not in the list, pick ddflare.DDNS and set a custom API endpoint with
  // dm.SetAPIEndpoint("$HTTP_ENDPOINT")
  dm, err := ddflare.NewDNSManager(ddflare.NoIP)
  if err != nil {
    log.Fatal(err)
  }

  // init the DNSManager with the API credentials
  if err = dm.Init(authToken); err != nil {
    log.Fatal(err)
  }

  // retrieve the current Public IP address
  pubIP, err := ddflare.GetPublicIP()
  if err != nil {
    log.Fatal(err)
  }
  // set the Public IP address just retrieved as the addres of `fqdn`
  if err = dm.UpdateFQDN(fqdn, pubIP); err != nil {
    log.Fatal("update failed")
  }

  log.Info("update successful")
}
```
