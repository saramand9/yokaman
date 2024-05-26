// 用于基本通讯，注册等 轻量请求
package yokaman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Response struct {
	TransId int    `json:"transid"`
	TestId  int    `json:"testid"`
	Name    string `json:"name"`
}

type MetricsCmdCli struct {
	metricaddr string //metric svr 地址
	mu         sync.Mutex
	wg 			sync.WaitGroup
	ch chan struct{}

}

func NewMetricsCmdCli() *MetricsCmdCli {
	m := MetricsCmdCli{}
	m.metricaddr = "127.0.0.1"
	m.ch = make(chan struct{})
	//defer m.conn.Close()
	return &m
}

func (m* MetricsCmdCli) Register(transname string, testid uint32)  (uint8, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	//defer m.wg.Done()
	url := fmt.Sprintf("http://%s:2381/metrics/request", m.metricaddr)

	data := []byte(fmt.Sprintf(`{"name": "%s", "testid": %d}`, transname, testid))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0xFF, err
	}
	defer resp.Body.Close()

	var trans Response
	err = json.NewDecoder(resp.Body).Decode(&trans)
	if err != nil {
		return 0xFF, err
	}

	id := trans.TransId
	//m.ch <- struct{}{}

	return uint8(id), nil

}
func (m *MetricsCmdCli) RegisterRequest(transname string, testid uint32) (uint8, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	//defer m.wg.Done()
	url := fmt.Sprintf("http://%s:2381/metrics/request", m.metricaddr)

	data := []byte(fmt.Sprintf(`{"name": "%s", "testid": %d}`, transname, testid))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0xFF, err
	}
	defer resp.Body.Close()

	var trans Response
	err = json.NewDecoder(resp.Body).Decode(&trans)
	if err != nil {
		return 0xFF, err
	}

	id := trans.TransId
	//m.ch <- struct{}{}

	return uint8(id), nil
}

type StopReq struct {
	TestId int `json:"testid"`
}

type StopResp struct {
	TestId int `json:"testid"`
	code   int `json:"code"`
}

func (m *MetricsCmdCli) StopTest(testid uint32) (uint8, error) {
	url := fmt.Sprintf("http://%s:2381/test/stop", m.metricaddr)
	data := []byte(fmt.Sprintf(`{"nodeid": 0, "testid": %d}`, testid))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0xFF, err
	}

	defer resp.Body.Close()

	var trans StopResp
	err = json.NewDecoder(resp.Body).Decode(&trans)
	if err != nil {
		return 0xFF, err
	}
	fmt.Println(trans)
	id := trans.TestId
	return uint8(id), nil
}
