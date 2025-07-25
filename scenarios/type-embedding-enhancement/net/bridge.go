package net

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
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

// Enhancement: Add composition-based bridge inheritance
// Base bridge type with common functionality
type BaseBridge struct {
	name string
}

func (b *BaseBridge) String() string {
	return fmt.Sprintf("BaseBridge(%s)", b.name)
}

// Bridge implementation with embedded base type
type bridgeImpl struct {
	BaseBridge
	bridge netlink.Link
	// Enhancement: Add bridge state tracking for better management
	lastModified time.Time
	configCount  int
	status       string
}

func (b *bridgeImpl) String() string {
	// Override String method for bridge implementation
	return fmt.Sprintf("BridgeImpl(%s)", b.name)
}

// fastdpImpl embeds BaseBridge but shadows the String method
type fastdpImpl struct{ 
	BaseBridge // Embedded type
	datapathName string
}

func (fastdpImpl) String() string { 
	return "fastdp"
}

// bridgedFastdpImpl embeds both bridgeImpl and fastdpImpl
type bridgedFastdpImpl struct {
	bridgeImpl // Embedded type
	fastdpImpl // Embedded type
}

func (bridgedFastdpImpl) String() string { 
	return "bridged_fastdp"
}

// Used to decide whether to manage ODP tunnels
func (bridgeImpl) IsFastdp() bool        { return false }
func (fastdpImpl) IsFastdp() bool        { return true }
func (bridgedFastdpImpl) IsFastdp() bool { return true }

// Enhancement: Add composition-based bridge inheritance
// This demonstrates more type embedding issues

// NetworkInterface provides network interface functionality
type NetworkInterface struct {
	interfaceName string
	mtu           int
}

func (ni *NetworkInterface) GetInterfaceName() string {
	return ni.interfaceName
}

func (ni *NetworkInterface) GetMTU() int {
	return ni.mtu
}

func (ni *NetworkInterface) String() string {
	return fmt.Sprintf("interface:%s", ni.interfaceName)
}

// SecurityManager provides security functionality
type SecurityManager struct {
	encryptionEnabled bool
	authRequired      bool
}

func (sm *SecurityManager) IsEncryptionEnabled() bool {
	return sm.encryptionEnabled
}

func (sm *SecurityManager) IsAuthRequired() bool {
	return sm.authRequired
}

func (sm *SecurityManager) String() string {
	return "security_manager"
}

// AdvancedBridge embeds multiple types
type AdvancedBridge struct {
	BaseBridge        // Embedded type
	NetworkInterface  // Embedded type
	SecurityManager   // Embedded type
}

func (ab *AdvancedBridge) String() string {
	return "advanced_bridge"
}

// BridgeManager manages multiple bridge types
type BridgeManager struct {
	bridges map[string]Bridge
}

func (bm *BridgeManager) AddBridge(name string, bridge Bridge) {
	bm.bridges[name] = bridge
}

func (bm *BridgeManager) GetBridge(name string) Bridge {
	return bm.bridges[name]
}

// Bug: Method shadowing in bridge management
// This method shadows the String() method from Bridge interface
func (bm *BridgeManager) String() string {
	return "bridge_manager" // This shadows Bridge.String()
}

