# wg-easy-vpn

![Build](https://github.com/asiffer/wg-easy-vpn/workflows/Build/badge.svg)
![Test](https://github.com/asiffer/wg-easy-vpn/workflows/Test/badge.svg)
[![codecov](https://codecov.io/gh/asiffer/wg-easy-vpn/branch/master/graph/badge.svg)](https://codecov.io/gh/asiffer/wg-easy-vpn)



A command-line tool to ease Wireguard VPN setup

wg-easy-vpn is a tool designed to ease the set-up of a
[WireGuard](https://www.wireguard.com/) VPN. In particular you can easily create a server
and then add clients. You can also export the clients
configurations through QR codes.
When your vpn is set up, you just have to invoke `wg-quick`
for instance.

## Installation

### From sources

Basically you can download the sources from this repo and install it with
`go` tools:

```bash
go get -u github.com/asiffer/wg-easy-vpn
go install github.com/asiffer/wg-easy-vpn
```

The advantage is that the tool is build according to your architecture. The drawback is the need to have `Go` installed on your host.

### Debian package

Debian packages will be soon available for different architectures
to distribute `wg-easy-vpn` to various debian-like platforms.

## Usage

We suppose you have a server with a public address
(reachable through the following domain name: wg.example.net), and you
want to connect some clients to it. 
By default server files are located in `/etc/wireguard` and clients
files are located in `/etc/wireguard/clients`, therefore the following
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
to the right locations. You can also add the `--export`
flag to print QR code to the cli (android app can notably take 
this QR code as input).

Finally you can remove some clients:
```bash
wg-easy-vpn rm -c iphone wg0
```

## Advanced usage

### Custom server

By default `wg-easy-vpn` makes the server listen on port 52820, but this 
can be changed with the `--port` option:

```bash
wg-easy-vpn create --endpoint wg.example.net --port 10000 wg0
```

When you create a server, you can define a custom DNS (even several). This can be added to your configuration through the `--dns` option.

```bash
wg-easy-vpn create --endpoint wg.example.net --dns 1.1.1.1 wg0
```

The VPN created by `wg-easy-vpn` uses the network `192.168.0.0/24`. It can be modified with the `--net` option:

```bash
wg-easy-vpn create --endpoint wg.example.net --net 10.10.10.0/16 wg0
```

As previously said, the server configuration is saved to `/etc/wireguard` 
(plus some metadata saved in the `.wg-easy-vpn` file). The parameter 
`--server-dir` can be used to customize the location of these files.


### Custom clients

By default `wg-easy-vpn` creates VPN where all the clients' trafic is 
routed through (`0.0.0.0/0` and `::/0`). You can restrict theses routes:

```bash
wg-easy-vpn add -c newDevice --route "10.0.0.0/8" wg0
```

You can export clients config through QR code with the `--export` flag.
In this case the QR code is printed to the terminal but you can saved it
to an image file instead by setting `--export-format` (`jpg`, `png` and `txt` are recognized). The image file is saved to the clients directory.

```bash
wg-easy-vpn add -c newDevice --export --export-format png wg0
```

The client configuration is saved to `/etc/wireguard/clients` by default 
The parameter `--client-dir` can be used to customize the location of these files.

## Issues

Currently, this tool does not manage very well IP of clients when the number of 
clients is high or when the specified mask size is greater that 24 (/30 may not
be well supported for instance).