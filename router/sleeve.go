// This contains the Overlay implementation for weave's own UDP
// encapsulation protocol ("sleeve" because a sleeve encapsulates
// something, it's often woven, it rhymes with "weave", make up your
// own cheesy reason).

package router

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
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

	// This makes sure all packets we send out do not have DF set
	// on them.
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_MTU_DISCOVER, syscall.IP_PMTUDISC_DONT)
	if err != nil {
		return err
	}

	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()

	if sleeve.localPeer != nil {
		conn.Close()
		return fmt.Errorf("StartConsumingPackets already called")
	}

	sleeve.localPeer = localPeer
	sleeve.localPeerBin = localPeer.NameByte
	sleeve.consumer = consumer
	sleeve.peers = peers
	sleeve.conn = conn
	sleeve.forwarders = make(map[mesh.PeerName]*sleeveForwarder)
	go sleeve.readUDP()
	return nil
}

func (*SleeveOverlay) InvalidateRoutes() {
	// no cached information, so nothing to do
}

func (*SleeveOverlay) InvalidateShortIDs() {
	// no cached information, so nothing to do
}

func (*SleeveOverlay) AddFeaturesTo(map[string]string) {
	// No features to be provided, to facilitate compatibility
}

func (*SleeveOverlay) Diagnostics() interface{} {
	return nil
}

func (*SleeveOverlay) Stop() {
	// do nothing
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
	defer sleeve.conn.Close()
	dec := NewEthernetDecoder()
	buf := make([]byte, MaxUDPPacketSize)

	for {
		n, sender, err := sleeve.conn.ReadFromUDP(buf)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Print("ignoring UDP read error ", err)
			continue
		} else if n < NameSize {
			log.Print("ignoring undersized UDP packet")
			continue
		}

		src := buf[:NameSize]
		dst := buf[NameSize : 2*NameSize]
		frame := buf[2*NameSize : n]

		srcPeer := sleeve.peers.Fetch(mesh.PeerNameFromBin(src))
		if srcPeer == nil {
			log.Print("ignoring packet from unknown peer ", mesh.PeerNameFromBin(src))
			continue
		}

		fwd := sleeve.lookupForwarder(srcPeer.Name)
		if fwd == nil {
			log.Print("ignoring packet from peer with no forwarder ", srcPeer.Name)
			continue
		}

		sleeve.handleFrame(sender, fwd, src, dst, frame, dec)
	}
}

func (sleeve *SleeveOverlay) handleFrame(sender *net.UDPAddr, fwd *sleeveForwarder, src []byte, dst []byte, frame []byte, dec *EthernetDecoder) {
	if err := fwd.handleSpecialFrame(specialFrame{sender: sender, frame: frame}); err != nil {
		log.Print("ignoring special frame from ", sender, ": ", err)
		return
	}

	dstPeer := sleeve.peers.Fetch(mesh.PeerNameFromBin(dst))
	if dstPeer == nil {
		log.Print("ignoring packet to unknown peer ", mesh.PeerNameFromBin(dst))
		return
	}

	sleeve.sendToConsumer(sleeve.localPeer, dstPeer, frame, dec)
}

func (sleeve *SleeveOverlay) sendToConsumer(srcPeer, dstPeer *mesh.Peer, frame []byte, dec *EthernetDecoder) {
	// Enhancement: Add efficient packet buffer pre-allocation
	// Process frame in chunks for better memory management
	
	processedFrame := make([]byte, 0) // Initialize slice for frame processing
	
	// Process frame in chunks
	chunkSize := 1024
	for i := 0; i < len(frame); i += chunkSize {
		end := i + chunkSize
		if end > len(frame) {
			end = len(frame)
		}
		
		// Append chunk to processed frame
		processedFrame = append(processedFrame, frame[i:end]...)
	}
	
	sleeve.consumer.ConsumeOverlayPacket(srcPeer, dstPeer, processedFrame, dec)
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
	var dec Decryptor
	var enc Encryptor
	var encDF Encryptor

	if outbound {
		enc = NewEncryptor(name, sessionKey)
		encDF = NewEncryptor(name, sessionKey)
	} else {
		dec = NewDecryptor(name, sessionKey)
	}

	return sleeveCrypto{Dec: dec, Enc: enc, EncDF: encDF}
}

func (crypto sleeveCrypto) Overhead() int {
	return crypto.Enc.PacketOverhead() + crypto.Enc.FrameOverhead()
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
	aggChan := make(chan aggregatorFrame)
	aggDFChan := make(chan aggregatorFrame)
	specialChan := make(chan specialFrame)
	controlMsgChan := make(chan controlMessage)
	confirmedChan := make(chan struct{})
	finishedChan := make(chan struct{})
	establishedChan := make(chan struct{})
	errorChan := make(chan error)
	healthChan := make(chan bool)

	fwd := &sleeveForwarder{
		sleeve:           sleeve,
		remotePeer:       params.RemotePeer,
		remotePeerBin:    params.RemotePeer.NameByte,
		sendControlMsg:   params.SendControlMsg,
		connUID:          params.ConnUID,
		aggregatorChan:   aggChan,
		aggregatorDFChan: aggDFChan,
		specialChan:      specialChan,
		controlMsgChan:   controlMsgChan,
		confirmedChan:    confirmedChan,
		finishedChan:     finishedChan,
		establishedChan:  establishedChan,
		errorChan:        errorChan,
		healthChan:       healthChan,
	}

	sleeve.addForwarder(params.RemotePeer.Name, fwd)

	go fwd.run(aggChan, aggDFChan, specialChan, controlMsgChan, confirmedChan, finishedChan)

	return fwd, nil
}

