package main

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	PORT     = 8888
	IP       = "0.0.0.0"
	RICKROLL = "rickroll.it"
	RESET    = "\033[0m"
	RED      = "\033[31m"
)

func main() {

	// Address of this DNS server
	laddr := net.UDPAddr{
		IP:   net.ParseIP(IP),
		Port: PORT,
	}
	u, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		log.Printf("error while listening: %v", err)
	}

	log.Printf("Listening for DNS requests on %s:%d", IP, PORT)

	// listen for DNS requests
	for {
		tmp := make([]byte, 1024)
		_, clientAddr, err := u.ReadFrom(tmp)
		if err != nil {
			log.Printf("error reading request: %v", err)
		}
		// parse the UDP request
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, ok := dnsPacket.(*layers.DNS)
		if !ok {
			log.Printf("unsupported type packet: %T", dnsPacket)
		}

		// serve the answer
		err = serveDNS(u, &clientAddr, tcp)
		if err != nil {
			log.Print(err)
		}
	}
}
