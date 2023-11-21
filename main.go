package main

type Message struct {
	Header     Header
	Question   Question
	Answer     Answer
	Authority  Authority
	Additional Additional
}

func (m *Message) GetMessage() string {
	return ""
}

type Header struct {
	ID      uint16
	QR      bool
	OPCODE  uint8
	AA      bool
	TC      bool
	RD      bool
	RA      bool
	RCODE   uint8
	QDCOUNT uint16
	ANCOUNT uint16
	NSCOUNT uint16
	ARCOUNT uint16
}

type Question struct {
	QNAME  []byte
	QTYPE  uint16
	QCLASS uint16
}

type Answer struct{}
type Authority struct{}
type Additional struct{}
