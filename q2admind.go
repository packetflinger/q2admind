package main

import (
    "bufio"
    "fmt"
    "math/rand"
    "net"
    "os"
    "crypto/rsa"
    "strconv"
    "strings"
    "time"
)

// use a custom buffer struct to keep track of where
// we are in the stream of bytes internally
type MessageBuffer struct {
	buffer []byte
	index  int32
	length int32 // maybe not needed
}

type Player struct {
	clientid     int8
	name         string
	userinfo     string
	frags        int16
	deaths       int16
	suicides     int16
	teleports    int16
	lastteleport int32
}

// this is a Quake 2 Gameserver, and also a client to us
type Server struct {
    id         int32 // this is the database index
    key        int32
    index      int32
    version    int32 // what version are we running
    name       string
    ipaddress  string
    port       int16 // default 27910
    connected  bool
    currentmap string
    enabled    bool
    connection *net.Conn
    players    []Player
    message    MessageBuffer
    encrypted  bool
    publickey  *rsa.PublicKey
    aeskey     []byte          // 16 (128bit)
    nonce      []byte          // 12 for gcm
}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}

		result := strconv.Itoa(rand.Intn(100)) + "\n"
		c.Write([]byte(string(result)))
	}
	c.Close()
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	port := ":" + arguments[1]
	listener, err := net.Listen("tcp", port) // v4 + v6
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	rand.Seed(time.Now().Unix())

	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}

func init() {
    // testing stuff
    public, err := LoadPublicKey("public.pem")
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(public)
    os.Exit(1)
}
