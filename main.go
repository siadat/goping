package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	ProtocolICMP = 1
	EchoMessage  = "hello-from-goping"
	usage        = `Usage: goping <ip> [-c count] [-i interval]
Options:
	-c count      Limit number of requests
	-i interval   Set sleep time duration
`
)

func newMsg(seq int) (msg icmp.Message) {
	msg = icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			Seq:  seq,
			ID:   os.Getpid() & 0xffff,
			Data: []byte(EchoMessage),
		},
	}
	return
}

func marshalMsg(msg icmp.Message) (buff []byte) {
	buff, err := msg.Marshal(nil)
	if err != nil {
		log.Fatal("Marshal", err)
	}
	return
}

func start(ip string, count int, interval time.Duration, iface string) {
	for counter := 1; !shouldStop(counter-1, count, interval); counter++ {
		c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			log.Fatal("ListenPacket", err)
		}
		c.SetDeadline(time.Now().Add(10 * time.Second))

		msg := newMsg(counter)
		writeBuff := marshalMsg(msg)

		if _, err := c.WriteTo(writeBuff, &net.IPAddr{IP: net.ParseIP(ip), Zone: iface}); err != nil {
			color.Set(color.FgRed)
			log.Printf("#%d ERROR %s", counter, err.Error())
			color.Unset()
			continue
		}

		readBuff := make([]byte, 1500)
		startTime := time.Now()
		n, peer, err := c.ReadFrom(readBuff)
		if err != nil {
			color.Set(color.FgRed)
			log.Printf("#%d ERROR %s", counter, err.Error())
			color.Unset()
			continue
		}

		endTime := time.Now()

		rm, err := icmp.ParseMessage(ProtocolICMP, readBuff[:n])
		if err != nil {
			log.Fatal("ParseMessage", err)
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			log.Printf("#%d %d bytes from %v %v", counter, n, peer, endTime.Sub(startTime))
		default:
			color.Set(color.FgRed)
			log.Printf("#%d ERROR got %+v", counter, rm)
			color.Unset()
		}
		c.Close()
	}
}

func shouldStop(counter, max int, interval time.Duration) bool {
	shouldSleep := true
	if counter == 0 {
		shouldSleep = false
	}

	if max == 0 {
		if shouldSleep {
			time.Sleep(interval)
		}
		return false
	}
	if counter >= max {
		return true
	}
	if shouldSleep {
		time.Sleep(interval)
	}
	return false
}

func main() {
	interval := flag.Duration("i", time.Second, "Set sleep time duration")
	count := flag.Int("c", 0, "Limit number of requests")
	flag.Parse()
	ip := flag.Arg(0)
	iface := "en0"
	start(ip, *count, *interval, iface)
}
