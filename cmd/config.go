package cmd

import (
	"net"

	"github.com/asiffer/wg-easy-vpn/utils"
)

const (
	// DefaultConfigSuffix is the .conf extension of the configuration files
	DefaultConfigSuffix = ".conf"
	// DefaultServerConfigDirectory is the path where the server configuration file is stored
	DefaultServerConfigDirectory = "/etc/wireguard"
	// DefaultClientConfigDirectory is the path where the client configuration files are stored
	DefaultClientConfigDirectory = "/etc/wireguard/clients"
	// DefaultNetwork is the VPN network when it is not specified
	DefaultNetwork = "10.8.0.1/24"
	// DefaultListeningPort is the default UDP port the server listens
	DefaultListeningPort = uint16(52820)
	// DefaultConnectionName is the name commonly used
	DefaultConnectionName = "wg0"
	// DefaultMetadataFile is the name of the file where metadata are stored
	DefaultMetadataFile = ".wg-easy-vpn.conf"
	// DefaultQRCodeFormat is the extension of the image file containing qrcode
	DefaultQRCodeFormat = "png"
)

const WIREGUARD_DIR = "/etc/wireguard"
const CONFIG_SUFFIX = ".conf"

// var (
// 	connName     = DefaultConnectionName
// 	serverDir    = DefaultServerConfigDirectory
// 	clientDir    = DefaultClientConfigDirectory
// 	routes       = []string{"0.0.0.0/0", "::/0"}
// 	networks     = []string{DefaultNetwork}
// 	noPSK        = false
// 	force        = false
// 	port         = DefaultListeningPort
// 	endpoint     = ""
// 	dns          = []string{}
// 	clients      = []string{}
// 	exportFormat = DefaultQRCodeFormat
// 	export       = false
// 	keepFile     = false
// 	out          = ""
// )

// var config *Config

type Config struct {
	connName  string
	serverDir string
	clientDir string
	routes    []string
	networks  []string
	noPSK     bool
	force     bool
	port      uint16
	endpoint  string
	dns       []string
	client    string
	qrcode    bool
	keepFile  bool
}

func defaultConfig() *Config {
	return &Config{
		connName:  "",
		serverDir: DefaultServerConfigDirectory,
		clientDir: DefaultClientConfigDirectory,
		routes:    []string{"0.0.0.0/0", "::/0"},
		networks:  []string{DefaultNetwork},
		noPSK:     false,
		force:     false,
		port:      DefaultListeningPort,
		endpoint:  "",
		dns:       []string{},
		client:    "",
		qrcode:    false,
		keepFile:  false,
	}
}

// var config *Config = defaultConfig()

// func resetConfig() {
// 	config.connName = ""
// 	config.serverDir = DefaultServerConfigDirectory
// 	config.clientDir = DefaultClientConfigDirectory
// 	config.routes = []string{"0.0.0.0/0", "::/0"}
// 	config.networks = []string{DefaultNetwork}
// 	config.noPSK = false
// 	config.force = false
// 	config.port = DefaultListeningPort
// 	config.endpoint = ""
// 	config.dns = []string{}
// 	config.client = ""
// 	config.qrcode = false
// 	config.keepFile = false
// }

