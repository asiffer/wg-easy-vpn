# wg-easy-vpn
A command-line tool to ease Wireguard VPN setup

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

Debian packages will be soon available for different architectures
to distribute `wg-easy-vpn` to various debian-like platforms.

## Usage

We suppose you have a server with a public address
(reachable through the following domain name: wg.example.net), and you
want to connect some clients to it. 
By default server files will be located in `/etc/wireguard` and clients
files will be located in `/etc/wireguard/clients`, therefore the following
commands are likely to be run as root.

First, let us create the server (`wg0` is the name of the connection):

```bash
wg-easy-vpn create --endpoint wg.example.net wg0
```

Then you probably need to add several clients:

```bash
wg-easy-vpn add -c iphone -c myDesktop wg0
```

Now you can transfer the clients' configuration files
to the right locations. You can also add the `--qrcode-cli`
option to print QR code to the cli (android app can take 
this qr code as input).

## Advanced usage

## Issues