// 用于基本通讯，注册等 轻量请求
package yokaman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	TransId int    `json:"transid"`
	TestId  int    `json:"testid"`
	Name    string `json:"name"`
}

type MetricsCmdCli struct {
	//conn         net.Conn      //tcp链接句柄
	metricaddr string //metric svr 地址
	/*buf          *bytes.Buffer //发送用的字节缓冲区
	lastPkgIn    time.Time     //最近一次收到待发送业务包的时间戳，当时间超过MaxDelay立即发送
	stackedPkg   uint32        //堆积的业务包量
	totalSendPkg uint32        //已发送的数据包
	//IntervalSendPkg uint32 //已发送的数据包
	metrics2send chan RecodeMetrics //待发送缓冲区*/
}

func NewMetricsCmdCli() *MetricsCmdCli {
	m := MetricsCmdCli{}
	m.metricaddr = "127.0.0.1"
	//defer m.conn.Close()
	return &m
}

// 如果
func (m *MetricsCmdCli) RegisterRequest(transname string, testid uint32) (uint8, error) {
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
	return uint8(id), nil
}
