package dnsResolver

import (
	"log"
	"net"
	"time"

	"github.com/mcombeau/dns-tools/dns"
)

var rootServers []string

func FetchRootServers() {
	rootServerNames := []string{
		"a.root-servers.net.",
		"b.root-servers.net.",
		"c.root-servers.net.",
		"d.root-servers.net.",
		"e.root-servers.net.",
		"f.root-servers.net.",
		"g.root-servers.net.",
		"h.root-servers.net.",
		"i.root-servers.net.",
		"j.root-servers.net.",
		"k.root-servers.net.",
		"l.root-servers.net.",
		"m.root-servers.net.",
	}

	var rootServerIPs []string
	for _, serverName := range rootServerNames {
		ip, err := resolveWithPublicDNS(serverName)
		if err != nil {
			log.Printf("failed to resolve IP for server %s: %v", serverName, err)
			continue
		}
		rootServerIPs = append(rootServerIPs, ip)
	}

	if len(rootServerIPs) < 1 {
		log.Panic("could not resolve IPs for root servers")
	}

	rootServers = rootServerIPs
	log.Printf("fetched root servers: %v", rootServers)
}

func resolveWithPublicDNS(serverName string) (ip string, err error) {
	publicDNSServer := "1.1.1.1:53"

	conn, err := net.Dial("udp", publicDNSServer)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	query, err := dns.CreateDNSQuery(serverName, dns.A, false)
	if err != nil {
		return "", err
	}

	_, err = conn.Write(query)
	if err != nil {
		return "", err
	}

	response := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Read(response)
	if err != nil {
		return "", err
	}

	decodedResponse, err := dns.DecodeMessage(response)
	if err != nil {
		return "", err
	}

	return decodedResponse.Answers[0].RData.String(), nil
}
