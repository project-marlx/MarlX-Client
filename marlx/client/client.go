package client

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/MattMoony/MarlX-Client/crypto/RSAWrapper"
	"github.com/MattMoony/MarlX-Client/marlx/config"
	"github.com/MattMoony/MarlX-Client/marlx/conversations"
	"github.com/MattMoony/MarlX-Client/socks"
)

var (
	streams       = map[string]*os.File{}
	streams_mutex = sync.RWMutex{}
)

func HandleCommand(cmd string, handle *config.ClientHandle) {
	args := strings.Split(cmd, " ")

	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "start":
		if !handle.Quit {
			return
		}

		handle.Quit = false
		handle.Channel = make(chan error)

		tip := "127.0.0.1"
		fip := "127.0.0.1"

		tcpConn, err := socks.GetConnectedSocket(tip, fip)
		if err != nil {
			log.Println(err.Error())
			fmt.Println("[-] Server is not up! ")
			handle.Quit = true
			return
		}

		fmt.Printf("[MarlX-Client]: Connected to %s:8024\n", tip)

		priv, err := RSAWrapper.GenerateKey()
		if err != nil {
			fmt.Println("[-] Key generation failed!")
			handle.Quit = true
			return
		}

		go conversations.Handle(tcpConn, priv, streams, &streams_mutex, handle)

		if succ := <-handle.Channel; succ != nil {
			fmt.Println("[-] Connection to server failed!")
			handle.Quit = true
		} else {
			fmt.Println("[+] Connected to server!")
		}
	case "quit":
		handle.Quit = true
		os.Exit(0)
	case "exit":
		handle.Quit = true
		os.Exit(0)
	case "check":
		if len(args) < 2 {
			fmt.Println("[-] Minimum 2 arguments required for this command!")
			return
		}

		switch args[1] {
		case "connection":
			if !handle.Quit {
				fmt.Println("[MarlX-Client]: Connected ... ")
			} else {
				fmt.Println("[MarlX-Client]: Not connected ... ")
			}
		}
	case "get":
		if len(args) < 2 {
			fmt.Println("[-] Minimum 2 arguments required for this command!")
			return
		}

		switch args[1] {
		case "storage-location":
			conf, err := conversations.GetClientConfiguration("./client.json")
			if err != nil {
				log.Println(err.Error())
				break
			}

			fmt.Printf("[MarlX-Client]: storage location = \"%s\"\n", conf.Store_dir)
		case "mtu":
			conf, err := conversations.GetClientConfiguration("./client.json")
			if err != nil {
				log.Println(err.Error())
				break
			}

			fmt.Printf("[MarlX-Client]: MTU = \"%d\"\n", conf.MTU)
		case "token":
			conf, err := conversations.GetClientConfiguration("./client.json")
			if err != nil {
				log.Println(err.Error())
				break
			}

			fmt.Printf("[MarlX-Client]: token = \"%s\"\n", conf.Token)
		}
	}
}
