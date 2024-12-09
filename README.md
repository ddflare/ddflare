<h1 align="center">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/fgiudici/ddflare/refs/heads/main/assets/logo/ddflare-logotype.svg#gh-light-mode-only" width=130 alt="logo">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/fgiudici/ddflare/refs/heads/main/assets/logo/ddflare-logotype-dark.svg#gh-dark-mode-only" width=130 alt="logo">
  <br>

  [![example workflow](https://github.com/fgiudici/ddflare/actions/workflows/build.yaml/badge.svg)](https://github.com/fgiudici/ddflare/actions/workflows/build.yaml)
  [![example workflow](https://github.com/fgiudici/ddflare/actions/workflows/container-image.yaml/badge.svg)](https://github.com/fgiudici/ddflare/actions/workflows/container-image.yaml)
</h1>

ddflare provides a [DDNS (Dynamic DNS)](https://en.wikipedia.org/wiki/Dynamic_DNS) client that allows DNS type A record updates via the [DynDNS update prococol v3](https://help.dyn.com/remote-access-api/perform-update/) or via the [Cloudflare API](https://developers.cloudflare.com/api/).


>[!NOTE]
>ddflare acts as a client for Dynamic DNS services. You can used ddflare with yor Cloudflare registered
>DNS domain via the [Cloudflare API](https://developers.cloudflare.com/api/) or register to a Dynamic DNS
>service provider (e.g., [No-IP](https://www.noip.com), [Dyn](https://account.dyn.com/), ...) and update
>your DNS name via the [DynDNS update prococol v3](https://help.dyn.com/remote-access-api/).
>
>The [DynDNS update prococol v3](https://help.dyn.com/remote-access-api/) is a HTTP based API introduced
>back in the day by dyndns.org (now [Dyn](https://account.dyn.com/), part of Oracle) which is used by
>many of the available DDNS service providers.

While ddflare main functionality is to run in a loop to update the target FQDN to the current public IP
when a change is detected, it allows to:
* retrieve and display the current public IP address
* update a target domain name (FQDN, recorded as a type A record) to point to the current public address
(or a custom IP)
* resolve any domain name (acting as a simple DNS client)

<br>

Get ddflare
====
ddflare is released as statically compiled binaries for different OS/architetures that you can grab
from the [release page](https://github.com/fgiudici/ddflare/releases/latest).

Get a x86_64 linux binary example:
```bash
wget https://github.com/fgiudici/ddflare/releases/download/v0.1.0/ddflare-linux-amd64
sudo install ddflare-linux-amd64 /usr/local/bin/ddflare
```

Container images are availble as well on the github repository. Run ddflare
via docker with:

```bash
docker run -ti --rm ghcr.io/fgiudici/ddflare:0.3.0
```

Quickstart
====
ddflare has two main commands:  `set`, to update the type A record of the target FQDN
to the current public ip (or a custom IP address) and `get`, to retrieve the current public IP address
(or resolve the FQDN passed as argument), and

Run `ddflare help` to display all the available commands and flags.

Update domain name via [DynDNS update prococol v3](https://help.dyn.com/remote-access-api/)
----
>[!NOTE]
>You should register to the DDNS service first, select your FQDN and retrieve the _username_ and _password_
>required to authenticate to the service

To update a domain name (FQDN, type A record) to the current public IP address of the host run:

```bash
ddflare set -s <DDNS_URL> -t <AUTH_TOKEN> <FQDN>
```
where `<DDNS_URL>` is the http endpoint of the DDNS service,
`<FQDN>` is the domain to be updated and `<AUTH_TOKEN>` is the API authentication token.

Real example with:
* [No-IP](https://www.noip.com/) DDNS provider
* `myhost.ddns.net` FQDN
* user `68816xj`
* password `v4UMHzugodpE`
```bash
ddflare set -s noip -t 68816xj:v4UMHzugodpE  myhost.ddns.net
```

Update domain name via [Cloudflare API](https://developers.cloudflare.com/api/)
----
>[!TIP]
>In order to update domain names you need a Cloudflare API token with `Zone.DNS` `Edit` permission.
>You can create one following the
>[Cloudflare docs](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/).

To update a domain name (FQDN, type A record) to the current public IP address of the host run:

```bash
ddflare set -t <CLOUDFLARE_TOKEN> <FQDN>
```
where `<FQDN>` is the domain to be updated and `<CLOUDFLARE_TOKEN>` is a Cloudflare API token.

Real example with FQDN `myhost.example.com` and
Cloudflare API token `gB1fOLbl2d5XynhfIOJvzX8Y4rZnU5RLPW1hg7cM`:
```bash
ddflare set -t gB1fOLbl2d5XynhfIOJvzX8Y4rZnU5RLPW1hg7cM  myhost.example.com
```

>[!NOTE]
>ddflare can <ins>update</ins> existing A records but cannot create new ones, so the record
>should be created in advance.
>
>To create an A type record see Cloudflare's
>[Manage DNS Records](https://developers.cloudflare.com/dns/manage-dns-records/how-to/create-dns-records/)
>docs.

>[!TIP]
>When creating a type A record pay attention to the value of the `TTL` field:
>it tracks the number of seconds DNS clients and DNS resolvers are allowed to
>cache the resolved IP address for the record.
>You may want to keep the TTL low (min is 60 secs) if you plan to use the record
>to track the (dynamic) IP address of a host in a DDNS scenario.

Get the current public IP address
----
Retrieving the current public IP address is as easy as:
```bash
ddflare get
```
ddflare queries the `ipify.org` service under the hood, which detects the public IP address used to reach the service.