// Package net provides TCP and UDP networking for SpaceOS, modeled on
// Go's stdlib net package. It offers Dial, Listen, and PacketConn
// backed by the kernel's lwIP stack.
//
// The API intentionally mirrors the net package shapes users already
// know, so a program can roughly read like:
//
//	c, err := spnet.Dial("tcp", "10.0.2.2:80")
//	if err != nil { ... }
//	defer c.Close()
//	io.WriteString(c, "GET / HTTP/1.0\r\n\r\n")
//	io.Copy(os.Stdout, c)
//
// Not everything stdlib net does is supported — no IPv6, no DNS (the
// address must be "dotted-quad:port"), no SetDeadline on TCP. If you
// need DNS, use spaceos/net/dns (TODO) or pre-resolve yourself.
package net

import (
	"errors"
	"strconv"
	"unsafe"
)

// ---------------------------------------------------------------- errors

var (
	ErrClosed  = errors.New("spaceos/net: connection closed")
	ErrTimeout = errors.New("spaceos/net: i/o timeout")
	ErrBadAddr = errors.New("spaceos/net: bad address")
)

// ---------------------------------------------------------------- addresses

type Addr struct {
	IP   [4]byte
	Port uint16
}

func (a Addr) String() string {
	return itoa(int(a.IP[0])) + "." + itoa(int(a.IP[1])) + "." +
		itoa(int(a.IP[2])) + "." + itoa(int(a.IP[3])) + ":" + itoa(int(a.Port))
}

func parseAddr(s string) (Addr, error) {
	// dotted-quad:port
	var a Addr
	colon := -1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ':' {
			colon = i
			break
		}
	}
	if colon < 0 {
		return a, ErrBadAddr
	}
	ip := s[:colon]
	port, err := strconv.Atoi(s[colon+1:])
	if err != nil || port < 0 || port > 0xffff {
		return a, ErrBadAddr
	}
	a.Port = uint16(port)

	// Parse dotted quad.
	part := 0
	val := 0
	haveDigit := false
	for i := 0; i <= len(ip); i++ {
		if i == len(ip) || ip[i] == '.' {
			if !haveDigit || val > 255 || part >= 4 {
				return a, ErrBadAddr
			}
			a.IP[part] = byte(val)
			part++
			val = 0
			haveDigit = false
			continue
		}
		c := ip[i]
		if c < '0' || c > '9' {
			return a, ErrBadAddr
		}
		val = val*10 + int(c-'0')
		haveDigit = true
	}
	if part != 4 {
		return a, ErrBadAddr
	}
	return a, nil
}

