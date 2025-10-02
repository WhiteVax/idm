package common

import "net"

type Addr string

type CustomListener struct {
	net.Listener
	Url string
}

func (f Addr) Network() string { return "tcp" }
func (f Addr) String() string  { return string(f) }

func (c CustomListener) Addr() net.Addr {
	return Addr(c.Url)
}
