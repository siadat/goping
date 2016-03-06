package main

import (
	"github.com/docopt/docopt-go"
	"github.com/fatih/color"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	interval time.Duration = 1 * time.Second
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
		log.Fatalln("FATAL:Marshal", err)
	}
	return
}

func start(ip string, count int, interval time.Duration, iface string) {

	for counter := 1; !shouldEnd(counter-1, count); counter++ {
		c := newConn()
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
			log.Fatalln("FATAL:ParseMessage", err)
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

func shouldEnd(counter, max int) bool {
	shouldSleep := true
	if counter == 0 {
		shouldSleep = false
	}

	if max == -1 {
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

func newConn() (c *icmp.PacketConn) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	c.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Fatalln("FATAL:ListenPacket", err)
	}
	return
}

func main() {
	args, err := docopt.Parse(usage, nil, true, "0.0.0", false, true)
	ip := args["<ip>"].(string)
	iface := "en0"
	count := -1
	if args["-i"] != nil {
		sec, err := strconv.ParseFloat(args["-i"].(string), 64)
		if err != nil || sec < 0 {
			log.Fatalln("-i should be a float >= 0", err)
		}
		interval = time.Duration(sec*1000) * time.Millisecond
	}

	if args["-c"] != nil {
		count, err = strconv.Atoi(args["-c"].(string))
		if err != nil || count < 1 {
			log.Fatalln("-c value should be >= 1", err)
		}
	}
	start(ip, count, interval, iface)
}
