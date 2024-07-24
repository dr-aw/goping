package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"time"
)

func main() {
	attempts := flag.Int("a", 3, "number of ping attempts")
	timeInt := flag.Int("t", 2, "ping timeout in seconds")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: go run ping.go [-a attempts] [-t timeout] <host>")
		return
	}
	addr := flag.Arg(0)
	timeOut := time.Duration(*timeInt) * time.Second

	fmt.Printf("___________________________\nHost:\t\t%s\n", addr)
	fmt.Printf("Attempts:\t%d\n", *attempts)
	fmt.Printf("Timeout:\t%v\n___________________________\n", timeOut)

	for n := 1; n < *attempts+1; n++ {
		err := ping(addr, timeOut)
		if err != nil {
			fmt.Printf("%d\tPing to %s failed: %v\n", n, addr, err)
		} else {
			fmt.Printf("%d\tPing to %s succeeded\n", n, addr)
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func ping(addr string, timeOut time.Duration) error {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer c.Close()

	dst, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
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
		return err
	}

	start := time.Now()

	// Send Echo Request
	if _, err := c.WriteTo(msgBytes, dst); err != nil {
		return err
	}

	// Set time-out
	c.SetDeadline(time.Now().Add(timeOut * time.Second))

	reply := make([]byte, 1500)
	n, peer, err := c.ReadFrom(reply)
	if err != nil {
		return err
	}

	duration := time.Since(start)

	rm, err := icmp.ParseMessage(1, reply[:n])
	if err != nil {
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
