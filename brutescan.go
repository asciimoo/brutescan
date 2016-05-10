package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sort"
	"sync"
	"syscall"
	"time"
)

type PortState int

const (
	Open PortState = iota
	Closed
	Timeout
)

func init() {
	flag.Usage = usage
}

func main() {

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
		return
	}

	openFileLimit := uint(rLimit.Cur) - 24

	minPort := flag.Uint("pmin", 1, "lowest target port")
	maxPort := flag.Uint("pmax", 65535, "highest target port")
	timeout := flag.Uint("timeout", 3000, "connection timeout in milliseconds")
	verbose := flag.Bool("verbose", false, "set verbose mode on")
	poolSize := flag.Uint("pool", openFileLimit, "concurrent pool size")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		return
	}

	if *poolSize == 0 || *poolSize > openFileLimit {
		log.Fatal("error: invalid pool size")
	}

	if *maxPort-*minPort < 1 {
		log.Fatal("error: invalid port range")
	}

	host := flag.Arg(0)

	ip, err := net.ResolveIPAddr("ip", host)

	if err != nil {
		log.Fatal("host resolv error:", err)
	}

	target := ip.String()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	fmt.Printf("Scanning %v (%v)\n", host, target)
	fmt.Printf("Port range: %v-%v (%v ports)\n", *minPort, *maxPort, *maxPort-*minPort)
	fmt.Printf("Concurrent connections: %v\n\n", *poolSize)

	scanStartTime := time.Now()
	openPorts := scanPorts(target, *minPort, *maxPort, *poolSize, *timeout)
	elapsedTime := time.Since(scanStartTime)

	if len(openPorts) == 0 {
		fmt.Println("no open ports found")
	} else {
		sort.Ints(openPorts)
		for _, port := range openPorts {
			fmt.Printf("port %-5d open\n", port)
		}
	}
	fmt.Println("\nScan finished in", elapsedTime)
}

func scanPorts(ip string, minPort, maxPort, poolSize, timeout uint) []int {
	results := make([]int, 0, 16)
	portsNum := maxPort - minPort + 1

	ports := make(chan uint, portsNum)

	for i := uint(0); i < portsNum; i++ {
		ports <- minPort + i
	}

	wg := sync.WaitGroup{}
	wg.Add(int(portsNum))

	for i := uint(0); i < poolSize; i++ {
		go func(pool chan uint) {
			for port := range pool {
				switch checkPortState(ip, port, timeout) {
				case Timeout:
					log.Println("Warning: port", port, "timeoutted")
				case Open:
					log.Println("found open port:", port)
					// dirty hack - int has built in sorting support
					results = append(results, int(port))
				}
				wg.Done()
			}
		}(ports)
	}
	wg.Wait()
	close(ports)
	return results
}

func checkPortState(ip string, port uint, timeout uint) PortState {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", ip, port), time.Duration(timeout)*time.Millisecond)
	if err != nil {
		switch t := err.(type) {
		case *net.OpError:
			switch t.Op {
			case "dial":
				if t.Timeout() {
					return Timeout
				}
				if t.Temporary() {
					// TODO maybe retry?
					fmt.Println("tmp error:", t.Err)
				}
			default:
				fmt.Println("unknown error:", t.Op)
			}
		}
		return Closed
	}
	conn.Close()
	return Open
}

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] host/ip\n", os.Args[0])
	flag.PrintDefaults()
}
