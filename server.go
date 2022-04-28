package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User
	MapLock   sync.RWMutex
	// 消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		s.MapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.MapLock.Unlock()
	}
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)
	user.Online()
	isLive := make(chan bool)
	//接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read Err:", err)
				return
			}
			//提取用户的消息(去除'\n')
			msg := string(buf[:n-1])
			//用户针对msg进行消息处理
			user.DoMessage(msg)
			isLive <- true
		}
	}()
	// 当前handler阻塞
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 300):
			user.SendMsg("你被踢了")
			close(user.C)
			conn.Close()
			return
		}
	}

}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listener.Close()
	// 启动监听Message的goroutine
	go s.ListenMessager()
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go s.Handler(conn)
	}
}
