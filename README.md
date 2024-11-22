<h1 align="center">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/fgiudici/ddflare/refs/heads/main/assets/logo/ddflare-logotype.svg#gh-light-mode-only" width=130 alt="logo">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/fgiudici/ddflare/refs/heads/main/assets/logo/ddflare-logotype-dark.svg#gh-dark-mode-only" width=130 alt="logo">
  <br>

  [![example workflow](https://github.com/fgiudici/ddflare/actions/workflows/build.yaml/badge.svg)](https://github.com/fgiudici/ddflare/actions/workflows/build.yaml)
  [![example workflow](https://github.com/fgiudici/ddflare/actions/workflows/container-image.yaml/badge.svg)](https://github.com/fgiudici/ddflare/actions/workflows/container-image.yaml)
</h1>

ddflare is a client to manage DNS type A records via the [Cloudflare API](https://developers.cloudflare.com/api/).

It's primary usage is to provide a [DDNS (Dynamic DNS)](https://en.wikipedia.org/wiki/Dynamic_DNS) client leveraging the Cloudflare API.

>[!IMPORTANT]
>Since ddflare operates via the [Cloudflare APIs](https://developers.cloudflare.com/api/),
>ddflare can update dns records for domains managed by Cloudflare only: this means
>your domain has been registered at or transfered to Cloudflare.

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

If you prefer a container image, you can pull ddflare from the github repository and run it
via docker as usual:

```bash
docker run -ti --rm ghcr.io/fgiudici/ddflare
```

Usage
====
ddflare has two main commands: `get`, to retrieve the current public IP address
(or resolve the FQDN passed as argument), and `set`, to update the type A record of the target FQDN
to the current public ip (or a custom IP address).

Run `ddflare help` to display all the available commands and flags.

Update domain name to current public IP address ([DDNS client](https://en.wikipedia.org/wiki/Dynamic_DNS))
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
>To create an A record see Cloudflare's
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