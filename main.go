package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/0xAX/notificator"
	"github.com/codegangsta/negroni"
	"github.com/cratonica/trayhost"
)

var laddr *net.UDPAddr
var raddr *net.UDPAddr

var con *net.UDPConn

var servers = make(map[string]Server)
var arrServers []Server

var notify *notificator.Notificator

func main() {
	arrServers = append(arrServers, Server{
		IP:         "192.168.1.24",
		Hostname:   "Gangster Swag",
		Map:        "mp_ambush",
		MaxPlayers: 64,
		GameType:   "Deathmatch",
	})

	arrServers = append(arrServers, Server{
		IP:         "192.168.1.99",
		Hostname:   "John's Server",
		Map:        "mp_district",
		MaxPlayers: 16,
		GameType:   "Deathmatch",
	})

	arrServers = append(arrServers, Server{
		IP:         "192.168.1.103",
		Hostname:   "Teambork",
		Map:        "mp_backlot",
		MaxPlayers: 24,
		GameType:   "Team Deathmatch",
	})

	// EnterLoop must be called on the OS's main thread
	runtime.LockOSThread()

	go func() {
		trayhost.SetUrl("http://localhost:5050")

		mux := http.NewServeMux()

		mux.HandleFunc("/servers.json", handleServers)

		n := negroni.Classic()
		n.UseHandler(mux)
		n.Run(":5050")
	}()

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

	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)

			var buf []byte = make([]byte, 500)

			n, address, err := ln.ReadFromUDP(buf)

			if err != nil {
				fmt.Println("Error reading from connection:", err)
				return
			}

			if address != nil && n > 0 {
				parseMessage(address.String(), strings.TrimSpace(string(buf)))
			}
		}
	}()

	trayhost.EnterLoop("Call of Duty 4 Server Browser", []byte{0xff})
}

func handleServers(w http.ResponseWriter, req *http.Request) {
	askForServer()

	time.Sleep(1000 * time.Millisecond)

	data, _ := json.Marshal(arrServers)

	w.Write(data)
}

func askForServer() {
	_, err := con.Write([]byte{0xff, 0xff, 0xff, 0xff, 0x67, 0x65, 0x74, 0x69, 0x6e, 0x66, 0x6f, 0x20, 0x78, 0x78, 0x78})

	if err != nil {
		fmt.Println("Error writing broadcast for servers:", err)
		return
	}
}

func parseMessage(ip string, msg string) {
	// Ignore getInfo requests
	if strings.Contains(msg, "getinfo xxx") {
		return
	}

	if strings.Contains(msg, "infoResponse") {
		parseServer(ip, msg)
	}
}

func parseServer(ip string, msg string) {
	split := strings.Split(msg, "\\")

	server := Server{
		Hostname: split[6],
		Map:      split[8],
		GameType: split[12],
	}

	server.IP = ip
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
	IP           string
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
