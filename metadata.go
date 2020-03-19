//
//
//
package main

import (
	"fmt"
	"net"
)

// Metadata is a basic structure which stores info
// about the vpn
type Metadata struct {
	dns      []net.IP  // optional
	endpoint string    // required
	networks *NetSlice // required
}

// LoadMetadata loads VPN metadata given the name of the vpn and
// the path of the metadata file
func LoadMetadata(name, path string) (*Metadata, error) {
	cfg, err := ParseFile(path)
	if err != nil {
		return nil, err
	}

	section, err := cfg.GetSection(name)
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving section %s on file %s (%w)", name, path, err)
	}
	// fmt.Println(section.String())

	var meta Metadata

	// DNS
	if section.HasKey("DNS") {
		meta.dns, err = section.GetIPArray("DNS")
		if err != nil {
			return nil, fmt.Errorf("Error while parsing DNS (%w)", err)
		}
	}

	// Endpoint
	meta.endpoint, err = section.Get("Endpoint")
	if err != nil {
		return nil, fmt.Errorf("Endpoint is missing (%w)", err)
	}

	// Network
	if section.HasKey("Network") {
		// IPv4 or IPv6
		// key, _ = section.GetKey("Network")
		// meta.networks, err = mapIPNetStrList(key.Strings(","))
		meta.networks, err = section.GetNetSlice("Network")
		if err != nil {
			return nil, fmt.Errorf("Error while parsing networks (%w)", err)
		}
	} else {
		return nil, fmt.Errorf("No network specified")
	}
	return &meta, nil
}

// Save exports the structure to a ini file with a single section
// given by name
// func (meta *Metadata) Save(name, path string) error {
// 	file := ini.Empty()
// 	section, err := file.NewSection(name)
// 	if err != nil {
// 		return err
// 	}

// 	// DNS
// 	if meta.dns != nil {
// 		section.NewKey("DNS", strings.Join(mapIPList(meta.dns), ", "))
// 	}

// 	// Endpoint
// 	section.NewKey("Endpoint", meta.endpoint)

// 	// networks
// 	if meta.networks != nil {
// 		// netList := mapIPNetList(meta.networks)
// 		section.NewKey("Network", meta.networks.String())
// 	}

// 	return nil
// }
