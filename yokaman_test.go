package yokaman_test

import (
	"github.com/saramand9/yokaman"
	"testing"
)

func TestMap(t *testing.T) {

	yokacli := yokaman.YoKaManCli()
	//设置测试信息，所有数据针对testid独立
	yokacli.SetTestInfo(1)
	//设置服务器ip
	yokacli.SetMetricsSvrAddr("172.25.0.1")
	//启动客户端
	yokacli.Start() //启动yoka上报客户端

	yokacli.StatReqMetrics(yokacli.ReqMetrics{})

	return
}
