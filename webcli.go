// 用于基本通讯，注册等 轻量请求
package yokaman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

/*
type Response struct {
	TransId int    `json:"transid"`
	TestId  int    `json:"testid"`
	Name    string `json:"name"`
}
*/

type WebCli struct {
	webaddr   string //metric svr 地址
	projectid uint
	nodeid    string
	mu        sync.Mutex
	wg        sync.WaitGroup
	ch        chan struct{}
	reportid  uint
}

func NewWebCli() *WebCli {
	m := WebCli{}
	m.webaddr = "127.0.0.1"
	m.ch = make(chan struct{})
	//defer m.conn.Close()
	return &m
}

type StartTestRequest struct {
	NodeID    string   `json:"nodeid" validate:"required"`    //机器人的节点id
	ProjectID uint     `json:"projectid" validate:"required"` //项目id
	CaseName  string   `json:"casename" validate:"required"`  //测试用例名
	RobotNum  uint     `json:"robotnum" validate:"required"`
	TagList   []string `json:"taglist" validate:"required"`   //字符串数组标签列表
	StartTime int64    `json:"starttime" validate:"required"` //毫秒 UTC
}

type ResponseNew struct {
	Msg  string `json:"msg"`
	Data struct {
		ReportID  int `json:"reportid"`
		NewReport int `json:"new_report"`
	} `json:"data"`
	Cd int `json:"cd"`
}

type StopTestRequest struct {
	NodeID    string `json:"nodeid" validate:"required"`    //机器人的节点id
	ProjectID uint   `json:"projectid" validate:"required"` //项目id
	ReportID  uint   `json:"reportid" validate:"required"`  //项目id
}

type StopTestResponse struct {
	Msg  string `json:"msg"`
	Data struct {
		ProjectID uint `json:"projectid"`
		ReportID  uint `json:"reportid"`
		Status    int  `json:"status"`
	} `json:"data"`
	Cd int `json:"cd"`
}

func (m *WebCli) StartTest(projectID uint, NodeID string) (int, error) {
	m.projectid = projectID
	m.nodeid = NodeID

	m.mu.Lock()
	defer m.mu.Unlock()
	//defer m.wg.Done()
	url := fmt.Sprintf("http://%s:23366/api/v1/test", m.webaddr)

	data := StartTestRequest{
		NodeID:    NodeID,
		ProjectID: projectID,
		CaseName:  "登入",
		RobotNum:  10000,
		TagList:   make([]string, 0),
		StartTime: time.Now().UnixMilli(),
	}
	jsonData, err := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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
	var trans ResponseNew
	err = json.NewDecoder(resp.Body).Decode(&trans)
	if err != nil {
		fmt.Println("Decode faild")
		return 0xFF, err
	}
	fmt.Printf("%+v\n", trans)
	id := trans.Data.ReportID
	m.reportid = uint(trans.Data.ReportID)
	return id, nil
}

/*
type StopReq struct {
	TestId int `json:"testid"`
}

type StopResp struct {
	TestId int `json:"testid"`
	code   int `json:"code"`
}*/

func (m *WebCli) StopTest( /*testid uint32*/ ) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	//defer m.wg.Done()
	url := fmt.Sprintf("http://%s:23366/api/v1/test/1/stop", m.webaddr)

	type StopTestRequest struct {
		NodeID    string `json:"nodeid" validate:"required"`    //机器人的节点id
		ProjectID uint   `json:"projectid" validate:"required"` //项目id
		ReportID  uint   `json:"reportid" validate:"required"`  //项目id
	}

	data := StopTestRequest{
		NodeID:    m.nodeid,
		ProjectID: m.projectid,
		ReportID:  m.reportid,
	}

	fmt.Printf("%+v\n", data)

	jsonData, err := json.Marshal(data)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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
	var trans StopTestResponse
	err = json.NewDecoder(resp.Body).Decode(&trans)
	if err != nil {
		fmt.Println("Decode faild")
		return 0xFF, err
	}
	fmt.Printf("%+v\n", trans)
	status := trans.Data.Status
	return status, nil
}
