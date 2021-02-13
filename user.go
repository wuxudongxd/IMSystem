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
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}
	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
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
