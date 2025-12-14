package znet

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func ClientTest() {
	fmt.Println("Client Test")
	time.Sleep(3 * time.Second)
	 conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8888,
	})
	if err != nil {
		fmt.Printf("[ERROR]DialTCP failed, err: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		_, err := conn.Write([]byte("hello world"))
		if err != nil {
			fmt.Printf("[ERROR]Write failed, err: %v\n", err)
			continue
		}
		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("[ERROR]Read failed, err: %v\n", err)
			continue
		}
		fmt.Printf("[RECV]%s\n", buf[:n])
	}
}

func TestServer(t *testing.T) {
	s := NewServer("test")
	go ClientTest()
	s.Serve()
}