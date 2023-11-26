package main

import (
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	PORT     = 8888
	IP       = "127.0.0.1"
	RICKROLL = "rickroll.it"
    RESET  = "\033[0m"
    RED    = "\033[31m"
)

func main() {
	printASCII()
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

	log.Printf("Resolving %s to %s", questionRecord, RICKROLL)

	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = net.ParseIP("127.0.0.1")
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

func printASCII() {
	fmt.Println(RED+`

 _______  _        _            _______  _______  _______  ______   _______      _        _______  _______  ______      _________ _______      _______  _        _______             _   
(  ___  )( \      ( \          (  ____ )(  ___  )(  ___  )(  __  \ (  ____ \    ( \      (  ____ \(  ___  )(  __  \     \__   __/(  ___  )    (  ___  )( (    /|(  ____ \           ( \  
| (   ) || (      | (          | (    )|| (   ) || (   ) || (  \  )| (    \/    | (      | (    \/| (   ) || (  \  )       ) (   | (   ) |    | (   ) ||  \  ( || (    \/     _      \ \ 
| (___) || |      | |          | (____)|| |   | || (___) || |   ) || (_____     | |      | (__    | (___) || |   ) |       | |   | |   | |    | |   | ||   \ | || (__        (_)      ) )
|  ___  || |      | |          |     __)| |   | ||  ___  || |   | |(_____  )    | |      |  __)   |  ___  || |   | |       | |   | |   | |    | |   | || (\ \) ||  __)                | |
| (   ) || |      | |          | (\ (   | |   | || (   ) || |   ) |      ) |    | |      | (      | (   ) || |   ) |       | |   | |   | |    | |   | || | \   || (           _       ) )
| )   ( || (____/\| (____/\    | ) \ \__| (___) || )   ( || (__/  )/\____) |    | (____/\| (____/\| )   ( || (__/  )       | |   | (___) |    | (___) || )  \  || (____/\    (_)_    / / 
|/     \|(_______/(_______/    |/   \__/(_______)|/     \|(______/ \_______)    (_______/(_______/|/     \|(______/        )_(   (_______)    (_______)|/    )_)(_______/      ( )  (_/  
                                                                                                                                                                               |/        

`+RESET)
}
