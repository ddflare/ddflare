<h1 align="center">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/ddflare/ddflare/refs/heads/main/assets/logo/ddflare-logotype.svg#gh-light-mode-only" width=130 alt="logo">
  <img align="center" style="padding-bottom:10px" src="https://raw.githubusercontent.com/ddflare/ddflare/refs/heads/main/assets/logo/ddflare-logotype-dark.svg#gh-dark-mode-only" width=130 alt="logo">
  <br>

  [![release workflow](https://github.com/ddflare/ddflare/actions/workflows/build.yaml/badge.svg)](https://github.com/ddflare/ddflare/actions/workflows/build.yaml)
  [![container workflow](https://github.com/ddflare/ddflare/actions/workflows/container-image.yaml/badge.svg)](https://github.com/ddflare/ddflare/actions/workflows/container-image.yaml)
</h1>

ddflare is a [DDNS (Dynamic DNS)](https://en.wikipedia.org/wiki/Dynamic_DNS) go library that allows DNS
record updates via either the [Cloudflare API](https://developers.cloudflare.com/api/) or
the [DynDNS update prococol v3](https://help.dyn.com/remote-access-api/perform-update/).
<br>
It comes with a CLI tool built on top of the library and released for different architectures.

ddflare allows to:
* update a target domain name (FQDN, recorded as a type A record) to point to the current public address
or a custom IP
* retrieve and display the current public IP address
* resolve any domain name (acting as a simple DNS client)

Project documentation at https://ddflare.org

