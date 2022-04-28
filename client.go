package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前客户端的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net dial err:", err)
		return nil
	}
	client.conn = conn
	return client
}

func (c *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

func (c *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err: ", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (c *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err: ", err)
		return
	}
}
func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	c.SelectUsers()
	fmt.Println(">>>>请输入聊天对象的[用户名], exit退出: ")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容，exit退出:")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err: ", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>请输入消息内容，exit退出:")
			fmt.Scanln(&chatMsg)
		}
		c.SelectUsers()
		fmt.Println(">>>>请输入聊天对象的[用户名], exit退出: ")
		fmt.Scanln(&remoteName)

	}
}

func (c *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名：")
	fmt.Scanln(&c.Name)
	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (c *Client) DealResponse() {
	// 一旦c.conn有数据，直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) run() {
	for c.flag != 0 {
		for !c.menu() {
		}
		switch c.flag {
		case 1:
			// 公聊模式
			c.PublicChat()
		case 2:
			// 私聊模式
			c.PrivateChat()
		case 3:
			//更新用户名
			c.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>连接服务器失败")
		return
	}
	fmt.Println(">>>>>连接服务器成功")
	// 单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()
	client.run()
}
