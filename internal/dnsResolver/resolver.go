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

	return response, nil
}

func queryServers(serverList []string, dnsRequest []byte) (response []byte, err error) {
	dnsParsedRequest, err := dns.DecodeMessage(response)
	if err != nil {
		log.Printf("failed to parse client request: %v", err)
		return nil, err
	}
	domainQuery := dnsParsedRequest.Questions[0].Name

	for _, server := range serverList {
		log.Printf("--> Querying server %s for %s", server, domainQuery)
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

		if dnsParsedResponse.Header.AnswerRRCount > 0 {
			log.Printf("--> Got authoritative answer from server %s for %s", server, domainQuery)
			// fmt.Println("-------------------")
			// dns.PrintMessage(dnsParsedResponse)
			// fmt.Println("-------------------")

			return response, nil
		}

		if dnsParsedResponse.Header.AdditionalRRCount > 0 {
			authorityServers := extractAuthorityServerIPsFromAdditionals(dnsParsedResponse)
			if len(authorityServers) > 0 {
				return queryServers(authorityServers, dnsRequest)
			}
		}
	}
	return nil, fmt.Errorf("failed to resolve DNS query")
}

func extractAuthorityServerIPsFromAdditionals(dnsMessage dns.Message) (serverList []string) {
	for _, additional := range dnsMessage.Additionals {
		if additional.RType == dns.A {
			aRecord := additional.RData.String()
			_, err := netip.ParseAddr(aRecord)
			if err == nil {
				serverList = append(serverList, aRecord)
			}
		}
	}
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

	return receivedResponse[:n], nil
}
