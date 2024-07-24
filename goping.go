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
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run ping.go <host> <timeout> <attempts>")
		return
	}
	addr := os.Args[1]
	timeStr := os.Args[2]
	attemptStr := os.Args[3]
	attempts, err := strconv.Atoi(attemptStr)
	if err != nil || attempts < 1 {
		fmt.Println("Attempts should be a number")
	}
	timeInt, err := strconv.Atoi(timeStr)
	if err != nil || timeInt < 1 {
		fmt.Println("TimeOut should be a number")
	}
	timeOut := time.Duration(timeInt)

	for n := 1; n < attempts+1; n++ {
		err := ping(addr, timeOut)
		if err != nil {
			fmt.Printf("%d\tPing to %s failed: %v\n", n, addr, err)
		} else {
			fmt.Printf("%d\tPing to %s succeeded\n", n, addr)
			break
		}
		time.Sleep(5 * time.Second) // Добавляем паузу между попытками
	}
}

func ping(addr string, timeOut time.Duration) error {
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
	c.SetDeadline(time.Now().Add(timeOut * time.Second))

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
		fmt.Printf("Ping to %s succeeded, reply from %s in %v\n\a", addr, peer, duration)
	default:
		fmt.Printf("Got unexpected reply from %s: %+v\n", peer, rm)
	}
	return nil
}
