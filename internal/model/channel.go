package model

import "net"

type Channel struct {
	Conn   *net.TCPConn
	Signal chan *Data
}
