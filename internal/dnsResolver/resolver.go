package dnsResolver

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"

	"github.com/mcombeau/dns-tools/dns"
)

func ResolveDNSQuery(dnsRequest []byte) (response []byte, err error) {
	log.Printf("ResolveDNSQuery() got DNS query: %v", dnsRequest)

	response, err = queryServers(rootServers[:], dnsRequest)
	if err != nil {
		return nil, err
	}

	return response, fmt.Errorf("failed to get responses from any root server")
}

func queryServers(serverList []string, dnsRequest []byte) (response []byte, err error) {
	for _, server := range serverList {
		fmt.Printf("============ QUERYING SERVER %s\n", server)
		response, err := sendDNSQuery(server, dnsRequest)
		if err != nil {
			log.Printf("failed to query server %s: %v", server, err)
			continue
		}

		dnsParsedResponse, err := dns.DecodeMessage(response)
		if err != nil {
			log.Printf("failed to parse response from server %s: %v", server, err)
			return nil, err
		}

		if len(dnsParsedResponse.Answers) > 0 {
			fmt.Print("-------- GOT ANSWER\n")
			dns.PrintMessage(dnsParsedResponse)
			return response, nil
		}

		if len(dnsParsedResponse.NameServers) > 0 {
			authorityServers := extractAuthorityServers(dnsParsedResponse)
			if len(authorityServers) > 0 {
				return queryServers(authorityServers, dnsRequest)
			}
		}
	}
	return nil, fmt.Errorf("failed to resolve DNS query")
}

func extractAuthorityServers(dnsMessage dns.Message) (serverList []string) {
	// for _, authority := range dnsMessage.NameServers {
	// 	if authority.RType == dns.NS {
	// 		fmt.Printf("++ found a nameserver record: %d\n", authority.RType)
	// 		nsRecord := authority.RData.String()
	// 		fmt.Printf("++ NS RECORD: %s\n", nsRecord)
	// 		ip, err := netip.ParseAddr(nsRecord)
	// 		fmt.Printf("++ parsed IP: %v\n", ip)
	// 		if err == nil {
	// 			serverList = append(serverList, nsRecord)

	// 			fmt.Printf("++ serverList: %v\n", serverList)

	// 		}
	// 	}
	// }
	for _, additional := range dnsMessage.Additionals {
		if additional.RType == dns.A {
			fmt.Printf("++ found a nameserver record: %d\n", additional.RType)
			aRecord := additional.RData.String()
			fmt.Printf("++ NS RECORD: %s\n", aRecord)
			ip, err := netip.ParseAddr(aRecord)
			fmt.Printf("++ parsed IP: %v\n", ip)
			if err == nil {
				serverList = append(serverList, aRecord)

				fmt.Printf("++ serverList: %v\n", serverList)

			}
		}
	}
	fmt.Printf("--- SERVER LIST: %v\n", serverList)
	return serverList
}

func sendDNSQuery(server string, dnsRequest []byte) (response []byte, err error) {
	serverAddr, err := net.ResolveUDPAddr("udp", server+":53")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(dnsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send DNS request to server: %w", err)
	}

	receivedResponse := [4096]byte{}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(receivedResponse[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read response from root server: %w", err)
	}

	fmt.Printf("Response received from root server was length: %d\n", n)

	return receivedResponse[:n], nil
}
