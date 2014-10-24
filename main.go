package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/0xAX/notificator"
)

var laddr *net.UDPAddr
var raddr *net.UDPAddr

var con *net.UDPConn

var servers = make(map[string]Server)

var notify *notificator.Notificator

func main() {
	addresses, _ := net.InterfaceAddrs()

	currentAddr := addresses[0]

	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/cod.png",
		AppName:     "CoD4 Server Notifier",
	})

	laddr, _ = net.ResolveUDPAddr("udp", ":28961")
	raddr, _ = net.ResolveUDPAddr("udp", "255.255.255.255:28960")

	var errs error

	con, errs = net.DialUDP("udp", laddr, raddr)

	if errs != nil {
		fmt.Println("Error dialing UDP:", errs)
		return
	}

	udpAddress, err2 := net.ResolveUDPAddr("udp4", currentAddr.String()+":28961")

	if err2 != nil {
		fmt.Println("Error resolving UDP address:", err2)
		return
	}

	ln, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		fmt.Println("There was a problem listening:", err)
		return
	}

	fmt.Println("We're watching for Call of Duty 4 Servers on the local network")

	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)

			askForServer()
		}
	}()

	for {
		time.Sleep(1000 * time.Millisecond)

		var buf []byte = make([]byte, 200)

		n, address, err := ln.ReadFromUDP(buf)

		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}

		if address != nil && n > 0 {
			parseMessage(strings.TrimSpace(string(buf)))
		}
	}
}

func askForServer() {
	_, err := con.Write([]byte{0xff, 0xff, 0xff, 0xff, 0x67, 0x65, 0x74, 0x69, 0x6e, 0x66, 0x6f, 0x20, 0x78, 0x78, 0x78})

	if err != nil {
		fmt.Println("Error writing broadcast for servers:", err)
		return
	}
}

func parseMessage(msg string) {
	// Ignore getInfo requests
	if strings.Contains(msg, "getinfo xxx") {
		return
	}

	if strings.Contains(msg, "infoResponse") {
		parseServer(msg)
	}
}

func parseServer(msg string) {
	split := strings.Split(msg, "\\")

	server := Server{
		Hostname: split[6],
		Map:      split[8],
		GameType: split[12],
	}

	server.MaxPlayers, _ = strconv.ParseInt(split[10], 10, 64)
	server.Pure = parseBool(split[14])
	server.FriendlyFire = parseBool(split[16])
	server.Hardcore = parseBool(split[18])
	server.Something, _ = strconv.ParseInt(split[20], 10, 64)
	server.Modded = parseBool(split[22])
	server.Voice = parseBool(split[24])
	server.Punkbuster = parseBool(split[26])

	if _, ok := servers[server.Hostname]; !ok {
		servers[server.Hostname] = server

		dir, _ := os.Getwd()

		notify.Push("New Server!", server.String(), dir+"/icon/cod.png")
	}
}

type Server struct {
	Hostname     string
	Map          string
	MaxPlayers   int64
	GameType     string
	Pure         bool
	FriendlyFire bool
	Hardcore     bool
	Something    int64
	Modded       bool
	Voice        bool
	Punkbuster   bool
}

func (s Server) String() (data string) {
	data += "Hostname: " + s.Hostname + "\n"
	data += "Map: " + s.Map + "\n"
	data += "MaxPlayers: " + strconv.FormatInt(s.MaxPlayers, 10) + "\n"
	data += "GameType: " + s.GameType + "\n"
	data += "Pure: " + strconv.FormatBool(s.Pure) + "\n"
	data += "FriendlyFire: " + strconv.FormatBool(s.FriendlyFire) + "\n"
	data += "Hardcore: " + strconv.FormatBool(s.Hardcore) + "\n"
	data += "Modded: " + strconv.FormatBool(s.Modded) + "\n"
	data += "Voice: " + strconv.FormatBool(s.Voice) + "\n"
	data += "Punkbuster: " + strconv.FormatBool(s.Punkbuster) + "\n"

	return data
}

func parseBool(str string) bool {
	return str == "1"
}
