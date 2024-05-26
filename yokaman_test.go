package yokaman_test

import (
	"fmt"
	"github.com/saramand9/yokaman"
	"testing"
	"time"
)

func TestMap(t *testing.T) {

	yokacli := yokaman.YoKaManCli()
	yokacli.SetTestInfo(200000, 10)        //设置testid, 所有数据是按照testid来区分计算的
	yokacli.SetMetricsSvrAddr("127.0.0.1") //设置服务器ip

	err := yokacli.Start() //启动数据上报客户端，在后台会启动线程上传
	if err != nil {
		fmt.Sprintf("start yoka client faild err %s", err)
		t.Fail()
		return
	}

	for i := 0; i < 100; i++ {
		req := yokaman.ReqMetrics{
			Trans:    "login",
			Reqtime:  time.Now().UnixMilli(),
			Resptime: time.Now().UnixMilli() + 1000,
			Code:     yokaman.SUCCESS,
			Robotid:  0,
		}
		yokacli.StatReqMetrics(req)
		time.Sleep(time.Second)
	}
	yokacli.Stop()
	return
}

func Test_WriteCSV(t *testing.T) {
	//l := localstorage()

}
