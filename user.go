package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	Coon   net.Conn
	Server *Server
}

func NewUser(coon net.Conn, server *Server) *User {
	userAddr := coon.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		Coon:   coon,
		Server: server,
	}
	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}

//用户的上线业务
func (u *User) Online() {
	u.Server.MapLock.Lock()
	u.Server.OnlineMap[u.Name] = u
	u.Server.MapLock.Unlock()
	u.Server.BroadCast(u, "已上线")
}

//用户的下线业务
func (u *User) Offline() {
	u.Server.MapLock.Lock()
	delete(u.Server.OnlineMap, u.Name)
	u.Server.MapLock.Unlock()
	u.Server.BroadCast(u, "下线")
}

func (u *User) SendMsg(msg string) {
	u.Coon.Write([]byte(msg))
}

//用户处理消息的业务
func (u *User) DoMessage(msg string) {
	if msg == "who" {
		u.Server.MapLock.Lock()
		for _, user := range u.Server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线...\n"
			u.SendMsg(onlineMsg)
		}
		u.Server.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := u.Server.OnlineMap[newName]
		if ok {
			u.SendMsg("当前用户名被使用\n")
		} else {
			u.Server.MapLock.Lock()
			delete(u.Server.OnlineMap, u.Name)
			u.Server.OnlineMap[newName] = u
			u.Server.MapLock.Unlock()
			u.Name = newName
			u.SendMsg("您已更新用户名:" + u.Name + "\n")
		}
	} else if len(msg) > 3 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式不正确，请使用\"to|张三|你好啊\"格式。\n")
			return
		}
		remoteUser, ok := u.Server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("该用户名不存在\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(u.Name + "对您说" + content)
	} else {
		u.Server.BroadCast(u, msg)
	}
}

// 监听当前user channel的方法，一旦有消息，直接发送给客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.Coon.Write([]byte(msg + "\n"))
	}
}