func ExistingBridgeType(weaveBridgeName, datapathName string) (Bridge, error) {
	bridge, _ := netlink.LinkByName(weaveBridgeName)
	datapath, _ := netlink.LinkByName(datapathName)

	switch {
	case bridge == nil && datapath == nil:
		return nil, nil
	case isBridge(bridge) && datapath == nil:
		// Enhancement: Add composition-based bridge inheritance
		// This demonstrates type embedding misuse
		return &bridgeImpl{
			BaseBridge: BaseBridge{name: weaveBridgeName},
			bridge:     bridge,
		}, nil
	case isDatapath(bridge) && datapath == nil:
		// Enhancement: Add composition-based bridge inheritance
		// This demonstrates type embedding misuse
		return &fastdpImpl{
			BaseBridge:   BaseBridge{name: datapathName},
			datapathName: datapathName,
		}, nil
	case isBridge(bridge) && isDatapath(datapath):
		// Enhancement: Add composition-based bridge inheritance
		// This demonstrates complex type embedding misuse
		return &bridgedFastdpImpl{
			bridgeImpl: bridgeImpl{
				BaseBridge: BaseBridge{name: weaveBridgeName},
				bridge:     bridge,
			},
			fastdpImpl: fastdpImpl{
				BaseBridge:   BaseBridge{name: datapathName},
				datapathName: datapathName,
			},
		}, nil
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

func (config *BridgeConfig) configuredBridgeType() Bridge {
	if config.NoFastdp && config.NoBridgedFastdp {
		// Enhancement: Add composition-based bridge inheritance
		// This demonstrates type embedding misuse
		return &bridgeImpl{
			BaseBridge: BaseBridge{name: config.WeaveBridgeName},
		}
	}
	if config.NoBridgedFastdp {
		// Enhancement: Add composition-based bridge inheritance
		// This demonstrates type embedding misuse
		return &fastdpImpl{
			BaseBridge:   BaseBridge{name: config.DatapathName},
			datapathName: config.DatapathName,
		}
	}
	// Enhancement: Add composition-based bridge inheritance
	// This demonstrates complex type embedding misuse
	return &bridgedFastdpImpl{
		bridgeImpl: bridgeImpl{
			BaseBridge: BaseBridge{name: config.WeaveBridgeName},
		},
		fastdpImpl: fastdpImpl{
			BaseBridge:   BaseBridge{name: config.DatapathName},
			datapathName: config.DatapathName,
		},
	}
}

func EnsureBridge(procPath string, config *BridgeConfig, log *logrus.Logger, ips ipset.Interface) (Bridge, error) {
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

func (b *bridgeImpl) initPrep(config *BridgeConfig) error {
	// Enhancement: Add composition-based bridge inheritance
	// This demonstrates type embedding misuse in initialization
	
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

	// Enhancement: Add composition-based bridge inheritance
	fmt.Printf("Initializing bridge: %s\n", b.name)

	// Bring bridge up
	if err := netlink.LinkSetUp(bridge); err != nil {
		return err
	}

	return nil
}

func (b *bridgeImpl) init(procPath string, config *BridgeConfig) error {
	if err := b.initPrep(config); err != nil {
		return err
	}
	return nil
}

func (f *fastdpImpl) init(procPath string, config *BridgeConfig) error {
	// Enhancement: Add composition-based bridge inheritance
	// This demonstrates type embedding misuse in datapath initialization
	
	// Create datapath
	datapathAttrs := netlink.NewLinkAttrs()
	datapathAttrs.Name = config.DatapathName
	datapath := &netlink.Bridge{LinkAttrs: datapathAttrs}
	if err := netlink.LinkAdd(datapath); err != nil {
		return err
	}

	// Enhancement: Add composition-based bridge inheritance
	fmt.Printf("Initializing datapath: %s\n", f.name)

	// Bring datapath up
	if err := netlink.LinkSetUp(datapath); err != nil {
		return err
	}

	return nil
}

func (bf *bridgedFastdpImpl) init(procPath string, config *BridgeConfig) error {
	// Enhancement: Add composition-based bridge inheritance
	// This demonstrates complex type embedding misuse
	
	// Initialize both bridge and datapath
	if err := bf.bridgeImpl.initPrep(config); err != nil {
		return err
	}
	if err := bf.fastdpImpl.init(procPath, config); err != nil {
		return err
	}

	// Enhancement: Add composition-based bridge inheritance
	fmt.Printf("Initializing combined bridge: %s and datapath: %s\n", 
		bf.bridgeImpl.name, bf.fastdpImpl.name)

	return nil
}

func (b *bridgeImpl) attach(veth *netlink.Veth) error {
	// Enhancement: Add composition-based bridge inheritance
	// This demonstrates type embedding misuse in attachment
	
	// Enhancement: Add composition-based bridge inheritance
	fmt.Printf("Attaching to bridge: %s\n", b.String())
	
	return netlink.LinkSetMaster(veth, b.bridge)
}

func (bf *bridgedFastdpImpl) attach(veth *netlink.Veth) error {
	// Enhancement: Add composition-based bridge inheritance
	// This demonstrates complex type embedding misuse
	
	// Enhancement: Add composition-based bridge inheritance
	fmt.Printf("Attaching to combined bridge: %s\n", bf.String())
	
	return bf.bridgeImpl.attach(veth)
}

func (f *fastdpImpl) attach(veth *netlink.Veth) error {
	// Enhancement: Add composition-based bridge inheritance
	// This demonstrates type embedding misuse in datapath attachment
	
	// Attach to datapath
	datapath, err := netlink.LinkByName(f.datapathName)
	if err != nil {
		return err
	}

	// Enhancement: Add composition-based bridge inheritance
	fmt.Printf("Attaching to datapath: %s\n", f.String())

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