// func init() {
// 	var err error
// 	config, err = NewConfig()
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func NewConfig() (*Config, error) {
// 	c := &Config{
// 		_config:      puzzle.NewConfig(),
// 		connName:     DefaultConnectionName,
// 		serverDir:    DefaultServerConfigDirectory,
// 		clientDir:    DefaultClientConfigDirectory,
// 		routes:       []string{"0.0.0.0/0", "::/0"},
// 		networks:     []string{DefaultNetwork},
// 		noPSK:        false,
// 		force:        false,
// 		port:         DefaultListeningPort,
// 		endpoint:     "",
// 		dns:          []string{},
// 		clients:      []string{},
// 		exportFormat: DefaultQRCodeFormat,
// 		export:       false,
// 		keepFile:     false,
// 		out:          "",
// 	}

// 	if err := puzzle.DefineVar(c._config, "conn", &c.connName, puzzle.WithDescription("Name of the network connection to use")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "server-dir", &c.serverDir, puzzle.WithDescription("Directory to store server configurations")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "client-dir", &c.clientDir, puzzle.WithDescription("Directory to store client configurations")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "no-psk", &c.noPSK, puzzle.WithDescription("Disable the use of preshared keys")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "force", &c.force, puzzle.WithDescription("Force overwriting existing files")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "port", &c.port, puzzle.WithDescription("Port for the WireGuard server to listen on")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "endpoint", &c.endpoint, puzzle.WithDescription("Endpoint for the WireGuard server (e.g. mydomain.com:51820)")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "dns", &c.dns, puzzle.WithDescription("DNS servers for the clients (comma separated)")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "routes", &c.routes, puzzle.WithDescription("Additional routes to push to clients (comma separated)")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "networks", &c.networks, puzzle.WithDescription("Additional networks to allow the server to access (comma separated)")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "clients", &c.clients, puzzle.WithDescription("List of clients to create (comma separated)")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "export-format", &c.exportFormat, puzzle.WithDescription("Export format for client configurations (png or jpg)")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "export", &c.export, puzzle.WithDescription("Export client configurations as QR codes")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "keep-file", &c.keepFile, puzzle.WithDescription("Keep the temporary configuration file when exporting as QR code")); err != nil {
// 		return nil, err
// 	}
// 	if err := puzzle.DefineVar(c._config, "out", &c.out, puzzle.WithDescription("Output file")); err != nil {
// 		return nil, err
// 	}
// 	return c, nil
// }

// func (c *Config) initFlags() []cli.Flag {
// 	filtered := c._config.Only("networks", "endpoint", "dns", "no-psk", "force", "port", "out")
// 	flags, err := urfave3.Build(filtered)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return flags
// }

// func Networks() []net.IPNet {
// 	ns, err := utils.ParseIPNetList(networks)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return ns
// }

// func Routes() []net.IPNet {
// 	ns, err := utils.ParseIPNetList(routes)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return ns
// }

// func DNS() []net.IP {
// 	ns, err := utils.ParseIPList(dns)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return ns
// }

// func (c *Config) SetConn(name string) {
// 	c.connName = name
// }

func (c *Config) Networks() []net.IPNet {
	ns, err := utils.ParseIPNetList(c.networks)
	if err != nil {
		panic(err)
	}
	return ns
}

func (c *Config) Routes() []net.IPNet {
	ns, err := utils.ParseIPNetList(c.routes)
	if err != nil {
		panic(err)
	}
	return ns
}

func (c *Config) DNS() []net.IP {
	ns, err := utils.ParseIPList(c.dns)
	if err != nil {
		panic(err)
	}
	return ns
}

// func (c *Config) ensureClientDirectoryExist() error {
// 	return os.MkdirAll(c.clientDir, 0744)
// }

// func (c *Config) SaveClient(
// 	name string,
// 	server *models.WGServer,
// 	client *models.WGClient,
// 	endpoint string) error {
// 	// check directory
// 	if err := c.ensureClientDirectoryExist(); err != nil {
// 		return fmt.Errorf("error while creating client config directory %s (%v)",
// 			c.clientDir, err)
// 	}

// 	// check if client file exist
// 	file := path.Join(c.clientDir, name+DefaultConfigSuffix)
// 	if utils.FileExist(file) && !c.force {
// 		return fmt.Errorf("a config file already exist for client %s", name)
// 	}
// 	// new config file
// 	f := utils.NewFile()

// 	// client section ([Interface])
// 	sec := f.AddSection("Interface")
// 	client.Section(sec)

// 	ns := make([]net.IPNet, 0)
// 	for _, r := range c.routes {
// 		ipnet, err := parseAddressAndMask(r)
// 		if err != nil {
// 			return err
// 		}
// 		ns.Append(ipnet)
// 	}

// 	// server as peer
// 	peer := server.ToPeer(ns, endpoint)
// 	// add client PSK to server as peer
// 	peer.psk = crypto.PresharedKey(client.PSK())

// 	// Peer section
// 	sec = f.AddSection("Peer")
// 	peer.Section(sec)

// 	// save
// 	return f.Save(file)
// }
