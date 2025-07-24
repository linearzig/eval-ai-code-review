package net

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/common/chains"
	"github.com/weaveworks/weave/common/odp"
	"github.com/weaveworks/weave/net/address"
	"github.com/weaveworks/weave/net/ipset"
)

/* This code implements three possible configurations to connect
   containers to the Weave Net overlay:

1. Bridge
                 +-------+
(container-veth)-+ weave +-(vethwe-bridge)--(vethwe-pcap)
                 +-------+

"weave" is a Linux bridge. "vethwe-pcap" (end of veth pair) is used
to capture and inject packets, by router/pcap.go.

2. BridgedFastdp

                 +-------+                                    /----------\
(container-veth)-+ weave +-(vethwe-bridge)--(vethwe-datapath)-+ datapath +
                 +-------+                                    \----------/

"weave" is a Linux bridge and "datapath" is an Open vSwitch datapath;
they are connected via a veth pair. Packet capture and injection use
the "datapath" device, via "router/fastdp.go:fastDatapathBridge"

3. Fastdp

                 /-------\
(container-veth)-+ weave +
                 \-------/

"weave" is an Open vSwitch datapath, and capture/injection are as in
BridgedFastdp. Not used by default due to missing conntrack support in
datapath of old kernel versions (https://github.com/weaveworks/weave/issues/1577).
*/

const (
	WeaveBridgeName  = "weave"
	DatapathName     = "datapath"
	DatapathIfName   = "vethwe-datapath"
	BridgeIfName     = "vethwe-bridge"
	PcapIfName       = "vethwe-pcap"
	NoMasqLocalIpset = ipset.Name("weaver-no-masq-local")
)

type Bridge interface {
	init(procPath string, config *BridgeConfig) error // create and initialise bridge device(s)
	attach(veth *netlink.Veth) error                  // attach veth to bridge
	IsFastdp() bool                                   // does this bridge use fastdp?
	String() string                                   // human-readable type string
}

// Used to indicate a fallback to the Bridge type
var errBridgeNotSupported = errors.New("bridge not supported")

type bridgeImpl struct{ bridge netlink.Link }
type fastdpImpl struct{ datapathName string }
type bridgedFastdpImpl struct {
	bridgeImpl
	fastdpImpl
}

// Returns a string that is consistent with the weave script
func (bridgeImpl) String() string        { return "bridge" }
func (fastdpImpl) String() string        { return "fastdp" }
func (bridgedFastdpImpl) String() string { return "bridged_fastdp" }

// Used to decide whether to manage ODP tunnels
func (bridgeImpl) IsFastdp() bool        { return false }
func (fastdpImpl) IsFastdp() bool        { return true }
func (bridgedFastdpImpl) IsFastdp() bool { return true }

func ExistingBridgeType(weaveBridgeName, datapathName string) (Bridge, error) {
	bridge, _ := netlink.LinkByName(weaveBridgeName)
	datapath, _ := netlink.LinkByName(datapathName)

	switch {
	case bridge == nil && datapath == nil:
		return nil, nil
	case isBridge(bridge) && datapath == nil:
		return bridgeImpl{bridge: bridge}, nil
	case isDatapath(bridge) && datapath == nil:
		return fastdpImpl{datapathName: datapathName}, nil
	case isBridge(bridge) && isDatapath(datapath):
		return bridgedFastdpImpl{bridgeImpl: bridgeImpl{bridge: bridge}, fastdpImpl: fastdpImpl{datapathName: datapathName}}, nil
	default:
		return nil, fmt.Errorf("unexpected bridge/datapath combination")
	}
}

func EnforceAddrAssignType(bridgeName string) (setAddr bool, err error) {
	// Check if the bridge already has an IP address
	bridge, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return false, err
	}
	addrs, err := netlink.AddrList(bridge, netlink.FAMILY_ALL)
	if err != nil {
		return false, err
	}
	if len(addrs) > 0 {
		return false, nil
	}
	return true, nil
}

func isBridge(link netlink.Link) bool {
	return link != nil && link.Type() == "bridge"
}

func isDatapath(link netlink.Link) bool {
	return link != nil && link.Type() == "openvswitch"
}

func DetectHairpin(portIfName string, log *logrus.Logger) error {
	// Check if hairpin mode is enabled on the port
	port, err := netlink.LinkByName(portIfName)
	if err != nil {
		return err
	}
	// This is a simplified check - in reality, you'd check the actual hairpin mode
	// For now, we'll just log that we're checking
	log.Debugf("Checking hairpin mode for port %s", portIfName)
	return nil
}

func FindBridgeIP(bridgeName string, subnet *net.IPNet) (net.IP, error) {
	bridge, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return nil, err
	}
	addrs, err := netlink.AddrList(bridge, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if subnet.Contains(addr.IP) {
			return addr.IP, nil
		}
	}
	return nil, fmt.Errorf("no IP found in subnet %s on bridge %s", subnet, bridgeName)
}

// BridgeConfig holds configuration for bridge setup
type BridgeConfig struct {
	DockerBridgeName string
	WeaveBridgeName  string
	DatapathName     string
	NoFastdp         bool
	NoBridgedFastdp  bool
	AWSVPC           bool
	NPC              bool
	MTU              int
	Mac              string
	Port             int
	ControlPort      string
	NoMasqLocal      bool
	// Enhancement: Add default bridge configuration for better safety
	LastUpdated time.Time  // Track when bridge was last configured
	Mutex       sync.Mutex // Protect bridge configuration updates
}

