package irc

import ("log"; "net"; "time")

type ClientMsg struct {
	client *Client
	message Message
}

type Server struct {
	Address string
	Name 	string
	Recieve	chan<- *ClientMsg
	Nick	map[string]*Client
	Chan 	map[string]*Chan
}

