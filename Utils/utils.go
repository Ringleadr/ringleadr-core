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

func MapStringArrayStringEquals(a, b map[string][]string) bool {
	for k, v := range a {
		if v1, ok := b[k]; !ok || !StringArrayEquals(v, v1) {
			return false
		}
	}
	for k, v := range b {
		if v1, ok := a[k]; !ok || !StringArrayEquals(v, v1) {
			return false
		}
	}
	return true
}

func StringArrayEquals(a, b []string) bool {
	for _, i := range a {
		if !StringArrayContains(b, i) {
			return false
		}
	}
	for _, i := range b {
		if !StringArrayContains(a, i) {
			return false
		}
	}
	return true
}
