package Utils

import (
	"log"
	"math"
	"net"
)

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func StringArrayContains(arr []string, element string) bool {
	for _, a := range arr {
		if a == element {
			return true
		}
	}
	return false
}

func GetMinFromStringIntMap(m map[string]int) string {
	minInt := math.MaxInt64
	minS := ""
	for k, v := range m {
		if v < minInt {
			minInt = v
			minS = k
		}
	}
	return minS
}
