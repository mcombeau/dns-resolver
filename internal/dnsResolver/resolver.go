package dnsResolver

import (
	"fmt"
	"log"
	"net"
	"time"
)

func ResolveDNSQuery(dnsRequest []byte) (response []byte, err error) {
	log.Printf("ResolveDNSQuery() got DNS query: %v", dnsRequest)

	for _, rootServer := range rootServers {
		rootAddr, err := net.ResolveUDPAddr("udp", rootServer+":53")
		if err != nil {
			log.Printf("failed to resolve root server address: %v", err)
			continue
		}

		conn, err := net.DialUDP("udp", nil, rootAddr)
		if err != nil {
			log.Printf("failed to dial root server: %v", err)
			continue
		}
		defer conn.Close()

		_, err = conn.Write(dnsRequest)
		if err != nil {
			log.Printf("failed to send DNS request to root server: %v", err)
			continue
		}

		receivedResponse := [4096]byte{}
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(receivedResponse[:])
		if err != nil {
			log.Printf("failed to read response from root server: %v", err)
			continue
		}
		fmt.Printf("Response received from root server was length: %d\n", n)

		return receivedResponse[:n], nil
	}

	return response, fmt.Errorf("failed to get responses from any root server")
}
