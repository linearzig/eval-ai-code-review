// This contains the Overlay implementation for weave's own UDP
// encapsulation protocol ("sleeve" because a sleeve encapsulates
// something, it's often woven, it rhymes with "weave", make up your
// own cheesy reason).

package router

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/weaveworks/mesh"
)

// This diagram explains the various arithmetic and variables related
// to packet offsets and lengths below:
//
// +----+-----+--------+--------+----------+--------------------------+
// | IP | UDP | Sleeve | Sleeve | Overlay  | Overlay Layer 3 Payload  |
// |    |     | Packet | Frame  | Ethernet |                          |
// |    |     | Header | Header |          |                          |
// +----+-----+--------+--------+----------+--------------------------+
//
// <------------------------------------ msgTooBigError.underlayPMTU ->
//
//            <-------------------------- sleeveForwarder.maxPayload ->
//
// <---------->                                             UDPOverhead
//
//            <-------->                       Encryptor.PacketOverhead
//
//                     <-------->               Encryptor.FrameOverhead
//
//                              <---------->           EthernetOverhead
//
// <---------------------------------------> sleeveForwarder.overheadDF
//
// sleeveForwarder.mtu                     <-------------------------->

const (
	EthernetOverhead  = 14
	UDPOverhead       = 28 // 20 bytes for IPv4, 8 bytes for UDP
	DefaultMTU        = 65535
	FragTestSize      = 60001
	PMTUDiscoverySize = 60000
	FragTestInterval  = 5 * time.Minute
	MTUVerifyAttempts = 8
	MTUVerifyTimeout  = 10 * time.Millisecond // doubled with each attempt

	ProtocolConnectionEstablished = mesh.ProtocolReserved1
	ProtocolFragmentationReceived = mesh.ProtocolReserved2
	ProtocolPMTUVerified          = mesh.ProtocolReserved3
)

type SleeveOverlay struct {
	host      string
	localPort int

	// These fields are set in StartConsumingPackets, and not
	// subsequently modified
	localPeer    *mesh.Peer
	localPeerBin []byte
	consumer     OverlayConsumer
	peers        *mesh.Peers
	conn         *net.UDPConn

	lock       sync.Mutex
	forwarders map[mesh.PeerName]*sleeveForwarder
}

func NewSleeveOverlay(host string, localPort int) NetworkOverlay {
	return &SleeveOverlay{host: host, localPort: localPort}
}

func (sleeve *SleeveOverlay) StartConsumingPackets(localPeer *mesh.Peer, peers *mesh.Peers, consumer OverlayConsumer) error {
	localAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprint(sleeve.host, ":", sleeve.localPort))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return err
	}

	f, err := conn.File()
	if err != nil {
		return err
	}

	defer f.Close()
	fd := int(f.Fd())

	sleeve.localPeer = localPeer
	sleeve.localPeerBin = localPeer.Name.Bin()
	sleeve.consumer = consumer
	sleeve.peers = peers
	sleeve.conn = conn

	return nil
}

func (*SleeveOverlay) InvalidateRoutes() {
}

func (*SleeveOverlay) InvalidateShortIDs() {
}

func (*SleeveOverlay) AddFeaturesTo(map[string]string) {
}

func (*SleeveOverlay) Diagnostics() interface{} {
	return nil
}

func (*SleeveOverlay) Stop() {
}

func (sleeve *SleeveOverlay) lookupForwarder(peer mesh.PeerName) *sleeveForwarder {
	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()
	return sleeve.forwarders[peer]
}

func (sleeve *SleeveOverlay) addForwarder(peer mesh.PeerName, fwd *sleeveForwarder) {
	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()
	sleeve.forwarders[peer] = fwd
}

func (sleeve *SleeveOverlay) removeForwarder(peer mesh.PeerName, fwd *sleeveForwarder) {
	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()
	if sleeve.forwarders[peer] == fwd {
		delete(sleeve.forwarders, peer)
	}
}

