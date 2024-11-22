<h1 align="center">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/fgiudici/ddflare/refs/heads/main/assets/logo/ddflare-logotype.svg#gh-light-mode-only" width=130 alt="logo">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/fgiudici/ddflare/refs/heads/main/assets/logo/ddflare-logotype-dark.svg#gh-dark-mode-only" width=130 alt="logo">
  <br>

  [![example workflow](https://github.com/fgiudici/ddflare/actions/workflows/build.yaml/badge.svg)](https://github.com/fgiudici/ddflare/actions/workflows/build.yaml)
  [![example workflow](https://github.com/fgiudici/ddflare/actions/workflows/container-image.yaml/badge.svg)](https://github.com/fgiudici/ddflare/actions/workflows/container-image.yaml)
</h1>

ddflare is a client to manage DNS type A records via [Cloudflare API](https://developers.cloudflare.com/api/).

It's primary usage is to enable DynDNS functionality via the Cloudflare API.

>[!IMPORTANT]
>Since ddflare operates via [Cloudflare APIs](https://developers.cloudflare.com/api/),
>ddflare can update dns records for domains managed by Cloudflare only: this means
>your have registered or transfered it at Cloudflare.

ddflare allows to:
* retrieve the current public IP address
* update domain names (FQDNs, recorded as type A records) to point to the current public address (or a custom IP)
* resolve any domain name (acting as a DNS client)

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

If you prefer container images, you can pull ddflare from the github repository and run the binary
via docker as usual:

```bash
docker run -ti --rm ghcr.io/fgiudici/ddflare
```

Usage
====
ddflare has two main commands: `get` and `set` to retrieve the current public IP address
(or resolve the FQDN passed as argument) and to set (update) the type A record of the target FQDN
to the current public ip (or a custom IP address).

Run `ddflare help` to display all the available commands and flags.

Update domain name to current public IP address (DynDNS client)
----
>[!TIP]
>In order to be able to update a domain name you need a Cloudflare API token.
>If you don't have one already, create it following the
>[Cloudflare docs](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/).

To update a domain name (FQDN, type A record) to the current public IP address of the host run:

```bash
ddflare set -t $CLOUDFLARE_TOKEN $FQDN
```
where `$FQDN` is the domain to be updated and `$CLOUDFLARE_TOKEN` is a Cloudflare API token.

>[!NOTE]
>ddflare can just update existing A records and not create new ones, so the record
>should have been already created (see Cloudflare's
>[Manage DNS Records](https://developers.cloudflare.com/dns/manage-dns-records/how-to/create-dns-records/)
>).
>
>When creating the type A record pay attention to the value of the `TTL` field:
>it represent the number of seconds DNS clients and DNS resolvers are allowed to
>cache the record.
>You may want to **keep the lowest value possible** (60 secs) if you are going to use it
>to track a dynamic IP address as in the DynDNS scenario.


Get current public IP address
----
As easy as:
```bash
ddflare get
```
ddflare queries the `ipify.org` service under the hood, returning the detected public IP address.