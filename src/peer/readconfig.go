package main

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strconv"
)

// ConfigData is used to create peer.
type ConfigData struct {
	myIP            *net.TCPAddr
	otherIPs        []*net.TCPAddr
	myIPString      string
	otherIPsString  []string
	torrentFilename string
	seedFilename    string
}

// ReadConfig returns information from configuration file.
func ReadConfig() (data ConfigData, err error) {
	// Open config file.
	file, err := os.Open("config")
	if err != nil {
		return data, err
	}
	// Read config file.
	scanner := bufio.NewScanner(file)
	for {
		if scanner.Scan() {
			// Read my IP.
			if scanner.Text() == "my IP:" {
				scanner.Scan()
				data.myIP, err = net.ResolveTCPAddr("tcp", scanner.Text())
				data.myIPString = scanner.Text()
				if err != nil {
					return data, err
				}
			}
			// Read other IPs.
			if scanner.Text() == "number other IPs:" {
				scanner.Scan()
				n, err := strconv.Atoi(scanner.Text())
				if err != nil {
					return data, err
				}
				scanner.Scan()
				for i := 0; i < n; i++ {
					scanner.Scan()
					otherIP, err := net.ResolveTCPAddr("tcp", scanner.Text())
					if err == nil {
						data.otherIPs = append(data.otherIPs, otherIP)
						data.otherIPsString = append(data.otherIPsString, scanner.Text())
					}
				}
			}
			// Read torrent filename.
			if scanner.Text() == "torrent filename:" {
				scanner.Scan()
				data.torrentFilename = scanner.Text()
			}
			// Read seed filename.
			if scanner.Text() == "seed filename:" {
				scanner.Scan()
				data.seedFilename = scanner.Text()
			}
		} else {
			break
		}
	}
	err = file.Close()
	if err != nil {
		return data, err
	}
	// Check torrent filename.
	if data.torrentFilename == "" {
		return data, errors.New("invalid torrent filename")
	}
	return data, err
}
