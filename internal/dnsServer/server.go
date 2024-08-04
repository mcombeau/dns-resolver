package dnsServer

import (
	"fmt"
	"log"
	"net"
)

const DNSResolverIP = "0.0.0.0"
const DNSResolverPort = 5553

const MaxUDPMessageLength = 512

func StartUDPServer() (err error) {
	addr := net.UDPAddr{
		Port: DNSResolverPort,
		IP:   net.ParseIP(DNSResolverIP),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("failed to set up UDP listener: %w", err)
	}
	defer conn.Close()

	log.Printf("DNS resolver server listening on port: %d", addr.Port)

	buffer := [MaxUDPMessageLength]byte{}
	for {
		_, clientAddr, err := conn.ReadFromUDP(buffer[:])
		if err != nil {
			log.Printf("error reading from UDP: %v", err)
			continue
		}
		// TODO: handle client request here
		log.Printf("read from client %s:%d over UDP: %v", clientAddr.IP.String(), clientAddr.Port, buffer)
	}
}