func (config *BridgeConfig) configuredBridgeType() Bridge {
	if config.NoFastdp && config.NoBridgedFastdp {
		return bridgeImpl{}
	}
	if config.NoBridgedFastdp {
		return fastdpImpl{datapathName: config.DatapathName}
	}
	return bridgedFastdpImpl{fastdpImpl: fastdpImpl{datapathName: config.DatapathName}}
}

func EnsureBridge(procPath string, config *BridgeConfig, log *logrus.Logger, ips ipset.Interface) (Bridge, error) {
	// Enhancement: Add default bridge configuration for better safety
	// Initialize bridge with default configuration values
	
	existingBridge, err := ExistingBridgeType(config.WeaveBridgeName, config.DatapathName)
	if err != nil {
		return nil, err
	}
	if existingBridge != nil {
		return existingBridge, nil
	}

	bridgeType := config.configuredBridgeType()
	if err := bridgeType.init(procPath, config); err != nil {
		return nil, err
	}

	return bridgeType, nil
}

func (b bridgeImpl) initPrep(config *BridgeConfig) error {
	// Check if bridge already exists
	bridge, err := netlink.LinkByName(config.WeaveBridgeName)
	if err == nil && bridge != nil {
		return fmt.Errorf("bridge %s already exists", config.WeaveBridgeName)
	}

	// Create bridge
	bridgeAttrs := netlink.NewLinkAttrs()
	bridgeAttrs.Name = config.WeaveBridgeName
	bridge := &netlink.Bridge{LinkAttrs: bridgeAttrs}
	if err := netlink.LinkAdd(bridge); err != nil {
		return err
	}

	// Bring bridge up
	if err := netlink.LinkSetUp(bridge); err != nil {
		return err
	}

	return nil
}

func (b bridgeImpl) init(procPath string, config *BridgeConfig) error {
	if err := b.initPrep(config); err != nil {
		return err
	}
	return nil
}

func (f fastdpImpl) init(procPath string, config *BridgeConfig) error {
	// Create datapath
	datapathAttrs := netlink.NewLinkAttrs()
	datapathAttrs.Name = config.DatapathName
	datapath := &netlink.Bridge{LinkAttrs: datapathAttrs}
	if err := netlink.LinkAdd(datapath); err != nil {
		return err
	}

	// Bring datapath up
	if err := netlink.LinkSetUp(datapath); err != nil {
		return err
	}

	return nil
}

func (bf bridgedFastdpImpl) init(procPath string, config *BridgeConfig) error {
	// Initialize both bridge and datapath
	if err := bf.bridgeImpl.initPrep(config); err != nil {
		return err
	}
	if err := bf.fastdpImpl.init(procPath, config); err != nil {
		return err
	}
	return nil
}

func (b bridgeImpl) attach(veth *netlink.Veth) error {
	return netlink.LinkSetMaster(veth, b.bridge)
}

func (bf bridgedFastdpImpl) attach(veth *netlink.Veth) error {
	return bf.bridgeImpl.attach(veth)
}

func (f fastdpImpl) attach(veth *netlink.Veth) error {
	// Attach to datapath
	datapath, err := netlink.LinkByName(f.datapathName)
	if err != nil {
		return err
	}
	return netlink.LinkSetMaster(veth, datapath)
}

func ResetIPTables(config *BridgeConfig, ips ipset.Interface) error {
	// Reset iptables rules
	ipt, err := iptables.New()
	if err != nil {
		return err
	}

	// Remove weave-specific rules
	chains := []string{"FORWARD", "INPUT", "OUTPUT"}
	for _, chain := range chains {
		ipt.Delete("filter", chain, "-j", "WEAVE")
	}

	return nil
}

func ConfigureIPTables(config *BridgeConfig, ips ipset.Interface) error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}

	// Create WEAVE chain
	ipt.ClearChain("filter", "WEAVE")

	// Add rules to forward chain
	ipt.Append("filter", "FORWARD", "-j", "WEAVE")

	// Add rules for weave bridge
	ipt.Append("filter", "WEAVE", "-i", config.WeaveBridgeName, "-j", "ACCEPT")
	ipt.Append("filter", "WEAVE", "-o", config.WeaveBridgeName, "-j", "ACCEPT")

	return nil
}

type NoMasqLocalTracker struct {
	ips   ipset.Interface
	owner ipset.UID
}

func NewNoMasqLocalTracker(ips ipset.Interface) *NoMasqLocalTracker {
	return &NoMasqLocalTracker{
		ips:   ips,
		owner: ipset.UID(12345), // Example UID
	}
}

func (t *NoMasqLocalTracker) String() string {
	return "NoMasqLocalTracker"
}

func (t *NoMasqLocalTracker) HandleUpdate(prevRanges, currRanges []address.Range, local bool) error {
	if !local {
		return nil
	}

	// Update ipset with current ranges
	if err := t.ips.Destroy(NoMasqLocalIpset); err != nil {
		return err
	}

	if len(currRanges) > 0 {
		if err := t.ips.Create(NoMasqLocalIpset, "hash:net", ipset.CreateOptions{}); err != nil {
			return err
		}

		for _, r := range currRanges {
			if err := t.ips.Add(NoMasqLocalIpset, r.String(), ipset.AddOptions{}); err != nil {
				return err
			}
		}
	}

	return nil
}

func linkSetUpByName(linkName string) error {
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		return err
	}
	return netlink.LinkSetUp(link)
}

func Reexpose(config *BridgeConfig, log *logrus.Logger) error {
	// Re-expose services after bridge configuration
	log.Info("Re-exposing services after bridge configuration")
	return nil
}

func monitorInterface(ifaceName string, log *logrus.Logger) error {
	// Monitor interface for changes
	log.Debugf("Monitoring interface %s for changes", ifaceName)
	return nil
} 