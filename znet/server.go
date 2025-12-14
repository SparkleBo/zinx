package znet

import (
	"fmt"
	"net"

	"github.com/SparkleBo/zinx/ziface"
)

type Server struct {
	Name string
	IPVersion string
	IP string
	Port int
}

func (s *Server) Start() {
	fmt.Printf("[START]Server Name: %s, IPVersion: %s, IP: %s, Port: %d\n", s.Name, s.IPVersion, s.IP, s.Port)
	go func() {
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Printf("[ERROR]ResolveTCPAddr failed, err: %v\n", err)
			return
		}
		// 监听 TCP 地址
		l, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Printf("[ERROR]ListenTCP failed, err: %v\n", err)
			return
		}
		defer l.Close()
		// 启动 server 网络连接业务
		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				fmt.Printf("[ERROR]AcceptTCP failed, err: %v\n", err)
				continue
			}
			// 针对每个 connection 都启动一个 goroutine
			go func() {
				for {
					buf := make([]byte, 512)
					n, err := conn.Read(buf)
					if err != nil {
						fmt.Printf("[ERROR]Read failed, err: %v\n", err)
						continue
					}
					fmt.Printf("[RECV]%s\n", buf[:n])
					// 回显
					if _, err := conn.Write(buf[:n]); err != nil {
						fmt.Printf("[ERROR]Write failed, err: %v\n", err)
						continue
					}
				}
			}()
		}
	}()
	println("Server Start")
}

func (s *Server) Stop() {
	fmt.Printf("[STOP]Server Name: %s, IPVersion: %s, IP: %s, Port: %d\n", s.Name, s.IPVersion, s.IP, s.Port)
}

func (s *Server) Serve() {
	s.Start()

	select{}
}


func NewServer(name string) ziface.IServer {
	return &Server{
		Name: name,
		IPVersion: "tcp4",
		IP: "127.0.0.1",
		Port: 8888,
	}
}
