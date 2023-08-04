package config

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

type LocalResourceManager interface {
	NetworkListenAddresses() []multiaddr.Multiaddr

	ManagementEndpoint() net.Addr
	IpfsGatewayEndpoint() net.Addr
	RelayListenEndpoint() net.Addr

	ListenMultiaddr(name string) multiaddr.Multiaddr
	ListenMultiaddrs(name string) []multiaddr.Multiaddr
	ListenEndpoint(name string) net.Addr
	ListenEndpoints(name string) []net.Addr
	StoragePath(path string) string
	ManagedNetworkInterface(s string) string
}

type localResourceManager struct {
	config *Config

	listenAddresses map[string][]multiaddr.Multiaddr
}

func NewLocalResourceManager(config *Config) LocalResourceManager {
	lrm := &localResourceManager{
		config:          config,
		listenAddresses: map[string][]multiaddr.Multiaddr{},
	}

	lrm.SetListenEndpoints("p2p", config.ListenAddresses)
	lrm.SetListenEndpoint("management", config.HttpEndpoint)
	lrm.SetListenEndpoint("relay", config.RelayEndpoint)
	lrm.SetListenEndpoint("ipfs-gateway", config.IpfsGatewayEndpoint)
	lrm.SetListenEndpoint("dns", config.DnsBindEndpoint)
	lrm.SetListenEndpoint("ssh", "/ip4/0.0.0.0/tcp/2222")

	return lrm
}

func (l *localResourceManager) SetListenEndpoint(name, endpoint string) {
	var addrs []string

	if endpoint != "" {
		addrs = []string{endpoint}
	}

	l.SetListenEndpoints(name, addrs)
}

func (l *localResourceManager) SetListenEndpoints(name string, endpoints []string) {
	addrs := make([]multiaddr.Multiaddr, 0, len(endpoints))

	for _, str := range endpoints {
		for _, addrStr := range strings.Split(str, ";") {
			ma, err := multiaddr.NewMultiaddr(addrStr)

			if err != nil {
				panic(err)
			}

			addrs = append(addrs, ma)
		}
	}

	l.SetListenMultiaddrs(name, addrs)
}

func (l *localResourceManager) SetListenMultiaddrs(name string, addrs []multiaddr.Multiaddr) {
	l.listenAddresses[name] = addrs
}

func (l *localResourceManager) ListenMultiaddr(name string) multiaddr.Multiaddr {
	addrs := l.ListenMultiaddrs(name)

	if len(addrs) == 0 {
		return nil
	}

	if len(addrs) > 1 {
		panic("cannot use multiple addresses")
	}

	return addrs[0]
}

func (l *localResourceManager) ListenMultiaddrs(name string) []multiaddr.Multiaddr {
	return l.listenAddresses[name]
}

func (l *localResourceManager) ListenEndpoint(name string) net.Addr {
	addrs := l.ListenEndpoints(name)

	if len(addrs) == 0 {
		return nil
	}

	if len(addrs) > 1 {
		panic("cannot use multiple addresses")
	}

	return addrs[0]
}
func (l *localResourceManager) ListenEndpoints(name string) []net.Addr {
	addrs, ok := l.listenAddresses[name]

	if !ok {
		panic("no port named: " + name)
	}

	results := make([]net.Addr, len(addrs))

	for i, v := range addrs {
		ma := l.patchMultiAddr(v)

		multiaddr.ForEach(ma, func(c multiaddr.Component) bool {
			switch c.Protocol().Code {
			case multiaddr.P_IP4:
			case multiaddr.P_IP6:
			case multiaddr.P_TCP:
			case multiaddr.P_UDP:
			case multiaddr.P_DNS:
			case multiaddr.P_DNS4:
			case multiaddr.P_DNS6:
				// Nothing

			default:
				ma = ma.Decapsulate(&c)
			}

			return true
		})

		a, err := manet.ToNetAddr(ma)

		if err != nil {
			endpointUrl, err := MultiaddrToUrl(v)

			if err != nil {
				panic(err)
			}

			udpPort, isUdp := ma.ValueForProtocol(multiaddr.P_UDP)

			if isUdp == nil && udpPort != "" {
				ip, err := net.ResolveUDPAddr("udp", endpointUrl.Host)

				if err != nil {
					panic(err)
				}

				a = ip
			} else {
				ip, err := net.ResolveTCPAddr("tcp", endpointUrl.Host)

				if err != nil {
					panic(err)
				}

				a = ip
			}
		}

		results[i] = a
	}

	return results
}

func (l *localResourceManager) ManagedNetworkInterface(name string) string {
	return fmt.Sprintf("%s%d", name, l.config.GlobalPortOffset)
}

func (l *localResourceManager) StoragePath(s string) string {
	p := path.Join(l.config.DatabasePath, l.config.PeerID.String(), s)
	p, err := filepath.Abs(p)

	if err != nil {
		panic(err)
	}

	_ = os.MkdirAll(p, 0755)

	return p
}

