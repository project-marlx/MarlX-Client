package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func main() {
	// test-tkn: 512e3ded82cf54bb146b94f9ada92b4f9d444d7d8e42a92b0035e43dc226a3fc
	// online-test-tkn: 335587594ca9431aa1ae5aba75abac95c35c32a46da9bd1f3d7241b8ea496add

	fmt.Println("[MarlX-Client]: Booting up ... \n")

	exe_dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// tip := "127.0.0.1"
	tip := "193.80.220.151"
	// fip := "127.0.0.1"
	fip := "192.168.0.102"

	tcpConn, err := socks.GetConnectedSocket(tip, fip)
	if err != nil {
		fmt.Println("[-] Server is not up! ")
		return
	}

	fmt.Printf("[MarlX-Client]: Connected to %s:8024\n", tip)

	priv, err := RSAWrapper.GenerateKey()
	if err != nil {
		fmt.Println("[-] Key generation failed!")
		return
	}

	var ch config.ClientHandle
	ch.Quit = false
	// ch.Channel = make(chan error)

	fmt.Println("[MarlX-Client]: Beginning conversation ... ")
	conversations.Handle(tcpConn, priv, streams, &streams_mutex, &ch, exe_dir)

	// in_reader := bufio.NewReader(os.Stdin)

	// for {
	// 	fmt.Print("user@marlx:$ ")
	// 	cmd, err := in_reader.ReadString('\n')

	// 	if err != nil {
	// 		log.Println("Error: " + err.Error())
	// 		continue
	// 	}

	// 	cmd = strings.Replace(cmd, "\n", "", -1)
	// 	cmd = strings.Replace(cmd, "\r", "", -1)

	// 	client.HandleCommand(cmd, &ch)
	// }
}
