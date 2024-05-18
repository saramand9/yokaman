package yokaman_test

import (
	"github.com/saramand9/yokaman"
	"testing"
	"time"
)

func TestMap(t *testing.T) {

	yokacli := yokaman.YoKaManCli()
	//设置测试信息，所有数据针对testid独立
	yokacli.SetTestInfo(1)
	//设置服务器ip
	//yokacli.SetMetricsSvrAddr("172.25.0.1")
	yokacli.SetMetricsSvrAddr("127.0.0.1")
	//启动客户端
	yokacli.Start() //启动yoka上报客户端

	for i := 0; i < 100; i++ {
		req := yokaman.ReqMetrics{
			Trans:    "login",
			Reqtime:  time.Now().UnixMilli(),
			Resptime: time.Now().UnixMilli(),
			Code:     0,
			Robotid:  0,
		}
		yokacli.StatReqMetrics(req)
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second * 10)
	return
}

func Test_WriteCSV(t *testing.T) {
	//l := localstorage()

}
