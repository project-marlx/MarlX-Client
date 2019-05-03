package socks

import (
	"fmt"
	"net"
)

// GetConnectedSocket is an alias for:
// GetConnectedSocketOnPort(hostname, 8024)
func GetConnectedSocket(hostname string, from_ip string) (*net.TCPConn, error) {
	return GetConnectedSocketOnPort(fmt.Sprintf("%s:8024", hostname), fmt.Sprintf("%s:8027", from_ip))
}

// GetConnectedSocketOnPort connects the client
// to the specified hostname on the specified
// port.
// Returns: net.Conn + an error (if any occurred)
func GetConnectedSocketOnPort(to_ip string, from_ip string) (*net.TCPConn, error) {
	remoteAddr, err := net.ResolveTCPAddr("tcp4", to_ip)
	if err != nil {
		return nil, err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", from_ip)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", tcpAddr, remoteAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func GetConnectedSocketDefault(hostname string) (*net.TCPConn, error) {
	return GetConnectedSocketOnPortDefault(fmt.Sprintf("%s:8024", hostname));
}

func GetConnectedSocketOnPortDefault(to_ip string) (*net.TCPConn, error) {
	remoteAddr, err := net.ResolveTCPAddr("tcp4", to_ip)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}