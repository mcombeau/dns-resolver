package main

// TODO: Gameplan:
// - Setup server as listener on port 5353
// - Query public DNS like 8.8.8.8 or 1.1.1.1 for root servers
// - When a query arrives:
// 		- query root servers, parse response,
// 		- query next server, parse response,
// 		- etc until authoritative response.
// Later:
// 		- add caching
//		- handle multiple concurrent client requests

func main() {
	err := dnsServer.startUDPServer()
}
