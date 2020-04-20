//
//
//
package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

const (
	// DefaultConfigSuffix is the .conf extension of the configuration files
	DefaultConfigSuffix = ".conf"
	// DefaultServerConfigDirectory is the path where the server configuration file is stored
	DefaultServerConfigDirectory = "/etc/wireguard"
	// DefaultClientConfigDirectory is the path where the client configuration files are stored
	DefaultClientConfigDirectory = "/etc/wireguard/clients"
	// DefaultNetwork is the VPN network when it is not specified
	DefaultNetwork = "192.168.0.1/24"
	// DefaultListeningPort is the default UDP port the server listens
	DefaultListeningPort = 52820
	// DefaultConnectionName is the name commonly used
	DefaultConnectionName = "wg0"
	// DefaultMetadataFile is the name of the file where metadata are stored
	DefaultMetadataFile = ".wg-easy-vpn.conf"
	// DefaultQRCodeFormat is the extension of the image file containing qrcode
	DefaultQRCodeFormat = "png"
)

// Application
var (
	app *cli.App
)

// colors
var (
	green      = color.New(color.FgGreen)
	greenBold  = color.New(color.FgGreen, color.Bold)
	yellow     = color.New(color.FgYellow)
	yellowBold = color.New(color.FgYellow, color.Bold)
)

// Runtime variables (set from cli.Context)
type Runtime struct {
	connName     string
	serverDir    string
	clientDir    string
	routes       *cli.StringSlice
	networks     *cli.StringSlice
	noPSK        bool
	force        bool
	port         int
	endpoint     string
	dns          *cli.StringSlice
	clients      *cli.StringSlice
	exportFormat string
	export       bool
	keepFile     bool
}

// RT is the global runtime of wg-easy-vpn. It gathers the default config
// and the cli args. It is initialized in the init() function.
var RT Runtime

func initRuntime() {
	RT.connName = DefaultConnectionName
	RT.serverDir = DefaultServerConfigDirectory
	RT.clientDir = DefaultClientConfigDirectory
	RT.routes = cli.NewStringSlice("0.0.0.0/0", "::/0")
	RT.networks = cli.NewStringSlice()
	RT.noPSK = false
	RT.force = false
	RT.port = DefaultListeningPort
	RT.endpoint = ""
	RT.dns = cli.NewStringSlice()
	RT.clients = cli.NewStringSlice()
	RT.exportFormat = ""
	RT.export = false
	RT.keepFile = false
}

var appDescription = `
	wg-easy-vpn is a tool designed to ease the set-up of a WireGuard VPN.  In 
	particular you can easily create a server and then add clients.  You  can 
	also  export  the clients configurations through QR codes.  When your vpn 
	is set up, you just have to invoke wg or wg-quick.`