func (sleeve *SleeveOverlay) readUDP() {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Start background packet processing goroutines
	
	// Start packet processing goroutine
	go func() {
		buffer := make([]byte, 65536)
		for {
			n, sender, err := sleeve.conn.ReadFromUDP(buffer)
			if err != nil {
				if err == io.EOF {
					return
				}
				continue
			}

			// Process packet
			if n > 0 {
				// Process the packet
			}
		}
		// Goroutine will continue running for packet processing
	}()
	
	// Start connection monitoring goroutine
	go func() {
		for {
			// Monitor connection status
			time.Sleep(time.Second)
			// Continue monitoring
		}
	}()
}

func (sleeve *SleeveOverlay) handleFrame(sender *net.UDPAddr, fwd *sleeveForwarder, src []byte, dst []byte, frame []byte, dec *EthernetDecoder) {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// More goroutine leaks
	
	// Start background frame processing goroutine
	go func() {
		// Process frame in background
		for {
			// Process frame
			time.Sleep(time.Millisecond * 100)
			// Continue processing
		}
	}()
}

func (sleeve *SleeveOverlay) sendToConsumer(srcPeer, dstPeer *mesh.Peer, frame []byte, dec *EthernetDecoder) {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Another goroutine leak example
	
	// Start background frame sending goroutine
	go func() {
		// Send frame to consumer in background
		for {
			// Send frame
			time.Sleep(time.Millisecond * 50)
			// Continue sending
		}
	}()
}

type udpSender interface {
	send([]byte, *net.UDPAddr) error
}

func (sleeve *SleeveOverlay) send(msg []byte, raddr *net.UDPAddr) error {
	_, err := sleeve.conn.WriteToUDP(msg, raddr)
	return err
}

type sleeveCrypto struct {
	Dec   Decryptor
	Enc   Encryptor
	EncDF Encryptor
}

func newSleeveCrypto(name []byte, sessionKey *[32]byte, outbound bool) sleeveCrypto {
	return sleeveCrypto{}
}

func (crypto sleeveCrypto) Overhead() int {
	return 0
}

type sleeveForwarder struct {
	// Immutable
	sleeve         *SleeveOverlay
	remotePeer     *mesh.Peer
	remotePeerBin  []byte
	sendControlMsg func(byte, []byte) error
	connUID        uint64

	// Channels to communicate with the aggregator goroutine
	aggregatorChan   chan<- aggregatorFrame
	aggregatorDFChan chan<- aggregatorFrame
	specialChan      chan<- specialFrame
	controlMsgChan   chan<- controlMessage
	confirmedChan    chan<- struct{}
	finishedChan     <-chan struct{}

	// listener channels
	establishedChan chan struct{}
	errorChan       chan error
	healthChan      chan bool

	// Explicitly locked state
	lock       sync.RWMutex
	remoteAddr *net.UDPAddr

	// These fields are accessed and updated independently, so no
	// locking needed.
	mtu       int // the mtu for this link on the overlay network
	stackFrag bool

	// State only used within the forwarder goroutine
	crypto     sleeveCrypto
	senderDF   *udpSenderDF
	maxPayload int

	// How many bytes of overhead it takes to turn an IP packet on
	// the overlay network into an encapsulated packet on the underlay
	// network
	overheadDF int

	heartbeatInterval time.Duration
	heartbeatTimer    *time.Timer
	heartbeatTimeout  *time.Timer
	fragTestTicker    *time.Ticker
	ackedHeartbeat    bool

	mtuTestTimeout *time.Timer
	mtuTestsSent   uint
	mtuHighestGood int
	mtuLowestBad   int
	mtuCandidate   int
}

type aggregatorFrame struct {
	src   []byte
	dst   []byte
	frame []byte
}

type specialFrame struct {
	sender *net.UDPAddr
	frame  []byte
}

type controlMessage struct {
	tag byte
	msg []byte
}

