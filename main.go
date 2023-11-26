package main

import (
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	PORT = 8888
	IP   = "127.0.0.1"
)

var records = map[string]string{
	"plswork.lol": "123.123.123.123",
}

func main() {

	laddr := net.UDPAddr{
		IP:   net.ParseIP(IP),
		Port: PORT,
	}
	u, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		log.Printf("error while listening: %v", err)
	}

	log.Printf("Listening for DNS requests on %s:%d", IP, PORT)

	for {
		tmp := make([]byte, 1024)
		_, clientAddr, err := u.ReadFrom(tmp)
		if err != nil {
			log.Printf("error reading request: %v", err)
		}
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, ok := dnsPacket.(*layers.DNS)
		if !ok {
			log.Printf("unsupported type packet: %T", dnsPacket)
		}
		err = serveDNS(u, &clientAddr, tcp)
		if err != nil {
			log.Print(err)
		}
	}
}

func serveDNS(u *net.UDPConn, clientAddr *net.Addr, request *layers.DNS) error {
	var dnsAnswer layers.DNSResourceRecord

	reply := request
	questionRecord := string(request.Questions[0].Name)

	ip, ok := records[questionRecord]
	if !ok {
		return fmt.Errorf("no record found for %s", questionRecord)
	}
	log.Printf("Resolving %s to %s", questionRecord, ip)

	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = net.ParseIP(ip)
	dnsAnswer.Name = []byte(questionRecord)
	dnsAnswer.Class = layers.DNSClassIN

	reply.QR = true
	reply.ANCount = 1
	reply.OpCode = layers.DNSOpCodeQuery
	reply.AA = true
	reply.Answers = append(reply.Answers, dnsAnswer)
	reply.ResponseCode = layers.DNSResponseCodeNoErr

	buf := gopacket.NewSerializeBuffer()
	err := reply.SerializeTo(buf, gopacket.SerializeOptions{})
	if err != nil {
		return fmt.Errorf("error serializing reply: %v", err)
	}

	u.WriteTo(buf.Bytes(), *clientAddr)
	return nil
}
