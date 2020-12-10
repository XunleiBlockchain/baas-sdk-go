package sdk

import (
	"math/rand"
	"net"
	"strings"
	"time"
)

func getRand() int {
	b := make([]byte, 2)
	_, err := rand.Read(b)
	if err != nil {
		return 0
	}
	return (int(b[0]) << 8) + int(b[1])
}

const alphabeta string = "0123456789abcdefghijklmnopqrstuvwxyz"

func getRandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	buf := make([]byte, n)
	rand.Read(buf[:])
	rc := byte(0)
	for k, v := range buf {
		rc = v & 0x23
		buf[k] = alphabeta[rc]
	}
	return string(buf)
}

// is isIPV4 or not
func isIPV4(add string) bool {
	ip := net.ParseIP(add)
	return ip != nil && strings.Contains(add, ".")
}

func strAND(s string) string {
	return "&" + s
}

func catchInterfacePanic() {
	if r := recover(); r != nil {
		sdklog.Error("catchInterfacePanic: %v", r)
	}
}

func checkStrRes(x interface{}, defaultRes string) string {
	res, ok := x.(string)
	if ok {
		return res
	}
	return defaultRes
}