func initApp() {
	app = &cli.App{
		Name:                   "wg-easy-vpn",
		ArgsUsage:              "interface",
		Version:                "1.0b",
		Authors:                []*cli.Author{&cli.Author{Name: "Alban Siffer", Email: "alban.siffer@gmail.com"}},
		Copyright:              "wg-easy-vpn source code is written under GPLv3 license (https://www.gnu.org/licenses/gpl-3.0.en.html)",
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Usage:       "Setup a Wireguard VPN simply",
		Description: appDescription,
		Action:      func(c *cli.Context) error { return nil },
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new Wireguard VPN from scratch",
				Action:    cmdCreate,
				Before:    setConnectionName,
				ArgsUsage: "<interface>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "no-psk",
						Usage:       "Do not generate preshared keys",
						Destination: &RT.noPSK,
						Required:    false,
					},
					&cli.PathFlag{
						Name:        "server-dir",
						Aliases:     []string{"d"},
						Value:       DefaultServerConfigDirectory,
						DefaultText: "directory",
						Destination: &RT.serverDir,
						Usage:       "Directory to store the server configuration",
					},
					&cli.StringSliceFlag{
						Name:        "net",
						Aliases:     []string{"n"},
						Usage:       "VPN networks",
						DefaultText: "network",
						Destination: RT.networks,
						Value:       cli.NewStringSlice(DefaultNetwork),
					},
					&cli.IntFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Usage:       "Listening UDP port",
						Destination: &RT.port,
						DefaultText: "port",
						Value:       DefaultListeningPort,
					},
					&cli.StringFlag{
						Name:        "endpoint",
						Aliases:     []string{"e"},
						Usage:       "Address or domain name of the server",
						DefaultText: "address",
						Destination: &RT.endpoint,
						Required:    true,
					},
					&cli.StringSliceFlag{
						Name:        "dns",
						Usage:       "IP address of the DNS to use",
						DefaultText: "ip",
						Destination: RT.dns,
						Required:    false,
					},
					&cli.BoolFlag{
						Name:        "force",
						Aliases:     []string{"f"},
						Usage:       "Override possible previous config",
						Destination: &RT.force,
						Value:       false,
						Required:    false,
					},
				},
			},
			{
				Name:      "add",
				Usage:     "Add a new client to the VPN",
				ArgsUsage: "[wg conn]",
				Action:    cmdAdd,
				Before:    setConnectionName,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "no-psk",
						Usage:       "Do not generate preshared keys",
						Destination: &RT.noPSK,
					},
					&cli.StringFlag{
						Name:        "server-dir",
						Value:       DefaultServerConfigDirectory,
						Usage:       "Directory to store the server configuration",
						DefaultText: "directory",
						Destination: &RT.serverDir,
					},
					&cli.StringFlag{
						Name:        "client-dir",
						Value:       DefaultClientConfigDirectory,
						Usage:       "Directory to store client configurations",
						DefaultText: "directory",
						Destination: &RT.clientDir,
					},
					&cli.StringSliceFlag{
						Name:        "client",
						Aliases:     []string{"c"},
						Usage:       "Client to add",
						DefaultText: "name",
						Required:    true,
						Destination: RT.clients,
					},
					&cli.StringSliceFlag{
						Name:    "route",
						Aliases: []string{"r"},
						Usage:   "Default routes managed by the VPN",
						// Value:       cli.NewStringSlice("0.0.0.0/0", "::/0"),
						DefaultText: "network",
						Destination: RT.routes,
					},
					&cli.StringSliceFlag{
						Name:        "dns",
						Usage:       "IP address of the DNS to use",
						DefaultText: "ip",
						Destination: RT.dns,
					},
					&cli.BoolFlag{
						Name:        "force",
						Aliases:     []string{"f"},
						Usage:       "Override possible previous config",
						Required:    false,
						Destination: &RT.force,
					},
					&cli.BoolFlag{
						Name:        "export",
						Aliases:     []string{"e"},
						Usage:       "Export the config to an image (png or jpg) through a qrcode",
						Required:    false,
						Value:       false,
						Destination: &RT.export,
					},
					&cli.StringFlag{
						Name:        "export-format",
						Aliases:     []string{"t"},
						Usage:       "Define the format of the exported qrcode (txt, jpg or png). If this flag is not set, the qrcode is printed to stdout",
						Required:    false,
						DefaultText: "format",
						Destination: &RT.exportFormat,
					},
				},
			},
			{
				Name:      "show",
				Usage:     "Show the clients of the VPN",
				ArgsUsage: "[wg conn]",
				Action:    cmdShow,
				Before:    setConnectionName,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "server-dir",
						Value:       DefaultServerConfigDirectory,
						Usage:       "Directory of the server configuration",
						DefaultText: "directory",
						Destination: &RT.serverDir,
					},
					&cli.StringFlag{
						Name:        "client-dir",
						Value:       DefaultClientConfigDirectory,
						Usage:       "Directory of the client configuration",
						DefaultText: "directory",
						Destination: &RT.clientDir,
					},
				},
			},
			{
				Name:      "rm",
				Usage:     "Remove a client from the VPN",
				Action:    cmdRm,
				Before:    setConnectionName,
				ArgsUsage: "[wg conn]",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:        "client",
						Aliases:     []string{"c"},
						Usage:       "Client to remove",
						DefaultText: "name/public-key",
						Required:    true,
						Destination: RT.clients,
					},
					&cli.BoolFlag{
						Name:        "keep-file",
						Aliases:     []string{"k"},
						Usage:       "Client config file will be kept",
						Required:    false,
						Destination: &RT.keepFile,
					},
					&cli.StringFlag{
						Name:        "server-dir",
						Value:       DefaultServerConfigDirectory,
						Usage:       "Directory of the server configuration",
						DefaultText: "directory",
						Destination: &RT.serverDir,
					},
					&cli.StringFlag{
						Name:        "client-dir",
						Value:       DefaultClientConfigDirectory,
						Usage:       "Directory of the client configuration",
						DefaultText: "directory",
						Destination: &RT.clientDir,
					},
				},
			},
		},
	}
}

func init() {
	// runtime
	initRuntime()
	// init App
	initApp()
}

func setConnectionName(c *cli.Context) error {
	args := c.Args()
	if c.NArg() == 0 {
		return fmt.Errorf("The name of the connection is not given")
	}
	last := args.Get(args.Len() - 1)
	RT.connName = cleanString(last)
	return nil
}

func saveServer(server *WGServer) error {
	// file := path.Join(c.Path("server-dir"), name+DefaultConfigSuffix)
	file := path.Join(RT.serverDir, RT.connName+DefaultConfigSuffix)
	if fileExist(file) && !RT.force {
		return fmt.Errorf("Configuration file already exists")
	}
	f := NewFile()
	sec := f.AddSection("Interface")
	server.Section(sec)
	return f.Save(file)
}

