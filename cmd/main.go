package main

import (
	"flag"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go-net-server/internal/model"

	log "github.com/golang/glog"
)

func main() {
	flag.Parse()
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 8080,
	})
	if err != nil {
		log.Errorf("net.ListenTCP(tcp, %s) error(%v)", ":8080", err)
		return
	}
	go accpetTcp(listener)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func accpetTcp(listener *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
	)

	for {
		if conn, err = listener.AcceptTCP(); err != nil {
			log.Errorf("listener.Accept(\"%s\") error(%v)", listener.Addr().String(), err)
			return
		}
		go serveTCP(conn)
	}
}

func serveTCP(conn *net.TCPConn) {
	// 封装连接
	channel := &model.Channel{
		Conn:   conn,
		Signal: make(chan *model.Data, 10),
	}

	channel.Signal <- &model.Data{
		Content: []byte("connect success!"),
	}

	// 开启协程处理写事件
	go dispatchTCP(channel)

	// 这个协程处理读事件
	for {
		// 解码数据，封装成Data，进行业务逻辑，然后写入Signal返回resp
		bytes := readTCP(channel)
		log.Info("server receive info:%s", bytes)

		data := &model.Data{
			Content: []byte("server receive!"),
		}

		channel.Signal <- data
	}
}

func readTCP(channel *model.Channel) []byte {
	conn := channel.Conn

	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 256)
	i := 0
	for i < 10 {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				//Error Handler
			}
			break
		}
		//fmt.Println("got", n, "bytes.")
		buf = append(buf, tmp[:n]...)
		i++
	}
	return buf
}

func dispatchTCP(channel *model.Channel) {
	for {
		data := <-channel.Signal
		channel.Conn.Write(data.Content)
	}
}
