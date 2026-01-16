# wg-easy-vpn

![Build](https://github.com/asiffer/wg-easy-vpn/workflows/Build/badge.svg)
![Test](https://github.com/asiffer/wg-easy-vpn/workflows/Test/badge.svg)
[![Coverage](https://codecov.io/gh/asiffer/wg-easy-vpn/branch/master/graph/badge.svg)](https://codecov.io/gh/asiffer/wg-easy-vpn)


![logo](assets/logo.png)

Setup a Wireguard VPN simply

---

wg-easy-vpn is a tool designed to ease the set-up of a[WireGuard®](https://www.wireguard.com/) VPN. 


## Get started

Download binary from the [latest release](https://github.com/asiffer/wg-easy-vpn/releases/latest).

```shell
curl -L -o wg-easy-vpn https://github.com/asiffer/wg-easy-vpn/releases/latest/download/wg-easy-vpn-amd64
```

Init a server configuration.

```shell
wg-easy-vpn init --endpoint wg.example.org wg0
```

You can then start the server with [wg-quick](https://man7.org/linux/man-pages/man8/wg-quick.8.html) for instance:

```shell
wg-quick up wg0
```

You can add a client (it will print its config to stdout).

```shell
wg-easy-vpn add -c new-client wg0
```


## Advanced configuration

You can customize the VPN through flags. 
In the following example, we set
- an overall DNS that will be provided to every client (unless customized)
- a custom network (default to `10.8.0.0/24`) when server and client IPs will be picked up
- custom routes tunneled by the VPN (passed to the clients) instead of all (`0.0.0.0/0` and `::/0`)
- custom listening UDP port (default to `52820`)

```shell
wg-easy-vpn init \
    --dns 1.1.1.1 \
    --networks 192.168.42.0/24 \
    --routes 0.0.0.0/0 \
    --port 2820 \
    --endpoint wg.example.org \
    wg0
```

On the client side, you can re-define the overall options passed to the server, and also export the config to a qrcode (stdout):

```shell
wg-easy-vpn add \
    -c new-client \
    --dns 8.8.8.8 \
    --routes 192.168.42.0/24 \
    --qrcode \
    wg0
```


## Changelog

**1.0b1**
- Full rewrite (modular approach)
- wg-easy-vpn config embedded in the server configuration file (top section)
- no more show command 
- client config is not saved to disk (stdout)

**1.0b**
- Better IP provisionning 
- automatic doc generation
- manpages in debian package
- some fixes around DNS override

**1.0a**

For this early release, the tool does not manage very well IP of clients when the number of 
clients is high or when the specified mask size is greater that 24 (/30 may not be well supported for instance).

Moreover, the IP (re-)assignement is likely to fail after a client has been removed. I will try to fix it firstly.

## Next 

- Support `PostUp` and `PostDown` options
- Manage `server-dir` and `client-dir` directly in the `.wg-easy-vpn.conf` file
- I think that many bugs are likely to occur, so I will probably spend time to test and fix.

