package sdk

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var (
	gTransport *http.Transport = &http.Transport{
		Dial: dialTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives:     false,
		MaxIdleConns:          200,
		IdleConnTimeout:       120 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
	}

	gHTTPClient *http.Client = &http.Client{
		Transport: gTransport,
	}
)

// ------------------------------ inner ------------------------------
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second*15)
}

// ------------------------------ http cli ------------------------------
func httpGetWithLongConn(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("httpGet req get error: %s", err.Error())
	}
	resp, err := gHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpGet client do error: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("httpGet io read error: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		sdklog.Error("httpGetWithLongConn HttpReqError, op: get", "url", url, "statusCode", resp.StatusCode)
		return nil, fmt.Errorf("httpGet status code not 200, url(%s) status code %d", url, resp.StatusCode)
	}
	if body == nil || len(body) == 0 {
		return nil, fmt.Errorf("httpGet resp body nil")
	}
	sdklog.Info("httpGetWithLongConn success", "url", url, "resp", string(body))
	return body, nil
}

func httpPostWithLongConn(url string, host string, contentType string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("httpPost req post error: %s", err.Error())
	}
	if host != "" {
		req.Host = host
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("User-Agent", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11")
	resp, err := gHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpPost client do error: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HttpsPost io read error: %s, body[%d]:%s", err.Error(), len(body), string(body))
	}
	if resp.StatusCode != 200 {
		sdklog.Error("httpPostWithLongConn HttpReqError, op: post", "url", url, "statusCode", resp.StatusCode)
		return nil, fmt.Errorf("HttpsPost status code not 200, url(%s) body(%s) status code %d", url, data, resp.StatusCode)
	}
	if body == nil || len(body) == 0 {
		return nil, fmt.Errorf("HttpsPost resp body nil")
	}
	sdklog.Info("httpPostWithLongConn success", "url", url, "resp", string(body))
	return body, nil
}

func httpGet(url string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("httpGet req get error: %s", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpGet client do error: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("httpGet io read error: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		sdklog.Error("httpGet HttpReqError, op: get", "url", url, "statusCode", resp.StatusCode)
		return nil, fmt.Errorf("httpGet status code not 200, url(%s) status code %d", url, resp.StatusCode)
	}
	if body == nil || len(body) == 0 {
		return nil, fmt.Errorf("httpGet resp body nil")
	}
	sdklog.Info("httpGet success", "url", url, "resp", string(body))
	return body, nil
}

func httpPost(url string, host string, contentType string, data []byte) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("httpPost req post error: %s", err.Error())
	}
	if host != "" {
		req.Host = host
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("User-Agent", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpPost client do error: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("httpPost io read error: %s, body[%d]:%s", err.Error(), len(body), string(body))
	}
	if resp.StatusCode != 200 {
		sdklog.Error("httpPost HttpReqError, op: post", "url", url, "statusCode", resp.StatusCode)
		return nil, fmt.Errorf("httpPost status code not 200, url(%s) body(%s) status code %d", url, data, resp.StatusCode)
	}
	if body == nil || len(body) == 0 {
		return nil, fmt.Errorf("httpPost resp body nil")
	}
	sdklog.Info("httpPost success", "url", url, "resp", string(body))
	return body, nil
}
