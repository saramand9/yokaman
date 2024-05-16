// 用于基本通讯，注册等 轻量请求
package service

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

// 如果
func Register(transname string, testid uint32) (uint8, error) {
	url := "http://127.0.0.1:2381/metrics/request"

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
