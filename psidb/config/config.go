package config

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/viper"
)

type Config struct {
	PeerID       peer.ID
	Identity     crypto.PrivKey
	IdentityFile string `mapstructure:"IDENTITY_FILE"`

	PublicAddresses []string
	ListenAddresses []string
	BootstrapPeers  []string

	PublicAddressList string `mapstructure:"PUBLIC_ADDRESS_LIST"`
	ListenAddressList string `mapstructure:"LISTEN_ADDRESS_LIST"`
	BootstrapPeerList string `mapstructure:"BOOTSTRAP_ADDRESS_LIST"`

	HttpEndpoint string `mapstructure:"HTTP_ENDPOINT"`

	GlobalPortOffset uint16 `mapstructure:"GLOBAL_PORT_OFFSET"`
	DisableIP6       bool   `mapstructure:"DISABLE_IP6"`

	DatabasePath         string `mapstructure:"DATABASE_PATH"`
	TrustMasterPublicKey string `mapstructure:"TRUST_MASTER_PUBLIC_KEY"`
}

func ReadConfig() (*Config, error) {
	cfg := &Config{}

	v := viper.New()
	v.SetEnvPrefix("psidb")
	v.SetConfigType("yaml")

	autoBindConfig(v, reflect.TypeOf(*cfg), "")

	v.AutomaticEnv()

	_ = v.ReadInConfig()

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if err := cfg.InitializeDefaults(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func autoBindConfig(v *viper.Viper, typ reflect.Type, base string) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		if !f.IsExported() {
			continue
		}

		tag, ok := f.Tag.Lookup("mapstructure")

		if !ok {
			tag = strcase.ToScreamingSnake(f.Name)
		}

		name := base + tag

		if f.Type.Kind() == reflect.Struct {
			autoBindConfig(v, f.Type, name+"_")
		} else {
			if err := v.BindEnv(name); err != nil {
				panic(err)
			}
		}
	}
}

func (c *Config) InitializeDefaults() error {
	if c.DatabasePath == "" {
		c.DatabasePath = "./.fti/psi"
	}

	if c.IdentityFile == "" {
		c.IdentityFile = path.Join(c.DatabasePath, "node.key")
	}

	if c.IdentityFile != "" {
		if err := c.ReadIdentityFile(); err != nil {
			return err
		}
	}

	if c.PeerID == "" && c.Identity != nil {
		pid, err := peer.IDFromPublicKey(c.Identity.GetPublic())

		if err != nil {
			return err
		}

		c.PeerID = pid
	}

	if len(c.ListenAddressList) > 0 {
		addresses := strings.Split(c.ListenAddressList, ";")

		c.ListenAddresses = append(c.ListenAddresses, addresses...)
	}

	if len(c.BootstrapPeerList) > 0 {
		addresses := strings.Split(c.BootstrapPeerList, ";")

		c.BootstrapPeers = append(c.BootstrapPeers, addresses...)
	}

	if len(c.ListenAddresses) == 0 {
		c.ListenAddresses = append(c.ListenAddresses,
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic",
			"/ip4/0.0.0.0/udp/0/quic-v1",
			"/ip4/0.0.0.0/udp/0/quic-v1/webtransport",
			"/ip6/::/tcp/0",
			"/ip6/::/udp/0/quic",
			"/ip6/::/udp/0/quic-v1",
			"/ip6/::/udp/0/quic-v1/webtransport",
		)
	}

	if len(c.ListenAddresses) == 0 {
		ifaces, err := net.Interfaces()

		if err != nil {
			panic(err)
		}

		for _, iface := range ifaces {
			if iface.Flags&net.FlagLoopback != 0 {
				continue
			}

			addrs, err := iface.Addrs()

			if err != nil {
				continue
			}

			for _, addr := range addrs {
				ip, ok := addr.(*net.IPNet)

				if !ok {
					continue
				}

				if ip.IP.IsLoopback() {
					continue
				}

				if ip.IP.IsInterfaceLocalMulticast() {
					continue
				}

				if ip.IP.IsMulticast() {
					continue
				}

				if ip.IP.IsLinkLocalMulticast() {
					continue
				}

				var addrStr string

				if ip.IP.To4() != nil {
					addrStr = fmt.Sprintf("/ip4/%s/tcp/9000", ip.IP.String())
				} else {
					addrStr = fmt.Sprintf("/ip6/%s/tcp/9000", ip.IP.String())
				}

				c.ListenAddresses = append(c.ListenAddresses, addrStr)
			}
		}
	}

	if c.HttpEndpoint == "" {
		c.HttpEndpoint = "/ip4/0.0.0.0/tcp/22440/http"
	}

	return nil
}

func (c *Config) GetBootstrapPeers() []peer.AddrInfo {
	result := make([]peer.AddrInfo, 0, len(c.BootstrapPeers))

	for _, addr := range c.BootstrapPeers {
		info, err := peer.AddrInfoFromString(addr)

		if err != nil {
			continue
		}

		if info == nil {
			continue
		}

		if info.ID == c.PeerID {
			continue
		}

		result = append(result, *info)
	}

	return result
}

func (c *Config) GenerateIdentity() error {
	if c.Identity != nil {
		return nil
	}

	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)

	if err != nil {
		return err
	}

	c.Identity = priv

	return nil
}

func (c *Config) ReadIdentityFile() error {
	if _, err := os.Stat(c.IdentityFile); os.IsNotExist(err) {
		if err := c.GenerateIdentity(); err != nil {
			return err
		}

		data, err := crypto.MarshalPrivateKey(c.Identity)

		if err != nil {
			return err
		}

		if err := os.MkdirAll(path.Dir(c.IdentityFile), 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(c.IdentityFile, os.O_CREATE|os.O_WRONLY, 0600)

		if err != nil {
			return err
		}

		defer f.Close()

		_, err = f.Write(data)

		if err != nil {
			return err
		}
	} else {
		f, err := os.OpenFile(c.IdentityFile, os.O_RDONLY, 0600)

		if err != nil {
			return err
		}

		defer f.Close()

		data, err := io.ReadAll(f)

		if err != nil {
			return err
		}

		key, err := crypto.UnmarshalPrivateKey(data)

		if err != nil {
			return err
		}

		c.Identity = key
	}

	data, err := crypto.MarshalPublicKey(c.Identity.GetPublic())

	if err != nil {
		return err
	}

	f, err := os.OpenFile(c.IdentityFile+".pub", os.O_CREATE|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)

	if err != nil {
		return err
	}

	return nil
}

func (c *Config) GetTrustMasterPublicKey() crypto.PubKey {
	if c.TrustMasterPublicKey == "" {
		return nil
	}

	data, err := base64.StdEncoding.DecodeString(c.TrustMasterPublicKey)

	if err != nil {
		return nil
	}

	pub, err := crypto.UnmarshalPublicKey(data)

	if err != nil {
		return nil
	}

	return pub
}
