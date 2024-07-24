package main

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run ping.go <host> <timeout>")
		return
	}
	addr := os.Args[1]
	time := os.Args[2]
	timeOut, err := strconv.Atoi(time)
	if err != nil {
		fmt.Println("TimeOut should be a number")
	}
	if err := ping(addr, timeOut); err != nil {
		fmt.Printf("Ping to %s failed: %v\n", addr, err)
	}
}

func ping(addr string, timeOut int) error {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Println("Listen")
		return err
	}
	defer c.Close()

	dst, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		fmt.Println("Resolve")
		return err
	}

	// Creating ICMP Message (Echo)
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("PING"),
		},
	}
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		fmt.Println("msgBytes")
		return err
	}

	start := time.Now()

	// Send Echo Request
	if _, err := c.WriteTo(msgBytes, dst); err != nil {
		fmt.Println("WriteTo")
		return err
	}

	// Set time-out
	c.SetDeadline(time.Now().Add(time.Duration(timeOut) * time.Second))

	reply := make([]byte, 1500)
	n, peer, err := c.ReadFrom(reply)
	if err != nil {
		fmt.Println("Reply")
		return err
	}

	duration := time.Since(start)

	rm, err := icmp.ParseMessage(1, reply[:n])
	if err != nil {
		fmt.Println("Parse")
		return err
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		fmt.Printf("Ping to %s succeeded, reply from %s in %v\n", addr, peer, duration)
	default:
		fmt.Printf("Got unexpected reply from %s: %+v\n", peer, rm)
	}
	return nil
}
