package sdk

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	gDNSCache *dnsCache
)

// ---------------------------- inner ----------------------------
func getDNSHost() string {
	return fmt.Sprintf("%s://%s", Conf.RPCProtocal, gDNSCache.getHost())
}

func updateDNSHost() {
	gDNSCache.update()
}

// ---------------------------- dnsCache ----------------------------
type dnsCache struct {
	host           string
	currIP         string
	updateInterval int64
	updatetime     int64
	locker         sync.RWMutex
	updatelock     sync.RWMutex
}

func initDNSCache() {
	gDNSCache = newDNSCache(Conf.XHost, int64(Conf.DNSCacheUpdateInterval))
}

func newDNSCache(host string, interval int64) *dnsCache {
	cache := &dnsCache{
		host:           host,
		updateInterval: interval,
	}
	sdklog.Info("DNSCache start", "host", host, "interval", interval)
	cache.doUpdate(true)
	go cache.flushLoop()
	return cache
}

// while request to the currIP returning error, update dsn cache and reset currIP
func (cache *dnsCache) update() {
	sdklog.Warn("DNSCache update called")
	cache.updatelock.Lock()
	defer cache.updatelock.Unlock()
	if time.Now().Unix()-cache.updatetime <= 1 {
		sdklog.Warn("DNSCache update frequency")
		return
	}
	cache.doUpdate(true)
	cache.updatetime = time.Now().Unix()
}

// -------------------- inner --------------------
func (cache *dnsCache) flushLoop() {
	timer := time.NewTicker(time.Duration(cache.updateInterval) * time.Second)
	for {
		select {
		case <-timer.C:
			cache.doUpdate(false)
		}
	}
}

func (cache *dnsCache) getHost() string {
	cache.locker.RLock()
	defer cache.locker.RUnlock()
	if len(cache.currIP) > 0 {
		return cache.currIP
	}
	sdklog.Warn("DNSCache getHost", "host", cache.host)
	return cache.host
}

func (cache *dnsCache) doUpdate(newIP bool) {
	cache.locker.Lock()
	defer cache.locker.Unlock()
	oldIP := cache.currIP
	defer func() {
		if newIP && oldIP == cache.currIP {
			cache.currIP = ""
		}
	}()
	ns, err := net.LookupHost(cache.host)
	if err != nil || len(ns) == 0 {
		sdklog.Error("doUpdate net.LookupHost()", "err", err.Error(), "PEER_DOMAIN:", cache.host)
		return
	}
	validIPs := make([]string, 0, len(ns))
	for _, ip := range ns {
		if !isIPV4(ip) {
			sdklog.Info("DNSCache not ipv4", "ip", cache.currIP)
			continue
		}
		if newIP && ip == cache.currIP {
			sdklog.Info("DNSCache update excluded", "ip", cache.currIP)
			continue
		}
		validIPs = append(validIPs, ip)
	}
	if len(validIPs) == 0 {
		sdklog.Warn("DNSCache validIPs not found")
		return
	}
	for _, ip := range validIPs {
		if ip == cache.currIP {
			sdklog.Info("DNSCache get same", "ip", cache.currIP)
			return
		}
	}
	cache.currIP = validIPs[getRand()%len(validIPs)]
	sdklog.Info("DNSCache update", "ip", cache.currIP)
}
