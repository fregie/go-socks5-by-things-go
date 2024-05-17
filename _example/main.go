package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"

	"github.com/things-go/go-socks5"
)

func main() {
	outIP := flag.String("outIP", "0.0.0.0", "local ip")
	serveAddr := flag.String("serveAddr", "0.0.0.0:5001", "local port")
	udpIPStr := flag.String("udpIP", "", "udp ip")
	username := flag.String("username", "", "username")
	password := flag.String("password", "", "password")
	flag.Parse()

	udpip := net.ParseIP(*udpIPStr)
	if udpip == nil {
		log.Printf("udp ip is invalid: %s", *udpIPStr)
	}

	cred := socks5.StaticCredentials{}
	if *username != "" && *password != "" {
		cred[*username] = *password
	}
	localAddr := *outIP + ":0"
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		switch network {
		case "tcp", "tcp4", "tcp6":
			tcpLocalAddr, err := net.ResolveTCPAddr(network, localAddr)
			if err != nil {
				return nil, err
			}
			rAddr, err := net.ResolveTCPAddr(network, addr)
			if err != nil {
				return nil, err
			}
			return net.DialTCP(network, tcpLocalAddr, rAddr)
		case "udp", "udp4", "udp6":
			udpLocalAddr, err := net.ResolveUDPAddr(network, localAddr)
			if err != nil {
				return nil, err
			}
			rAddr, err := net.ResolveUDPAddr(network, addr)
			if err != nil {
				return nil, err
			}
			return net.DialUDP(network, udpLocalAddr, rAddr)
		}
		return nil, errors.New("unsupported network")
	}

	opts := []socks5.Option{
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
		socks5.WithBindIP(udpip),
		socks5.WithDial(dial),
	}
	if len(cred) > 0 {
		opts = append(opts, socks5.WithCredential(cred))
	}
	// Create a SOCKS5 server
	server := socks5.NewServer(opts...)

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", *serveAddr); err != nil {
		log.Fatal(err)
	}
}