func ensureClientDirectoryExist() error {
	return os.MkdirAll(RT.clientDir, 0744)
}

func saveMetadata(name string, meta *Metadata) error {
	file := path.Join(RT.serverDir, DefaultMetadataFile)
	f := NewFile()
	if fileExist(file) && !RT.force {
		f, err := ParseFile(file)
		if err != nil {
			return fmt.Errorf("Error while loading metadata file (%w)", err)
		}
		if f.HasSection(name) {
			return fmt.Errorf("The file %s already contains information about the %s connection", file, name)
		}
	}

	sec := f.AddSection(name)

	if err := sec.Set("Endpoint", meta.endpoint); err != nil {
		return fmt.Errorf("Error while creating 'Endpoint' key: %v", err)
	}

	if err := sec.Set("Network", meta.networks.String()); err != nil {
		return fmt.Errorf("Error while creating 'Network' key: %v", err)
	}

	if len(meta.dns) > 0 {
		if err := sec.Set("DNS", strings.Join(mapIPList(meta.dns), ", ")); err != nil {
			return fmt.Errorf("Error while creating 'DNS' key: %v", err)
		}
	}

	return f.Save(file)
}

func saveClient(name string,
	server *WGServer,
	client *WGClient,
	endpoint string) error {
	// check directory
	if err := ensureClientDirectoryExist(); err != nil {
		return fmt.Errorf("Error while creating client config directory %s (%v)",
			RT.clientDir, err)
	}

	// check if client file exist
	file := path.Join(RT.clientDir, name+DefaultConfigSuffix)
	if fileExist(file) && !RT.force {
		return fmt.Errorf("A config file already exist for client %s", name)
	}
	// new config file
	f := NewFile()

	// client section ([Interface])
	sec := f.AddSection("Interface")
	client.Section(sec)

	ns := NewNetSlice()
	for _, r := range RT.routes.Value() {
		ipnet, err := parseAddressAndMask(r)
		if err != nil {
			return err
		}
		ns.Append(ipnet)
	}

	// server as peer
	peer := server.ToPeer(&ns, endpoint)
	// add client PSK to server as peer
	peer.psk = PresharedKey(client.psk)

	// Peer section
	sec = f.AddSection("Peer")
	peer.Section(sec)

	// save
	return f.Save(file)
}

func cmdCreate(c *cli.Context) error {
	// Server ---
	var err error

	// Get networks
	nets := NewNetSlice()
	if c.IsSet("net") {
		nets, err = NewNetSliceFromStringSlice(RT.networks.Value())
		if err != nil {
			return err
		}
	} else {
		n, err := parseAddressAndMask(DefaultNetwork)
		if err != nil {
			return err
		}
		nets.Append(n)
	}

	// create server
	server := NewWGServer(&nets, !RT.noPSK, RT.port)

	// Metadata ---
	meta := Metadata{}

	// endpoint
	meta.endpoint = c.String("endpoint")

	// networks
	meta.networks = &nets

	// dns
	if c.IsSet("dns") {
		meta.dns, err = mapIPStrList(RT.dns.Value())
		if err != nil {
			return fmt.Errorf("Error while mapping DNS IP (%v)", err)
		}
	}

	// save server
	if err := saveServer(server); err != nil {
		return fmt.Errorf("Error while saving server configuration: %v", err)
	}
	green.Printf("The connection %s has been set up (%s)\n",
		RT.connName, path.Join(RT.serverDir, RT.connName+DefaultConfigSuffix))

	// save metadata
	if err := saveMetadata(RT.connName, &meta); err != nil {
		return fmt.Errorf("Error while saving connection metadata: %v", err)
	}
	green.Printf("Metadata about the connection %s has been saved to %s\n",
		RT.connName, path.Join(RT.serverDir, DefaultMetadataFile))

	return nil
}

func exportClientConfig(clientName string) error {
	// get the path of the client file
	file := path.Join(RT.clientDir, clientName+DefaultConfigSuffix)
	// read
	r, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error while opening %s (%v)", file, err)
	}
	// outfile
	var w *os.File
	switch strings.ToLower(RT.exportFormat) {
	case "png", "jpg", "txt":
		file = strings.Replace(file, DefaultConfigSuffix, "."+RT.exportFormat, 1)
		w, err = os.Create(file)
		if err != nil {
			return fmt.Errorf("Error while creating output file %s (%v)", file, err)
		}
		break
	default:
		w = os.Stdout
		break
	}

	if err := ExportConfig(r, w); err != nil {
		return fmt.Errorf("Error while exporting qrcode (%v)", err)
	}
	r.Close()

	if w != os.Stdout {
		w.Close()
	}

	// changing permissions
	if os.Chmod(file, 0600); err != nil {
		return fmt.Errorf("Error while changing permissions of %s (%v)", file, err)
	}
	return nil
}