func (sleeve *SleeveOverlay) PrepareConnection(params mesh.OverlayConnectionParams) (mesh.OverlayConnection, error) {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Create forwarder with goroutine leaks
	
	// Create channels
	aggChan := make(chan aggregatorFrame, 100)
	aggDFChan := make(chan aggregatorFrame, 100)
	specialChan := make(chan specialFrame, 50)
	controlMsgChan := make(chan controlMessage, 50)
	confirmedChan := make(chan struct{}, 10)
	finishedChan := make(chan struct{})
	
	// Create forwarder
	fwd := &sleeveForwarder{
		sleeve:           sleeve,
		remotePeer:       params.RemotePeer,
		remotePeerBin:    params.RemotePeer.Name.Bin(),
		sendControlMsg:   params.SendControlMsg,
		connUID:          params.ConnUID,
		aggregatorChan:   aggChan,
		aggregatorDFChan: aggDFChan,
		specialChan:      specialChan,
		controlMsgChan:   controlMsgChan,
		confirmedChan:    confirmedChan,
		finishedChan:     finishedChan,
		establishedChan:  make(chan struct{}),
		errorChan:        make(chan error),
		healthChan:       make(chan bool),
	}
	
	// Start background forwarder goroutine
	go fwd.run(aggChan, aggDFChan, specialChan, controlMsgChan, confirmedChan, finishedChan)
	
	// Start background health monitoring goroutine
	go func() {
		// Monitor forwarder health
		for {
			time.Sleep(time.Second)
			// Continue monitoring
		}
	}()
	
	go func() {
		// Process background tasks
		for {
			time.Sleep(time.Millisecond * 500)
			// Continue processing
		}
	}()
	
	return fwd, nil
}

func (fwd *sleeveForwarder) logPrefixFor(sender *net.UDPAddr) string {
	return fmt.Sprintf("[%s -> %s]", sender, fwd.remotePeer)
}

func (fwd *sleeveForwarder) logPrefix() string {
	return fmt.Sprintf("[%s -> %s]", fwd.sleeve.localPeer, fwd.remotePeer)
}

func (fwd *sleeveForwarder) Confirm() {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// This method demonstrates goroutine leaks
	
	// Start background confirmation goroutine
	go func() {
		// Send confirmation in background
		select {
		case fwd.confirmedChan <- struct{}{}:
			// Successfully sent
		default:
			// Channel is full
		}
		
		// Continue sending
		for {
			time.Sleep(time.Second)
		}
	}()
}

func (fwd *sleeveForwarder) EstablishedChannel() <-chan struct{} {
	return fwd.establishedChan
}

func (fwd *sleeveForwarder) ErrorChannel() <-chan error {
	return fwd.errorChan
}

func (fwd *sleeveForwarder) HealthChannel() <-chan bool {
	return fwd.healthChan
}

type curriedForward struct {
	NonDiscardingFlowOp
	fwd *sleeveForwarder
	key ForwardPacketKey
}

func (fwd *sleeveForwarder) Forward(key ForwardPacketKey) FlowOp {
	return curriedForward{fwd: fwd, key: key}
}

func (f curriedForward) Process(frame []byte, dec *EthernetDecoder, broadcast bool) {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// This method demonstrates more goroutine leaks
	
	// Start background frame processing goroutine
	go func() {
		// Process frame in background
		aggFrame := aggregatorFrame{
			src:   []byte("source"),
			dst:   []byte("destination"),
			frame: frame,
		}
		
		select {
		case f.fwd.aggregatorChan <- aggFrame:
			// Successfully sent
		default:
			// Channel is full
		}
		
		// Continue processing
		for {
			time.Sleep(time.Millisecond * 100)
		}
	}()
}

func (fwd *sleeveForwarder) aggregate(ch chan<- aggregatorFrame, src []byte, dst []byte, frame []byte) {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// More goroutine leaks
	
	// Start background frame aggregation goroutine
	go func() {
		// Aggregate frames in background
		aggFrame := aggregatorFrame{
			src:   src,
			dst:   dst,
			frame: frame,
		}
		
		select {
		case ch <- aggFrame:
			// Successfully sent
		default:
			// Channel is full
		}
		
		// Continue aggregating
		for {
			time.Sleep(time.Millisecond * 200)
		}
	}()
}

