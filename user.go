package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户的上线功能
func (user *User) Online() {
	// 用户上线，将用户加入的到OnlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	// 广播当前用户上线消息
	user.server.BroadCast(user, "已上线")
}

// 用户的下线功能
func (user *User) Offline() {
	// 用户下线，将用户从OnlineMap中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	//	广播当前用户上线消息
	user.server.BroadCast(user, "已下线")
}

// 给当前用户的客户端发消息
func (user *User) SendMsg(msg string) {
	_, err := user.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("user conn write err:", err)
	}
}

// 用户处理消息的业务
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		user.server.mapLock.Lock()
		for _, user := range user.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else {
		user.server.BroadCast(user, msg)
	}

}

// 监听当前user channel的方法，一旦channel有值，就直接发送给对应客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		_, err := user.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("conn write err:", err)
		}
	}
}
