package utils

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/asiffer/wg-easy-vpn/crypto"
	"github.com/rs/zerolog"
)

// KeyValue represents a key-value pair in a section
type KeyValue struct {
	Key   string
	Value string
}

// Section represents a basic block like [Interface] or [Peer]
type Section struct {
	name     string
	data     []KeyValue
	comments []string
}

// NewSection creates a new empty section
func NewSection(name string) *Section {
	return &Section{
		name:     name,
		data:     make([]KeyValue, 0),
		comments: make([]string, 0),
	}
}

// Name returns the name of the section
func (s *Section) Name() string {
	return s.name
}

// HasKey returns whether the section has the given key
func (s *Section) HasKey(key string) bool {
	for _, kv := range s.data {
		if kv.Key == key {
			return true
		}
	}
	return false
}

// Set defines a pair key/value (replaces existing key if present)
func (s *Section) Set(key string, value string) error {
	if err := checkKey(key); err != nil {
		return err
	}
	for i, kv := range s.data {
		if kv.Key == key {
			s.data[i].Value = value
			return nil
		}
	}
	s.data = append(s.data, KeyValue{Key: key, Value: value})
	return nil
}

// Add appends a key/value pair (allows duplicate keys)
func (s *Section) Add(key string, value string) error {
	if err := checkKey(key); err != nil {
		return err
	}
	s.data = append(s.data, KeyValue{Key: key, Value: value})
	return nil
}

// Get returns the raw value (string) related to a key (first match)
func (s *Section) Get(key string) (string, error) {
	for _, kv := range s.data {
		if kv.Key == key {
			return kv.Value, nil
		}
	}
	return "", fmt.Errorf("unknown key %s", key)
}

// GetInt returns a value given a key and tries to convert it
func (s *Section) GetInt(key string) (int, error) {
	value, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// GetUint16 returns a value given a key and tries to convert it
func (s *Section) GetUint16(key string) (uint16, error) {
	value, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(i), nil
}

// GetBytesFromBase64 returns the byte array representing the base64 encoded string
func (s *Section) GetBytesFromBase64(key string) ([]byte, error) {
	value, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GetKeyFromBase64 returns the key represented by the base64 encoded string
func (s *Section) GetKeyFromBase64(key string) (crypto.Key, error) {
	value, err := s.GetBytesFromBase64(key)
	if err != nil {
		return nil, err
	}
	k := crypto.NewKey()
	err = k.UpdateFromBytes(value)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// GetIPArray returns an array of net.IP objects parsed from the value
func (s *Section) GetIPArray(key string) ([]net.IP, error) {
	value, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	// split
	values := strings.Split(value, ",")
	ipList := make([]net.IP, len(values))
	for i, v := range values {
		// parse IP
		ip := net.ParseIP(strings.TrimSpace(v))
		// ip, _, err := net.ParseCIDR(strings.TrimSpace(v))
		if ip == nil {
			return nil, fmt.Errorf("error while parsing IP %s", s)
		}
		ipList[i] = ip
	}

	return ipList, nil
}

// GetNetworks returns an array of net.IPNet objects parsed from the value
func (s *Section) GetNetworks(key string) ([]net.IPNet, error) {
	addr, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("KEY: %s, VALUE: %s | FULL: %v\n", key, addr, strings.Split(addr, ","))
	networks := make([]net.IPNet, 0)
	for _, rawnet := range strings.Split(addr, ",") {
		if strings.TrimSpace(rawnet) == "" {
			continue
		}
		ip, network, err := net.ParseCIDR(strings.TrimSpace(rawnet))
		if err != nil {
			return nil, fmt.Errorf("error while parsing network %s", rawnet)
		}
		// re-set the initial IP (net.ParseCIDR sets it to the network address)
		network.IP = ip
		networks = append(networks, *network)
	}

	return networks, nil
}

func (s *Section) String() string {
	str := fmt.Sprintf("[%s]\n", s.name)
	return str + s.StringNoHeader()
}

func (s *Section) StringNoHeader() string {
	str := ""
	for _, comment := range s.comments {
		str += fmt.Sprintf("# %s\n", comment)
	}
	for _, kv := range s.data {
		str += fmt.Sprintf("%s = %s\n", kv.Key, kv.Value)
	}
	return str
}

func (s *Section) AddComment(comment string) {
	s.comments = append(s.comments, comment)
}

func (s *Section) Log(event *zerolog.Event) *zerolog.Event {
	for _, kv := range s.data {
		event = event.Str(fmt.Sprintf("%s.%s", s.name, kv.Key), kv.Value)
	}
	return event
}

func (s *Section) LogWithoutName(event *zerolog.Event) *zerolog.Event {
	for _, kv := range s.data {
		event = event.Str(kv.Key, kv.Value)
	}
	return event
}
