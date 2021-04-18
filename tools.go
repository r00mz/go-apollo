package apollo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func GetLocalIP() string {
	address, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range address {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}

		}
	}

	return ""
}

func GetNotification(ns []Notification) string {
	bytes, _ := json.Marshal(ns)
	return string(bytes)
}

func HttpGet(url string, timeout int, result interface{}) (int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, result)
	}

	return resp.StatusCode, err
}

func HttpGetWithTransport(url string, timeout int, isRetry bool, result interface{}) (int, error) {
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	tp := &http.Transport{
		MaxIdleConns:        defaultMaxConnes,
		MaxIdleConnsPerHost: defaultMaxConnes,
		DialContext: (&net.Dialer{
			KeepAlive: defaultKeepAliveSecond,
			Timeout:   defaultTimeoutBySecond,
		}).DialContext,
	}
	client.Transport = tp
	var err error
	retry := 0
	var retries = maxRetries
	if !isRetry {
		retries = 1
	}
	for {
		retry++
		if retry > retries {
			break
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return http.StatusServiceUnavailable, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return resp.StatusCode, err
		}
		defer resp.Body.Close()
		if resp == nil {
			time.Sleep(onErrorRetryInterval)
			continue
		}

		switch resp.StatusCode {
		case http.StatusOK:
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return resp.StatusCode, err
			}
			if resp.StatusCode == http.StatusOK {
				err = json.Unmarshal(body, result)
			}
			return resp.StatusCode, nil
		case http.StatusNotModified:
			return resp.StatusCode, nil
		default:
			time.Sleep(onErrorRetryInterval)
			continue
		}
	}
	if retry > retries {
		err = errors.New("over Max Retry Still Error")
	}
	return http.StatusServiceUnavailable, err
}
