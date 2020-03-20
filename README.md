# wg-easy-vpn
A tool to ease Wireguard VPN setup

wg-easy-vpn is a tool designed to ease the set-up of a
[WireGuard](https://www.wireguard.com/) VPN. In particular you can easily create a server
and then add clients. You can also export the clients
configurations through QR codes.
When your vpn is set up, you just have to invoke `wg-quick`
for instance.

## Installation

### From sources

Basically you download the binary from this repo:

```bash
go install github.com/asiffer/wg-easy-vpn
```

The advantage is that the tool is build according to your architecture. The drawback is
the need to have `Go` installed on your host.

### Debian package

I also made a debian package to make it available on various debian-like platforms.

## Usage

We suppose you have a server with a public address
(reachable through the following domain name: wg.example.net), and you
want to connect some clients to it.
First, we create the server:

```bash
wg-easy-vpn create --endpoint wg.example.net
```