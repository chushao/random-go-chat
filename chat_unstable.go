package main

/**
    TODO: USERMODE for users
          HOSTNAME GOES TO IP RIGHT NOW. need fixing
          Channels for different channels on one server
          Ways to save the chat logs?
          Ban list
**/


import ("log"; "net"; "bufio"; "fmt"; "container/list"; "strings")

//A client struct that contains
//The different buffers for input output
//Nickname - a display name
//User - the username, different from displayname as a username doesn't change
//Real - the realname
//Host - hostname
//IP - ip address
// TODO: USERMODE
type Client struct {
    In      *bufio.Reader
    Out     *bufio.Writer
    Nick    string
    User    string
    Real    string
    IP      string
    Host    string
    Conn    net.Conn
    ch chan<- string

}


func main() {
    //Create alist of clients that are going to be attached to this server
    clientList := list.New()
    //Accept incoming tcp connection on port 3005

    ln, _ := net.Listen("tcp", ":3005")
    //A for loop for each new user
    msgChan := make(chan string)
    addChan := make(chan Client)
    rmChan := make(chan net.Conn)
    go messageHandler(msgChan, addChan, rmChan)
    for {
        c, _ := ln.Accept()
        go connectionManagement(c, clientList, msgChan, addChan, rmChan)
        }
        
}
//Manages the connection when a goroutine gets called on

func connectionManagement(c net.Conn, clientList *list.List, msgChan chan<- string, addChan chan<- Client, rmChan chan<- net.Conn) {
    //Create a new reader and new writer for each client
    reader := bufio.NewReader(c)
    writer := bufio.NewWriter(c)
    //Ip Address
    ip := c.RemoteAddr().String()
    ch := make(chan string)
    msgs := make(chan string)
    //Initializes a client with a struct of reader/writer/connection/ip
    //BUG HOSTNAME GOES TO IP RIGHT NOW. NEED FIXING
    client := &Client{In: reader, Out: writer, Nick: "", User: "", Real: "",IP: ip, Host: ip, Conn: c, ch: ch}
    clientList.PushBack(*client)
    // Grabs NICK, USER, and REAL name before continuing
    for {
        input, err := client.In.ReadString('\n')
        fmt.Println(input)

        if err != nil {
            c.Close()
            log.Fatal(err)
            break
        } else {
            splitStr := strings.Split(input, " ")
            switch splitStr[0] {
            case "NICK":
                client.Nick = strings.Replace(splitStr[1], "\r\n", "", 1)
            case "USER":
                client.User = splitStr[1]
                client.Real = strings.Join(splitStr[4:], " ")[1:]
                }
            //Kick out of for loop once the username and nickname has been populated
            if client.User != "" && client.Nick != "" {
                break
            }
        }
    }
    //run a go routine of what to do after a connection has been set up
    go func() {
        defer close(msgs)
            cli := client.User + "@" + client.Host
            for {
                input, err := client.In.ReadString('\n')
                    if err != nil {
                        fmt.Println(client.Nick + " has quit.")
                            client.Conn.Close()
                            break
                    } else {
                        //IRC commands, most likely can do better, kinda hacked
                        splitInput := strings.Split(input, " ")
                        if splitInput[0] == "QUIT" {
                            go client.Conn.Close()
                            break
                        } else if splitInput[0] == "PING" {
                            msgs <- "PONG " + strings.Replace(splitInput[1], "\r\n", "", 1)
                        } else {
                            msgs <- cli + ": " + input
                            }
                        }
                    }
            }()
//go client.connectionPost(msgs)
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
            fmt.Printf("New Client: %v \n", client.Conn)
            clients[client.Conn] = client.ch
        case conn := <-rmChan:
            fmt.Printf("Client left: %v \n", conn)
            delete(clients, conn)
        }
    }
}
