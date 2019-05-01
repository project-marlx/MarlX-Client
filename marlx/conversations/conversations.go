// Package conversations handles client/server
// conversations for the MarlX-Project.
package conversations

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/MattMoony/MarlX-Client/marlx/config"

	"crypto/cipher"
	"crypto/rsa"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"

	"github.com/MattMoony/MarlX-Client/crypto/AESWrapper"
	"github.com/MattMoony/MarlX-Client/socks"
	"github.com/MattMoony/MarlX-Client/system/diskinfo"
)

func GetClientConfiguration(conf_loc string) (config.ClientConfig, error) {
	var conf config.ClientConfig

	c, err := ioutil.ReadFile(conf_loc)
	if err != nil {
		log.Println(err.Error())
		return conf, errors.New(fmt.Sprintf("Check your %s ... (existence, etc.) ", conf_loc))
	}

	err = json.Unmarshal(c, &conf)
	if err != nil {
		log.Println(err.Error())
		return conf, errors.New(fmt.Sprintf("Check your %s ... (json format, etc.) ", conf_loc))
	}

	return conf, nil
}

// GetHostDiskinfoUpdate returns the
// current state of the host.
func GetHostDiskinfoUpdate(conf_loc string) (socks.DiskinfoUpdate, error) {
	var diu socks.DiskinfoUpdate

	conf, err := GetClientConfiguration(conf_loc)
	if err != nil {
		return diu, err
	}

	info := diskinfo.GetDiskInfo(conf.Store_dir)
	diu.FreeBytes = info.Free
	diu.TotalBytes = info.Total

	diu.Hostname, _ = os.Hostname()
	diu.MTU = conf.MTU
	return diu, nil
}

// SendDiskUpdates sends disk updates
// to the server
func SendDiskUpdates(enc *gob.Encoder, aesgcm cipher.AEAD, spub *rsa.PublicKey, exe_dir string) {
	for {
		diu, err := GetHostDiskinfoUpdate(path.Join(exe_dir, "client.json"))
		if err != nil {
			log.Println(err.Error())
			continue
		}

		var ac socks.MarlXActionCommand
		ac.Action = socks.ACTION_DISKINFO_UPDATE

		encoded, _ := json.Marshal(diu)
		ac.Body = encoded
		encodedAc, _ := json.Marshal(ac)

		// log.Println("sending diskinfo update")
		socks.SendAESMessage(enc, aesgcm, spub, encodedAc)

		// send update every minute
		time.Sleep(60000 * time.Millisecond)
	}
}

// SendFile should be used to asynchronously
// respond with the contents of a file.
func SendFile(rfi socks.RequestedFileInfo, frh socks.FileResponseHeader, f *os.File, enc *gob.Encoder, aesgcm cipher.AEAD,
	spub *rsa.PublicKey) {
	var ac socks.MarlXActionCommand
	var ff socks.FileFragment

	ff.StreamToken = rfi.StreamToken
	ff.Total = int64(math.Ceil(float64(frh.Size) / float64(frh.MTU)))

	temp := make([]byte, frh.MTU)
	ac.Action = socks.ACTION_RESPOND_FILE_CONTENT

	for i := int64(0); i < ff.Total; i++ {
		if i == ff.Total-1 {
			temp = make([]byte, frh.Size-i*frh.MTU)
		}

		_, err := f.Read(temp)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		ff.Content = temp
		ff.Index = i

		encb, err := json.Marshal(ff)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		ac.Body = encb

		encb, err = json.Marshal(ac)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		socks.SendAESMessage(enc, aesgcm, spub, encb)
		// log.Println("sent next part")
	}

	f.Close()
	// log.Println("Transferred file ... ")
}

