package main

import ("log"; "net"; "os"; "bufio"; "fmt"; "strings"; "time"; "strconv")
//A client struct that contains connection information and channel info
type Client struct {
    conn net.Conn
    ch chan<- string
}

type User struct {
    Nick string
    User string
    Real string
    IP string
    Host string
}


func main() {
    //Accept incoming tcp connection on port 3005
    ln, err := net.Listen("tcp", ":3005")
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }
    //Creates a channel for communication (similar to irc)
    msgChan := make(chan string)
    addChan := make(chan Client)
    //unregister disconnected clients
    rmChan := make(chan net.Conn)
    go messageHandler(msgChan, addChan, rmChan)
    for {
        //accepts connections unless error, and then print out the error if that exists
        conn, err := ln.Accept()
            if err != nil {
                log.Println(err)
                    continue
            }

        //This is really cool, its a goroutine which will exec concurrently w/ other goroutines in the same address space
        go connectionManagement(conn, msgChan, addChan, rmChan)
    }
}
//Manages the connection when a goroutine gets called on

func connectionManagement(c net.Conn, msgChan chan<- string, addChan chan<- Client, rmChan chan<- net.Conn) {
    ch := make(chan string)
    msgs := make(chan string)
    addChan <- Client{c, ch}
    go func() {
        defer close(msgs)
        bufc := bufio.NewReader(c)

        c.Write([]byte("Welcome to this chatroom!\n What is your name?"))
        name, _, err := bufc.ReadLine()

        if err != nil {
            return
        }
        nick := string(name)
        msgs <- "New user " + nick + " has joined the room"
        //This is the chat part, just a map of displaying readline
        for {
            output, _, err := bufc.ReadLine()
            if err != nil {
                break
            } else {
                splitStr := strings.Split(string(output), " ")
                switch splitStr[0] {
                    case "/QUIT":
                        msgs <- "User " + nick + "has left the room. Reason:" + strings.Replace(splitStr[1], "\r\n", "", 1)
                        c.Close()
                    case "/PING":
                        msgs <- "PONG " + strings.Replace(splitStr[1], "\r\n", "", 1)
                    case "/PONG":
                        msgs <- "PONG " + strings.Replace(splitStr[1], "\r\n", "", 1) 
                    case "/NICK":
                        nick = strings.Replace(splitStr[1], "\r\n", "", 1)
                    case "/TIME":
                        //Need to convert out of unix time
                        msgs <- strconv.FormatInt(time.Now().Unix(), 10)
                    default:
                        msgs <- nick + ": " + string(output)
                }
            }

        }

        msgs <- "User " + nick + " has left the room"
    }()
LOOP:
    for {
            select {
            case msg, ok := <-msgs:
                if !ok {
                    break LOOP
                }
                msgChan <- msg
            case msg := <-ch:
                _, err := c.Write([]byte(msg))
                if err != nil {
                    break LOOP
                }
            }
    }
    c.Close()
    log.Printf("Connection from %v closed.", c.RemoteAddr())
    rmChan <- c
}

func messageHandler(msgChan <-chan string, addChan <-chan Client, rmChan <-chan net.Conn) {
    clients := make(map[net.Conn]chan<- string)
    //new clients are received and added to a map of active connections
    for {
        select {
        case msg := <-msgChan:
            fmt.Printf("new message: %s\n", msg)
            for _, ch := range clients {
                go func(mch chan<- string) { 
                    mch <- "\033[1;32;40m" + msg + "\033[m\r\n" } (ch)
            }
        case client := <-addChan:
            fmt.Printf("New Client: %v \n", client.conn)
            clients[client.conn] = client.ch
        case conn := <-rmChan:
            fmt.Printf("Client left: %v \n", conn)
            delete(clients, conn)
        }
    }
}