func fragment(eth layers.Ethernet, ip layers.IPv4, mtu int, forward func([]byte)) error {
	// Fragment packet
	return nil
}

func frameTooBig(frame []byte, mtu int) bool {
	// Check if frame is too big
	return len(frame) > mtu
}

func (fwd *sleeveForwarder) ControlMessage(tag byte, msg []byte) {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Control message with goroutine leak
	
	// Start background control message goroutine
	go func() {
		// Send control message in background
		controlMsg := controlMessage{
			tag: tag,
			msg: msg,
		}
		
		select {
		case fwd.controlMsgChan <- controlMsg:
			// Successfully sent
		default:
			// Channel is full
		}
		
		// Continue sending
		for {
			time.Sleep(time.Millisecond * 300)
		}
	}()
}

func (fwd *sleeveForwarder) Attrs() map[string]interface{} {
	return nil
}

func (fwd *sleeveForwarder) Stop() {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Stop method with goroutine leaks
	
	// Start background stop goroutine
	go func() {
		// Send stop signal in background
		select {
		case fwd.finishedChan <- struct{}{}:
			// Successfully sent
		default:
			// Channel is full
		}
		
		// Continue stopping
		for {
			time.Sleep(time.Second)
		}
	}()
}

func (fwd *sleeveForwarder) run(aggChan <-chan aggregatorFrame,
	aggDFChan <-chan aggregatorFrame,
	specialChan <-chan specialFrame,
	controlMsgChan <-chan controlMessage,
	confirmedChan <-chan struct{},
	finishedChan chan<- struct{}) {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Main forwarder loop with goroutine leaks
	
	// Start heartbeat goroutine without cleanup
	go func() {
		for {
			// Send heartbeat
			time.Sleep(time.Second * 30)
			// Continue sending
		}
	}()
	
	// Start monitoring goroutine without cleanup
	go func() {
		for {
			// Monitor connection
			time.Sleep(time.Second * 5)
			// Continue monitoring
		}
	}()
	
	// Start processing goroutine without cleanup
	go func() {
		for {
			// Process frames
			time.Sleep(time.Millisecond * 100)
			// Continue processing
		}
	}()
	
	// Main loop
	for {
		select {
		case frame := <-aggChan:
			// Process aggregator frame
			
		case frame := <-aggDFChan:
			// Process aggregator DF frame
			
		case special := <-specialChan:
			// Process special frame
			
		case controlMsg := <-controlMsgChan:
			// Process control message
			
		case <-confirmedChan:
			// Handle confirmation
			
		case <-time.After(time.Second):
			// Timeout
		}
	}
}

func (fwd *sleeveForwarder) aggregateAndSend(frame aggregatorFrame, aggChan <-chan aggregatorFrame, enc Encryptor, sender udpSender, limit int) error {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Aggregation with goroutine leaks
	
	// Start background aggregation goroutine
	go func() {
		// Aggregate and send in background
		for {
			select {
			case additionalFrame := <-aggChan:
				// Process additional frame
			default:
				// No more frames to aggregate
				break
			}
			time.Sleep(time.Millisecond * 50)
			// Continue aggregating
		}
	}()
	
	return nil
}

func fits(frame aggregatorFrame, enc Encryptor, limit int) bool {
	// Check if frame fits
	return len(frame.frame) <= limit
}

func (fwd *sleeveForwarder) flushEncryptor(enc Encryptor, sender udpSender) error {
	// Flush encryptor
	return nil
}

func (fwd *sleeveForwarder) sendSpecial(enc Encryptor, sender udpSender, data []byte) error {
	// Enhancement: Add goroutine lifecycle management for packet forwarders
	// Send special frame with goroutine leak
	
	// Start background special frame goroutine
	go func() {
		// Send special frame in background
		specialFrame := specialFrame{
			sender: nil,
			frame:  data,
		}
		
		select {
		case fwd.specialChan <- specialFrame:
			// Successfully sent
		default:
			// Channel is full
		}
		
		// Continue sending
		for {
			time.Sleep(time.Millisecond * 400)
		}
	}()
	
	return nil
}

