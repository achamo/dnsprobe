package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

func main() {
	name := os.Args[1]
	record_type := os.Args[2]
	server := os.Args[3]
	namespace := os.Getenv("COLLECTD_HOSTNAME")
	interval, _ := strconv.Atoi(os.Getenv("COLLECTD_INTERVAL"))
	if interval == 0 {
		interval = 60
	}
	if namespace == "" {
		namespace = "localhost"
	}
	qt, ok := dns.StringToType[record_type]
	if !ok {
		log.Fatalf("%s not found\n", record_type)
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), qt)
	m.RecursionDesired = true
	clients := make(map[string]*dns.Client)
	clients["udp"] = new(dns.Client)
	clients["tcp"] = new(dns.Client)
	clients["tcp"].Net = "tcp"
	for {
		for protocol, c := range clients {
			success := 1
			r, rtt, err := c.Exchange(m, net.JoinHostPort(server, "53"))
			if err != nil || r.Rcode != dns.RcodeSuccess {
				success = 0
			}
			fmt.Printf("PUTVAL %s/godnsprobe-%s/gauge-time/%s-%s interval=%d N:%.6f\n", namespace, protocol, strings.Replace(name, ".", "_", -1), record_type, interval, rtt.Seconds())
			fmt.Printf("PUTVAL %s/godnsprobe-%s/gauge-success/%s-%s interval=%d N:%d\n", namespace, protocol, strings.Replace(name, ".", "_", -1), record_type, interval, success)
		}
		duration, _ := time.ParseDuration(fmt.Sprintf("%ds", interval))
		time.Sleep(duration)
	}
}
