package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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

// serveDNS has a 1 in 10 chance to redirect the DNS query to a rickroll. It
// either acts like any other DNS and replies with an IP or just returns a CNAME
// to a rickroll.
func serveDNS(u *net.UDPConn, clientAddr *net.Addr, request *layers.DNS) error {
	var answer string
	var err error

	reply := request
	questionRecord := string(request.Questions[0].Name)

	log.Printf("Resolving %s", questionRecord)

	// 1 in 10 chance to resolve a rickroll
	if n := rand.Intn(10); n == 1 {
		questionRecord = RICKROLL
		printASCII()
	}

	answer, err = resolveHost(questionRecord)
	if err != nil {
		return err
	}

	replyData, err := DNSreply(reply, answer, questionRecord)
	if err != nil {
		return err
	}
	u.WriteTo(replyData, *clientAddr)
	return nil
}

// DNSreply returns the reply to the query with structured byte data.
func DNSreply(reply *layers.DNS, response, question string) ([]byte, error) {
	var dnsAnswer layers.DNSResourceRecord

	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = net.ParseIP(response)
	dnsAnswer.Name = []byte(question)
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
		return nil, fmt.Errorf("error serializing reply: %v", err)
	}

	return buf.Bytes(), nil
}

// resolveHost uses the system DNS service to repond with actual IP address of
// the request host.
func resolveHost(host string) (string, error) {
	resolver := net.Resolver{}
	ips, err := resolver.LookupHost(context.Background(), host)
	if err != nil {
		return "", fmt.Errorf("error resolving host: %w", err)
	}
	return ips[0], nil
}

// sick ASCII
func printASCII() {
	fmt.Println(RED + `

 _______  _        _            _______  _______  _______  ______   _______      _        _______  _______  ______      _________ _______      _______ _________ _______  _       
(  ___  )( \      ( \          (  ____ )(  ___  )(  ___  )(  __  \ (  ____ \    ( \      (  ____ \(  ___  )(  __  \     \__   __/(  ___  )    (  ____ )\__   __/(  ____ \| \    /\
| (   ) || (      | (          | (    )|| (   ) || (   ) || (  \  )| (    \/    | (      | (    \/| (   ) || (  \  )       ) (   | (   ) |    | (    )|   ) (   | (    \/|  \  / /
| (___) || |      | |          | (____)|| |   | || (___) || |   ) || (_____     | |      | (__    | (___) || |   ) |       | |   | |   | |    | (____)|   | |   | |      |  (_/ / 
|  ___  || |      | |          |     __)| |   | ||  ___  || |   | |(_____  )    | |      |  __)   |  ___  || |   | |       | |   | |   | |    |     __)   | |   | |      |   _ (  
| (   ) || |      | |          | (\ (   | |   | || (   ) || |   ) |      ) |    | |      | (      | (   ) || |   ) |       | |   | |   | |    | (\ (      | |   | |      |  ( \ \ 
| )   ( || (____/\| (____/\    | ) \ \__| (___) || )   ( || (__/  )/\____) |    | (____/\| (____/\| )   ( || (__/  )       | |   | (___) |    | ) \ \_____) (___| (____/\|  /  \ \
|/     \|(_______/(_______/    |/   \__/(_______)|/     \|(______/ \_______)    (_______/(_______/|/     \|(______/        )_(   (_______)    |/   \__/\_______/(_______/|_/    \/
                                                                                                                                                                                  

` + RESET)
}