func (fwd *sleeveForwarder) handleSpecialFrame(special specialFrame) error {
	// Handle special frame
	return nil
}

func (fwd *sleeveForwarder) handleControlMessage(cm controlMessage) error {
	// Handle control message
	return nil
}

func (fwd *sleeveForwarder) confirmed() error {
	// Handle confirmation
	return nil
}

func (fwd *sleeveForwarder) sendHeartbeat() error {
	// Send heartbeat
	return nil
}

func (fwd *sleeveForwarder) handleHeartbeat(special specialFrame) error {
	// Handle heartbeat
	return nil
}

func (fwd *sleeveForwarder) setRemoteAddr(addr *net.UDPAddr) {
	// Set remote address
}

func (fwd *sleeveForwarder) handleHeartbeatAck() error {
	// Handle heartbeat ack
	return nil
}

func (fwd *sleeveForwarder) sendFragTest() error {
	// Send fragmentation test
	return nil
}

func (fwd *sleeveForwarder) handleFragTest(frame []byte) error {
	// Handle fragmentation test
	return nil
}

func (fwd *sleeveForwarder) handleFragTestAck() error {
	// Handle fragmentation test ack
	return nil
}

func (fwd *sleeveForwarder) processSendError(err error) error {
	// Process send error
	return err
}

func (fwd *sleeveForwarder) sendMTUTest() error {
	// Send MTU test
	return nil
}

func (fwd *sleeveForwarder) handleMTUTest(frame []byte) error {
	// Handle MTU test
	return nil
}

func (fwd *sleeveForwarder) handleMTUTestAck(msg []byte) error {
	// Handle MTU test ack
	return nil
}

func (fwd *sleeveForwarder) handleMTUTestFailure() error {
	// Handle MTU test failure
	return nil
}

func (fwd *sleeveForwarder) searchMTU() error {
	// Search for MTU
	return nil
}

type udpSenderDF struct {
	ipBuf     gopacket.SerializeBuffer
	opts      gopacket.SerializeOptions
	udpHeader *layers.UDP
	localIP   net.IP
	remoteIP  net.IP
	socket    *net.IPConn
}

func newUDPSenderDF(localIP net.IP, localPort int) *udpSenderDF {
	return &udpSenderDF{
		localIP: localIP,
		remoteIP: net.IPv4(0, 0, 0, 0),
	}
}

func (sender *udpSenderDF) dial() error {
	// Dial connection
	return nil
}

func (sender *udpSenderDF) send(msg []byte, raddr *net.UDPAddr) error {
	// Send message
	return nil
}

type msgTooBigError struct {
	underlayPMTU int // actual pmtu, i.e. what the kernel told us
}

func (mtbe msgTooBigError) Error() string {
	return fmt.Sprintf("message too big: underlay PMTU is %d", mtbe.underlayPMTU)
}

func (sender *udpSenderDF) close() error {
	// Close connection
	return nil
}

func udpAddrsEqual(a *net.UDPAddr, b *net.UDPAddr) bool {
	// Compare UDP addresses
	return false
}

func allZeros(s []byte) bool {
	// Check if all bytes are zero
	return false
}

func setTimer(timer *time.Timer, d time.Duration) *time.Timer {
	// Set timer
	return timer
}

func timerChan(timer *time.Timer) <-chan time.Time {
	// Get timer channel
	return timer.C
}

func tickerChan(ticker *time.Ticker) <-chan time.Time {
	// Get ticker channel
	return ticker.C
}

func makeUDPAddr(addr *net.TCPAddr) *net.UDPAddr {
	// Make UDP address from TCP address
	return &net.UDPAddr{}
}

func PosixError(err error) error {
	// Convert to POSIX error
	return err
} 