func ipU32(a Addr) uint32 {
	return uint32(a.IP[0])<<24 | uint32(a.IP[1])<<16 |
		uint32(a.IP[2])<<8 | uint32(a.IP[3])
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// ---------------------------------------------------------------- tcp

// Kernel-side functions (tinygo_net.c).

//export spaceos_tcp_dial
func cTcpDial(ip uint32, port uint16, timeoutMs int32) uintptr

//export spaceos_tcp_read
func cTcpRead(conn uintptr, buf *byte, cap uintptr, timeoutMs int32) int32

//export spaceos_tcp_write
func cTcpWrite(conn uintptr, buf *byte, n uintptr, timeoutMs int32) int32

//export spaceos_tcp_close
func cTcpClose(conn uintptr)

//export spaceos_tcp_listen
func cTcpListen(port uint16) uintptr

//export spaceos_tcp_accept
func cTcpAccept(lst uintptr, timeoutMs int32) uintptr

//export spaceos_tcp_listener_close
func cTcpListenerClose(lst uintptr)

// Conn is the io.ReadWriteCloser for a TCP connection. It implements
// the common subset of net.Conn.
type Conn struct {
	h      uintptr
	remote Addr
	local  Addr
}

func (c *Conn) Read(p []byte) (int, error) {
	if c.h == 0 {
		return 0, ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := cTcpRead(c.h, &p[0], uintptr(len(p)), -1)
	if n < 0 {
		return 0, ErrClosed
	}
	if n == 0 {
		return 0, ErrClosed // EOF on peer close
	}
	return int(n), nil
}

func (c *Conn) Write(p []byte) (int, error) {
	if c.h == 0 {
		return 0, ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := cTcpWrite(c.h, &p[0], uintptr(len(p)), -1)
	if n < 0 {
		return 0, ErrClosed
	}
	return int(n), nil
}

func (c *Conn) Close() error {
	if c.h == 0 {
		return nil
	}
	h := c.h
	c.h = 0
	cTcpClose(h)
	return nil
}

func (c *Conn) RemoteAddr() Addr { return c.remote }
func (c *Conn) LocalAddr() Addr  { return c.local }

// Dial opens a TCP connection. Only network=="tcp" is supported.
func Dial(network, address string) (*Conn, error) {
	if network != "tcp" {
		return nil, errors.New("spaceos/net: unsupported network " + network)
	}
	a, err := parseAddr(address)
	if err != nil {
		return nil, err
	}
	h := cTcpDial(ipU32(a), a.Port, 10_000)
	if h == 0 {
		return nil, errors.New("spaceos/net: dial failed")
	}
	return &Conn{h: h, remote: a}, nil
}

// Listener accepts inbound TCP connections.
type Listener struct {
	h    uintptr
	port uint16
}

// Listen binds to port and starts accepting connections. Only
// network=="tcp" is supported; address must be ":port".
func Listen(network, address string) (*Listener, error) {
	if network != "tcp" {
		return nil, errors.New("spaceos/net: unsupported network " + network)
	}
	// Accept ":port" form.
	port := uint16(0)
	if len(address) >= 2 && address[0] == ':' {
		p, err := strconv.Atoi(address[1:])
		if err != nil || p <= 0 || p > 0xffff {
			return nil, ErrBadAddr
		}
		port = uint16(p)
	} else {
		a, err := parseAddr(address)
		if err != nil {
			return nil, err
		}
		port = a.Port
	}
	h := cTcpListen(port)
	if h == 0 {
		return nil, errors.New("spaceos/net: listen failed")
	}
	return &Listener{h: h, port: port}, nil
}

// Accept waits for a new connection.
func (l *Listener) Accept() (*Conn, error) {
	if l.h == 0 {
		return nil, ErrClosed
	}
	h := cTcpAccept(l.h, -1)
	if h == 0 {
		return nil, ErrClosed
	}
	return &Conn{h: h}, nil
}

func (l *Listener) Close() error {
	if l.h == 0 {
		return nil
	}
	h := l.h
	l.h = 0
	cTcpListenerClose(h)
	return nil
}

func (l *Listener) Port() uint16 { return l.port }

// ---------------------------------------------------------------- udp

//export spaceos_udp_open
func cUdpOpen(port uint16) uintptr

//export spaceos_udp_send_to
func cUdpSendTo(s uintptr, ip uint32, port uint16, buf *byte, n uint16) int32

//export spaceos_udp_recv_from
func cUdpRecvFrom(s uintptr, srcIP *uint32, srcPort *uint16, buf *byte, cap uint16, timeoutMs int32) int32

//export spaceos_udp_close
func cUdpClose(s uintptr)

// PacketConn is an unreliable datagram socket.
type PacketConn struct {
	h    uintptr
	port uint16
}

// ListenPacket opens a UDP socket. address is ":port" or "0.0.0.0:port".
// Pass ":0" to leave unbound (you can only send, not receive).
func ListenPacket(network, address string) (*PacketConn, error) {
	if network != "udp" {
		return nil, errors.New("spaceos/net: unsupported network " + network)
	}
	port := uint16(0)
	if len(address) >= 2 && address[0] == ':' {
		p, err := strconv.Atoi(address[1:])
		if err != nil || p < 0 || p > 0xffff {
			return nil, ErrBadAddr
		}
		port = uint16(p)
	} else {
		a, err := parseAddr(address)
		if err != nil {
			return nil, err
		}
		port = a.Port
	}
	h := cUdpOpen(port)
	if h == 0 {
		return nil, errors.New("spaceos/net: udp open failed")
	}
	return &PacketConn{h: h, port: port}, nil
}

// ReadFrom blocks until a datagram arrives. The returned Addr is the
// sender's address.
func (u *PacketConn) ReadFrom(p []byte) (int, Addr, error) {
	var addr Addr
	if u.h == 0 {
		return 0, addr, ErrClosed
	}
	if len(p) == 0 {
		return 0, addr, nil
	}
	var ip uint32
	var port uint16
	cap := len(p)
	if cap > 0xffff {
		cap = 0xffff
	}
	n := cUdpRecvFrom(u.h, &ip, &port, &p[0], uint16(cap), -1)
	if n < 0 {
		return 0, addr, ErrClosed
	}
	addr.IP[0] = byte(ip >> 24)
	addr.IP[1] = byte(ip >> 16)
	addr.IP[2] = byte(ip >> 8)
	addr.IP[3] = byte(ip)
	addr.Port = port
	return int(n), addr, nil
}

// WriteTo sends a datagram to dst.
func (u *PacketConn) WriteTo(p []byte, dst Addr) (int, error) {
	if u.h == 0 {
		return 0, ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	if len(p) > 0xffff {
		p = p[:0xffff]
	}
	n := cUdpSendTo(u.h, ipU32(dst), dst.Port, &p[0], uint16(len(p)))
	if n < 0 {
		return 0, errors.New("spaceos/net: udp send failed")
	}
	return int(n), nil
}

func (u *PacketConn) Close() error {
	if u.h == 0 {
		return nil
	}
	h := u.h
	u.h = 0
	cUdpClose(h)
	return nil
}

// ---------------------------------------------------------------- status

//export spaceos_net_status
func cNetStatus(out *cNetStatusRaw)

type cNetStatusRaw struct {
	IP           uint32
	Netmask      uint32
	Gateway      uint32
	Up           uint8
	DHCPBound    uint8
	_            uint16
	HTTPRequests uint64
}

// Status is a snapshot of the network interface state.
type Status struct {
	IP, Netmask, Gateway Addr
	Up, DHCPBound        bool
	HTTPRequests         uint64
}

func CurrentStatus() Status {
	var r cNetStatusRaw
	cNetStatus(&r)
	ipFrom := func(u uint32) Addr {
		return Addr{IP: [4]byte{byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}}
	}
	return Status{
		IP:           ipFrom(r.IP),
		Netmask:      ipFrom(r.Netmask),
		Gateway:      ipFrom(r.Gateway),
		Up:           r.Up != 0,
		DHCPBound:    r.DHCPBound != 0,
		HTTPRequests: r.HTTPRequests,
	}
}

// Forces the compiler to keep unsafe in scope even if unused elsewhere.
var _ = unsafe.Sizeof(cNetStatusRaw{})
