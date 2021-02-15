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

	//	在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message channel 广播消息的goroutine，一旦有消息就发送给全部在线的user
func (server *Server) ListenServerMsg() {
	for {
		msg := <-server.Message
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 广播消息的方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}
func (server *Server) Handler(conn net.Conn) {
	// 当前的业务
	//fmt.Println("连接建立成功")
	user := NewUser(conn, server)

	// 上线
	user.Online()

	// 初始化用户活跃状态值
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				// 下线
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn Read err:", err)
				return
			}

			// 提取用户的消息（去除"\n"）
			msg := string(buf[:n-1])
			//fmt.Println("msg:", msg)

			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			// 用户发送消息，代表当前用户活跃
			isLive <- true
		}
	}()

	// 将当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事，为了激活select，更新下面的定时器

		case <-time.After(time.Second * 10):
			// 已经超时
			// 将当前的用户强制关闭

			// 发送消息：强制踢除
			user.SendMsg("你由于长时间未活跃，已被强制下线")
			// 下线
			//user.Offline()
			// 销毁用的资源
			close(user.C)
			// 关闭连接
			err := conn.Close()
			if err != nil {
				fmt.Println("conn close err:", err)
			}
			// 退出当前channel
			return
		}
	}

}

// 启动服务器的接口
func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Lister err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go server.ListenServerMsg()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go server.Handler(conn)
	}
}
