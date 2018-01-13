dnsbench
========

dnsbench reports average RTT for a given query and DNS server.

Compile:

	go build dnsbench.go

Usage:

	Usage of ./dnsbench:
	-query value
		one or more queries to execute
	-seconds int
		seconds to run benchmark for (default 5)
	-server string
		hostname or IP address and port of DNS server to benchmark (default "127.0.0.1:53")
	-threads int
		number of threads (default 8)

Example:

	$ ./dnsbench -server 8.8.8.8:53 -query example.com:A
	iterations: 967; average rtt: 29.22 ms