// Handle handles the main client-server
// conversation
func Handle(conn *net.TCPConn, priv *rsa.PrivateKey, streams map[string]*os.File, streams_mutex *sync.RWMutex,
	handle *config.ClientHandle, exe_dir string) error {
	var enc = gob.NewEncoder(conn)
	var dec = gob.NewDecoder(conn)
	var err error

	var spub rsa.PublicKey

	err = socks.RSAKeyExchange(enc, dec, priv, &spub)
	if err != nil {
		log.Println(err.Error())
		// handle.Channel <- err
		return err
	}

	var AESKey = AESWrapper.GenerateKey()
	err = socks.SendRSAMessage(enc, priv, &spub, AESKey)
	if err != nil {
		log.Println(err.Error())
		// handle.Channel <- err
		return err
	}

	aesgcm, err := AESWrapper.GenerateAESGCM(AESKey)
	if err != nil {
		log.Println(err.Error())
		// handle.Channel <- err
		return err
	}

	// handle.Channel <- nil

	var actionCommand socks.MarlXActionCommand
	var plainmsg []byte

	for !handle.Quit {
		plainmsg, err = socks.ReceiveAESMessage(dec, aesgcm, priv)
		if err != nil {
			log.Fatal(err)
		}

		// log.Println(handle)

		// log.Printf("%s\n", plainmsg)

		if plainmsg == nil || len(plainmsg) == 0 {
			log.Fatal(err)
		}

		err = json.Unmarshal(plainmsg, &actionCommand)
		if err != nil {
			log.Fatal(err)
		}

		switch actionCommand.Action {
		// Server commands client to identify
		// itself.
		case socks.ACTION_IDENTIFY:
			actionCommand.Action = 1

			conf, err := GetClientConfiguration(path.Join(exe_dir, "client.json"))
			if err != nil {
				log.Println(err.Error())
				break
			}

			tkn, err := hex.DecodeString(conf.Token)
			if err != nil {
				log.Println(err.Error())
				break
			}

			actionCommand.Body = tkn

			msg, err := json.Marshal(actionCommand)
			if err != nil {
				log.Println(err.Error())
				break
			}

			err = socks.SendAESMessage(enc, aesgcm, &spub, msg)
			if err != nil {
				log.Println(err.Error())
				break
			}

			fmt.Println("[Marlx-Client]: Identified! ")
			go SendDiskUpdates(enc, aesgcm, &spub, exe_dir)
		// Server commands client to update
		// its diskinfo.
		case socks.ACTION_UPDATE_DISKINFO:
			diu, err := GetHostDiskinfoUpdate(path.Join(exe_dir, "client.json"))
			if err != nil {
				log.Println(err.Error())
				break
			}

			var ac socks.MarlXActionCommand
			ac.Action = socks.ACTION_DISKINFO_UPDATE

			encoded, _ := json.Marshal(diu)
			ac.Body = encoded
			encodedAc, _ := json.Marshal(ac)

			fmt.Println("[MarlX-Client]: Answering diskinfo-update request ... ")
			socks.SendAESMessage(enc, aesgcm, &spub, encodedAc)
		// Server sends info about a file to-be-transferred.
		case socks.ACTION_STORE_FILE_HEADER:
			var fih socks.FileInfoHeader
			err := json.Unmarshal(actionCommand.Body, &fih)
			if err != nil {
				log.Println(err.Error())
				break
			}

			conf, err := GetClientConfiguration(path.Join(exe_dir, "client.json"))
			if err != nil {
				log.Println(err.Error())
				break
			}

			u_folder_p := path.Join(conf.Store_dir, fmt.Sprintf("%x", fih.UserToken))

			if _, err := os.Stat(u_folder_p); os.IsNotExist(err) {
				os.Mkdir(u_folder_p, os.ModePerm)
			}

			f, err := os.Create(u_folder_p + "\\" + fmt.Sprintf("%x", fih.UniqueId))
			if err != nil {
				log.Println(err.Error())
				break
			}

			for i := int32(0); i < fih.FragCount; i++ {
				plainmsg, err = socks.ReceiveAESMessage(dec, aesgcm, priv)
				if err != nil {
					log.Println(err.Error())
					continue
				}

				var ac socks.MarlXActionCommand
				err = json.Unmarshal(plainmsg, &ac)
				if err != nil {
					log.Println(err.Error())
					continue
				}

				if ac.Action != socks.ACTION_STORE_FILE_CONTENT {
					log.Println("Expected file content!")
					continue
				}

				var ff socks.FileFragment
				err = json.Unmarshal(ac.Body, &ff)

				f.Write(ff.Content)
			}

			f.Close()
			log.Printf("[MarlX-Client]: Received encrypted file \"%s\" ... \n", fih.Name)
		case socks.ACTION_REQUEST_FILE:
			var rfi socks.RequestedFileInfo

			err := json.Unmarshal(actionCommand.Body, &rfi)
			if err != nil {
				log.Println(err.Error())
				break
			}

			conf, err := GetClientConfiguration(path.Join(exe_dir, "client.json"))
			if err != nil {
				log.Println(err.Error())
				break
			}

			u_folder_p := path.Join(conf.Store_dir, fmt.Sprintf("%x", rfi.UserToken))
			if _, err := os.Stat(u_folder_p); os.IsNotExist(err) {
				log.Println(err.Error())
				break
			}

			f, err := os.Open(u_folder_p + "\\" + fmt.Sprintf("%x", rfi.UniqueId))
			if err != nil {
				log.Println(err.Error())
				break
			}

			var frh socks.FileResponseHeader

			fi, err := os.Stat(u_folder_p + "\\" + fmt.Sprintf("%x", rfi.UniqueId))
			if err != nil {
				log.Println(err.Error())
				break
			}

			frh.Size = fi.Size()
			frh.MTU = int64(conf.MTU)
			frh.StreamToken = rfi.StreamToken

			var ac socks.MarlXActionCommand
			ac.Action = socks.ACTION_RESPOND_FILE_HEADER

			encb, err := json.Marshal(frh)
			if err != nil {
				log.Println(err.Error())
				break
			}
			ac.Body = encb

			encb, err = json.Marshal(ac)
			if err != nil {
				log.Println(err.Error())
				break
			}

			socks.SendAESMessage(enc, aesgcm, &spub, encb)
			fmt.Printf("[MarlX-Client]: Sending encrypted file \"%s\" ... \n", f.Name())
			go SendFile(rfi, frh, f, enc, aesgcm, &spub)
		case socks.ACTION_DELETE_FILE:
			var dreq socks.DeleteRequest

			err := json.Unmarshal(actionCommand.Body, &dreq)
			if err != nil {
				log.Println(err.Error())
				break
			}

			conf, err := GetClientConfiguration(path.Join(exe_dir, "client.json"))
			if err != nil {
				log.Println(err.Error())
				break
			}

			u_folder_p := path.Join(conf.Store_dir, strings.Split(dreq.UniqueId, "_")[0])
			if _, err := os.Stat(u_folder_p); os.IsNotExist(err) {
				log.Println(err.Error())
				break
			}

			if err = os.Remove(u_folder_p + "\\" + strings.Split(dreq.UniqueId, "_")[1]); err != nil {
				log.Println(err.Error())
				break
			}

			fmt.Printf("[MarlX-Client]: Irreversably deleted encrypted file \"%s\" ... \n", strings.Split(dreq.UniqueId, "_"))
		}
	}

	fmt.Println("\n[MarlX-Client]: Quitting now ... ")

	actionCommand.Action = socks.ACTION_CLOSE_SOCKET
	actionCommand.Body = []byte("")

	encb, err := json.Marshal(actionCommand)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	socks.SendAESMessage(enc, aesgcm, &spub, encb)

	err = conn.Close()
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Println("[MarlX-Client]: Quit! ")
	return nil
}