func mergeDNS(metadataDNS []net.IP) ([]net.IP, error) {
	dns := make([]net.IP, len(metadataDNS))
	copy(dns, metadataDNS)
	for _, ipstr := range RT.dns.Value() {
		ip := net.ParseIP(ipstr)
		if ip == nil {
			return nil, fmt.Errorf("Error while parsing DNS ip: %s", ipstr)
		}
		if findIP(ip, dns) < 0 {
			dns = append(dns, ip)
		}
	}
	return dns, nil
}

func cmdAdd(c *cli.Context) error {
	// Server ---

	// Read the VPN config
	connPath := path.Join(RT.serverDir, RT.connName+DefaultConfigSuffix)
	vpn, err := ReadVPN(connPath)
	if err != nil {
		return err
	}

	// get the path where metadata are stored
	metaPath := path.Join(RT.serverDir, DefaultMetadataFile)
	if err = vpn.AddMetadata(metaPath); err != nil {
		return err
	}

	// DNS override?
	dns, err := mergeDNS(vpn.metadata.dns)
	if err != nil {
		return fmt.Errorf("Error while merging DNS IP")
	}

	// now we are ready to create clients
	clients := make([]*WGClient, len(RT.clients.Value()))
	for i, clientName := range RT.clients.Value() {
		// assign a new netslice
		baseIP, err := vpn.ProvideNetSlice()
		if err != nil {
			return err
		}
		// create client
		clients[i] = NewWGClient(baseIP, !c.Bool("no-psk"), dns)
		// save client config
		if err := saveClient(clientName, vpn.server, clients[i], vpn.metadata.endpoint); err != nil {
			return fmt.Errorf("Error while saving client '%s': %v", clientName, err)
		}
		green.Printf("Client %s has been added (%s) to %s\n",
			clientName,
			path.Join(RT.clientDir, clientName+DefaultConfigSuffix),
			RT.connName,
		)
		// qrcode ?
		if RT.export {
			if err := exportClientConfig(clientName); err != nil {
				return err
			}
		}
		// add client to vpn (as peer)
		vpn.peers = append(vpn.peers, clients[i].ToPeer())
	}

	return vpn.Save(connPath)
}

func cmdShow(c *cli.Context) error {
	// Server ---

	// Read the VPN config
	connPath := path.Join(RT.serverDir, RT.connName+DefaultConfigSuffix)
	vpn, err := ReadVPN(connPath)
	if err != nil {
		return err
	}

	// retrieve public keys related to this connection
	keys := vpn.PeerPublicKeys()

	// extract key->client map from the client folder
	pairs := extractPairsFromFolder(RT.clientDir, false)

	// print config name
	greenBold.Print("interface")
	fmt.Print(": ")
	green.Println(RT.connName)

	// print clients
	for _, k := range keys {
		yellow.Printf("\t%s", k)
		fmt.Print(": ")
		if n, exists := pairs[k]; exists {
			yellowBold.Println(n)
		} else {
			yellowBold.Println("???")
		}
	}
	return nil
}

func cmdRm(c *cli.Context) error {
	// Read the VPN config
	connPath := path.Join(RT.serverDir, RT.connName+DefaultConfigSuffix)
	vpn, err := ReadVPN(connPath)
	if err != nil {
		return err
	}

	publicKeyToRemove := make([]string, 0)
	filesToRemove := make([]string, 0)

	// extract client_name->key map from the client folder
	pairs := extractPairsFromFolder(RT.clientDir, true)

	for _, c := range RT.clients.Value() {
		if key, exists := pairs[c]; exists {
			publicKeyToRemove = append(publicKeyToRemove, key)
			filesToRemove = append(filesToRemove,
				path.Join(RT.clientDir, c+DefaultConfigSuffix))
		} else {
			// maybe a public key has been given
			publicKeyToRemove = append(publicKeyToRemove, c)
		}
	}

	pk := NewKey()
	for _, pkstr := range publicKeyToRemove {
		if pk.UpdateFromBase64(pkstr) == nil {
			// remove publickey
			if vpn.RemovePeerFromPublicKey(pk) != nil {
				yellowBold.Printf("The client with publicKey %s has not been found\n",
					pk.Base64())
			} else {
				yellow.Printf("The client with public key %s has been removed from %s\n",
					pk.Base64(), RT.connName)
			}
		} else {
			yellowBold.Printf("The key %s is not valid\n", pkstr)
		}
	}

	// remove client configuration files
	if !RT.keepFile {
		for _, f := range filesToRemove {
			if err := os.Remove(f); err != nil {
				return fmt.Errorf("Failed to remove %s (%v)", f, err)
			}
			yellow.Printf("The file %s has been removed\n", f)
		}
	}

	return vpn.Save(connPath)
}