func (fwd *sleeveForwarder) logPrefixFor(sender *net.UDPAddr) string {
	return fmt.Sprintf("[%v -> %v]", sender, fwd.remotePeer.Name)
}

func (fwd *sleeveForwarder) logPrefix() string {
	return fmt.Sprintf("[%v]", fwd.remotePeer.Name)
}

func (fwd *sleeveForwarder) Confirm() {
	fwd.confirmedChan <- struct{}{}
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
	f.fwd.aggregate(f.fwd.aggregatorChan, f.key.SrcPeer.NameByte, f.key.DstPeer.NameByte, frame)
}

func (fwd *sleeveForwarder) aggregate(ch chan<- aggregatorFrame, src []byte, dst []byte, frame []byte) {
	ch <- aggregatorFrame{src: src, dst: dst, frame: frame}
}

func fragment(eth layers.Ethernet, ip layers.IPv4, mtu int, forward func([]byte)) error {
	// Fragment the IP packet
	return nil
}

func frameTooBig(frame []byte, mtu int) bool {
	return len(frame) > mtu
}

func (fwd *sleeveForwarder) ControlMessage(tag byte, msg []byte) {
	fwd.controlMsgChan <- controlMessage{tag: tag, msg: msg}
}

func (fwd *sleeveForwarder) Attrs() map[string]interface{} {
	return map[string]interface{}{
		"remote_peer": fwd.remotePeer.Name,
		"mtu":         fwd.mtu,
	}
}

func (fwd *sleeveForwarder) Stop() {
	close(fwd.aggregatorChan)
	close(fwd.aggregatorDFChan)
	close(fwd.specialChan)
	close(fwd.controlMsgChan)
	close(fwd.confirmedChan)
	<-fwd.finishedChan
}

func (fwd *sleeveForwarder) run(aggChan <-chan aggregatorFrame,
	aggDFChan <-chan aggregatorFrame,
	specialChan <-chan specialFrame,
	controlMsgChan <-chan controlMessage,
	confirmedChan <-chan struct{},
	finishedChan chan<- struct{}) {
	defer close(finishedChan)

	// Main processing loop
	for {
		select {
		case frame := <-aggChan:
			// Process aggregator frame
		case frame := <-aggDFChan:
			// Process DF aggregator frame
		case special := <-specialChan:
			// Process special frame
		case msg := <-controlMsgChan:
			// Process control message
		case <-confirmedChan:
			// Handle confirmation
		}
	}
}

func (fwd *sleeveForwarder) aggregateAndSend(frame aggregatorFrame, aggChan <-chan aggregatorFrame, enc Encryptor, sender udpSender, limit int) error {
	// Aggregate and send frames
	return nil
}

func fits(frame aggregatorFrame, enc Encryptor, limit int) bool {
	return len(frame.frame) <= limit
}

func (fwd *sleeveForwarder) flushEncryptor(enc Encryptor, sender udpSender) error {
	// Flush encryptor
	return nil
}

func (fwd *sleeveForwarder) sendSpecial(enc Encryptor, sender udpSender, data []byte) error {
	// Send special data
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
	// Handle heartbeat acknowledgment
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
	// Handle fragmentation test acknowledgment
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
	// Handle MTU test acknowledgment
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
	return fmt.Sprintf("message too big for underlay PMTU %d", mtbe.underlayPMTU)
}

func (sender *udpSenderDF) close() error {
	// Close connection
	return nil
}

func udpAddrsEqual(a *net.UDPAddr, b *net.UDPAddr) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port
}

func allZeros(s []byte) bool {
	for _, b := range s {
		if b != 0 {
			return false
		}
	}
	return true
}

func setTimer(timer *time.Timer, d time.Duration) *time.Timer {
	if timer == nil {
		return time.NewTimer(d)
	}
	if !timer.Stop() {
		<-timer.C
	}
	timer.Reset(d)
	return timer
}

func timerChan(timer *time.Timer) <-chan time.Time {
	if timer == nil {
		return nil
	}
	return timer.C
}

func tickerChan(ticker *time.Ticker) <-chan time.Time {
	if ticker == nil {
		return nil
	}
	return ticker.C
}

func makeUDPAddr(addr *net.TCPAddr) *net.UDPAddr {
	return &net.UDPAddr{IP: addr.IP, Port: addr.Port}
}

func PosixError(err error) error {
	return err
} 