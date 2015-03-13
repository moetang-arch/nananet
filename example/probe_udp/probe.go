package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"net"
)

var (
	data = []byte{0x01, 0x03, 0x05, 0x07, 0x06, 0x04, 0x02, 0x00}
)

var (
	toHost     string
	serverMode string
	bindAddr   string

	interval int
)

func init() {
	flag.StringVar(&serverMode, "s", "", "server mode. -s=XXX:XXX")
	flag.StringVar(&toHost, "c", "", "client mode. -c=XXX:XXX")
	flag.IntVar(&interval, "interval", 5, "keep-alive time. health-check time is 4 times as this value")
}

func main() {
	flag.Parse()
	fmt.Println("host:", toHost)
	fmt.Println("server mode:", serverMode)
	fmt.Println("interval:", interval, "secs")
	fmt.Println("health check:", interval*4, "secs")

	if serverMode != "" {
		runServerMode(interval * 4)
	} else {
		runClientMode()
	}
}

func runClientMode() {
	addr, err := net.ResolveUDPAddr("udp", toHost)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		<-time.After(time.Second * time.Duration(interval))
		n, err := conn.Write(data)
		fmt.Println("write:", n, err, time.Now())
		// perhaps should not uncomment these code. to recover from network problem
		//		if err != nil {
		//			os.Exit(1)
		//		}
	}
}

func runServerMode(healthCheckTime int) {
	addr, err := net.ResolveUDPAddr("udp", serverMode)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ch := make(chan []byte, 10)
	recvd := true
	go func() {
		for {
			b := <-ch
			fmt.Println("recv:", time.Now(), "data:", b)
			recvd = true
		}
	}()

	go func() {
		for {
			<-time.After(time.Second * time.Duration(healthCheckTime))
			if !recvd {
				fmt.Println("timeout:", time.Now())
			}
			recvd = false
		}
	}()

	recvBuf := make([]byte, len(data))
	for {
		n, remoteAddr, err := l.ReadFrom(recvBuf)
		fmt.Println("read from udp:", n, remoteAddr, err)
		if err != nil {
			os.Exit(1)
		}
		ch <- recvBuf[:n]
	}
}
