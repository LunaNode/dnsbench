package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Query struct {
	Name string
	TypeStr string
	Type uint16
}

func (query Query) Exec(client *dns.Client, server string) (time.Duration, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(query.Name), query.Type)
	_, rtt, err := client.Exchange(msg, server)
	return rtt, err
}

type Queries []*Query

func (queries *Queries) String() string {
	str := ""
	for _, query := range *queries {
		if len(str) > 0 {
			str += ","
		}
		str += fmt.Sprintf("%s:%s", query.Name, query.TypeStr)
	}
	return str
}

func (queries *Queries) Set(values string) error {
	for _, value := range strings.Split(values, ",") {
		parts := strings.SplitN(value, ":", 2)
		if len(parts) == 1 {
			*queries = append(*queries, &Query{value, "A", dns.TypeA})
		} else {
			dnsTypeStr := strings.ToLower(parts[1])
			var dnsType uint16
			if dnsTypeStr == "a" {
				dnsType = dns.TypeA
			} else if dnsTypeStr == "mx" {
				dnsType = dns.TypeMX
			} else if dnsTypeStr == "aaaa" {
				dnsType = dns.TypeAAAA
			} else if dnsTypeStr == "ns" {
				dnsType = dns.TypeNS
			} else if dnsTypeStr == "txt" {
				dnsType = dns.TypeTXT
			} else if dnsTypeStr == "cname" {
				dnsType = dns.TypeCNAME
			} else {
				return fmt.Errorf("invalid DNS type %s", dnsTypeStr)
			}
			*queries = append(*queries, &Query{parts[0], strings.ToUpper(dnsTypeStr), dnsType})
		}
	}
	return nil
}

func main() {
	// process flags
	var flagQueries Queries
	flag.Var(&flagQueries, "query", "one or more queries to execute")
	flagServer := flag.String("server", "127.0.0.1:53", "hostname or IP address and port of DNS server to benchmark")
	flagThreads := flag.Int("threads", 8, "number of threads")
	flagSeconds := flag.Int("seconds", 5, "seconds to run benchmark for")
	flag.Parse()

	if len(flagQueries) == 0 {
		fmt.Println("dnsbench: no queries specified; use -h for help")
		fmt.Println("example: dnsbench -query example.com:A -query example.com:MX -server 22.231.113.64:53")
		return
	}

	seconds := time.Duration(*flagSeconds) * time.Second
	endTime := time.Now().Add(seconds)

	// set up rtt/iterations reporting
	type report struct {
		rtt float64
		iterations int
	}
	reportChan := make(chan report)

	// launch threads
	for i := 0; i < *flagThreads; i++ {
		go func(i int) {
			var rttSum time.Duration
			var iterations int
			var errors int
			client := new(dns.Client)
			for time.Now().Before(endTime) {
				randIndex := rand.Intn(len(flagQueries))
				rtt, err := flagQueries[randIndex].Exec(client, *flagServer)
				if err != nil {
					fmt.Printf("query error: %s\n", err)
					errors++
					if errors > iterations * 5 + 5 {
						fmt.Printf("thread %d/%d: too many errors, quitting\n", i, *flagThreads)
						break
					} else {
						continue
					}
				}
				rttSum += rtt
				iterations++
			}
			reportChan <- report{float64(rttSum) / float64(iterations), iterations}
		}(i)
	}

	// collect reports
	var rttAvg float64
	var iterations int
	for i := 0; i < *flagThreads; i++ {
		r := <- reportChan
		rttAvg += r.rtt / float64(*flagThreads)
		iterations += r.iterations
	}
	fmt.Printf("iterations: %d; average rtt: %.2f ms\n", iterations, float64(rttAvg) / float64(time.Millisecond))

}
