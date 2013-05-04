package main
import ("fmt"
        "net"
)

func main() {
    go net.Listen("tcp", ":3003")
    conn, err := net.Dial("tcp", ":3003")
    if err != nil {
        fmt.Println(err)
        }
    hellovar2, err := conn.Read([]byte("Hello World"))
    hellovar, err := conn.Write(hellovar2)
    fmt.Println(hellovar2)
    fmt.Println(message)
    }