func (l *localResourceManager) PublicListenAddresses() []multiaddr.Multiaddr {
	var addrs []multiaddr.Multiaddr
	var candidates []multiaddr.Multiaddr

	if len(l.config.PublicAddresses) > 0 {
		candidates = make([]multiaddr.Multiaddr, len(l.config.PublicAddresses))

		for i, v := range l.config.PublicAddresses {
			ma, err := multiaddr.NewMultiaddr(v)

			if err != nil {
				panic(err)
			}

			candidates[i] = ma
		}
	} else {
		candidates = l.NetworkListenAddresses()
	}

	for _, addr := range candidates {
		skip := false

		multiaddr.ForEach(addr, func(c multiaddr.Component) bool {
			switch c.Protocol().Code {
			case multiaddr.P_IP4:
				fallthrough

			case multiaddr.P_IP6:
				ip := netip.MustParseAddr(c.Value())

				if !ip.IsValid() {
					skip = true
				} else if ip.IsLoopback() {
					skip = true
				}
			}

			if skip {
				return false
			}

			return true
		})

		if skip {
			continue
		}

		addrs = append(addrs, addr)
	}

	sort.SliceStable(addrs, func(i, j int) bool {
		a := addrs[i]
		b := addrs[j]
		s1 := 0
		s2 := 0

		if v, err := a.ValueForProtocol(multiaddr.P_IP4); err == nil && v != "" {
			s1++
		}

		if v, err := b.ValueForProtocol(multiaddr.P_IP4); err == nil && v != "" {
			s2++
		}

		if s1 != s2 {
			return s1 < s2
		}

		return a.String() < b.String()
	})

	return addrs
}

func (l *localResourceManager) NetworkListenAddresses() []multiaddr.Multiaddr {
	addrs := make([]multiaddr.Multiaddr, 0, len(l.config.ListenAddresses))

	for _, addr := range l.config.ListenAddresses {
		skip := false
		ma, err := multiaddr.NewMultiaddr(addr)

		if err != nil {
			continue
		}

		var components []multiaddr.Multiaddr

		multiaddr.ForEach(ma, func(c multiaddr.Component) bool {
			switch c.Protocol().Code {
			case multiaddr.P_IP6:
				if l.config.DisableIP6 {
					skip = true
				}
			case multiaddr.P_TCP:
				c = l.patchMultiAddrComponent(c)
			case multiaddr.P_UDP:
				c = l.patchMultiAddrComponent(c)
			}

			components = append(components, &c)

			if skip {
				return false
			}

			return true
		})

		if skip {
			continue
		}

		ma = multiaddr.Join(components...)

		addrs = append(addrs, ma)
	}

	return addrs
}

func (l *localResourceManager) ManagementEndpoint() net.Addr {
	return l.ListenEndpoint("management")
}

func (l *localResourceManager) IpfsGatewayEndpoint() net.Addr {
	return l.ListenEndpoint("ipfs-gateway")
}

func (l *localResourceManager) RelayListenEndpoint() net.Addr {
	return l.ListenEndpoint("relay")
}

func (l *localResourceManager) patchAddresses(network string, endpoints []string) []net.Addr {
	result := make([]net.Addr, len(endpoints))

	for i, v := range endpoints {
		result[i] = l.patchAddress(network, v)
	}

	return result
}

func (l *localResourceManager) patchAddress(network, endpoint string) net.Addr {
	addr, err := netip.ParseAddrPort(endpoint)

	if err != nil {
		panic(err)
	}

	addr = netip.AddrPortFrom(addr.Addr(), addr.Port()+l.config.GlobalPortOffset)

	return localAddress{network: network, endpoint: addr.String()}
}

func (l *localResourceManager) patchMultiAddr(ma multiaddr.Multiaddr) multiaddr.Multiaddr {
	var components []multiaddr.Multiaddr

	multiaddr.ForEach(ma, func(c multiaddr.Component) bool {
		switch c.Protocol().Code {
		case multiaddr.P_TCP:
			c = l.patchMultiAddrComponent(c)
		case multiaddr.P_UDP:
			c = l.patchMultiAddrComponent(c)
		}

		components = append(components, &c)

		return true
	})

	return multiaddr.Join(components...)
}

func (l *localResourceManager) patchMultiAddrComponent(c multiaddr.Component) multiaddr.Component {
	port, err := strconv.ParseUint(c.Value(), 10, 16)

	if err != nil {
		panic(err)
	}

	port = port + uint64(l.config.GlobalPortOffset)

	patched, err := multiaddr.NewComponent(c.Protocol().Name, strconv.FormatUint(port, 10))

	if err != nil {
		panic(err)
	}

	return *patched
}

type localAddress struct {
	network  string
	endpoint string
}

func (l localAddress) Network() string {
	return l.network
}

func (l localAddress) String() string {
	return l.endpoint
}

func MultiaddrToUrl(addr multiaddr.Multiaddr) (result *url.URL, err error) {
	var hostname string
	var port string

	result = &url.URL{}

	multiaddr.ForEach(addr, func(c multiaddr.Component) bool {
		switch c.Protocol().Code {
		case multiaddr.P_IP4:
			hostname = c.Value()

		case multiaddr.P_IP6:
			hostname = fmt.Sprintf("[%s]", c.Value())

		case multiaddr.P_DNS:
			fallthrough
		case multiaddr.P_DNS4:
			fallthrough
		case multiaddr.P_DNS6:
			hostname = c.Value()

		case multiaddr.P_TCP:
			fallthrough
		case multiaddr.P_UDP:
			port = c.Value()

		case multiaddr.P_HTTP:
			if result.Scheme == "tls" {
				result.Scheme = "https"
			} else {
				result.Scheme = "http"
			}

			result.Path = c.Value()

		case multiaddr.P_TLS:
			if result.Scheme == "http" {
				result.Scheme = "https"
			} else {
				result.Scheme = "tls"
			}

		case multiaddr.P_HTTPS:
			result.Scheme = "https"

			result.Path = c.Value()

		default:
			err = errors.New("invalid protocol")
			return false
		}

		return true
	})

	result.Host = fmt.Sprintf("%s:%s", hostname, port)

	return
}
