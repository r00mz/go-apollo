package apollo

import (
	"encoding/json"
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
