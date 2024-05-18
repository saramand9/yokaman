// 用于数据通讯，性能指标等高性能通讯

package yokaman

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"time"
)

// 最大延迟发送时间
const MaxLatency = 100 //100毫秒

const MTU = 1440

// 最大延迟发送包量
var MaxLatencyPkg uint32

func init() {
	MaxLatencyPkg = uint32(math.Floor(float64(MTU / GetNetStructSize(RecodeMetrics{}))))
}

// 待发送队列最大值
const MAX_CHAN_SIZE = (1024 * 1024)

type MetricsNetCli struct {
	conn         net.Conn      //tcp链接句柄
	metricaddr   string        //metric svr 地址
	buf          *bytes.Buffer //发送用的字节缓冲区
	lastPkgIn    time.Time     //最近一次收到待发送业务包的时间戳，当时间超过MaxDelay立即发送
	stackedPkg   uint32        //堆积的业务包量
	totalSendPkg uint32        //已发送的数据包
	//IntervalSendPkg uint32 //已发送的数据包
	metrics2send chan RecodeMetrics //待发送缓冲区
}

func NewMetricsNetCli() *MetricsNetCli {
	m := MetricsNetCli{}
	m.metricaddr = "127.0.0.1"
	m.lastPkgIn = time.Unix(32500886400, 0) //默认设置为3000年，仅当收到第一个待发送业务包时，开始计时
	m.metrics2send = make(chan RecodeMetrics, MAX_CHAN_SIZE)
	m.buf = new(bytes.Buffer)
	m.stackedPkg = 0
	//defer m.conn.Close()
	return &m
}

func (m *MetricsNetCli) ConnectSvr() error {
	dialer := net.Dialer{}
	c, err := dialer.Dial("tcp", fmt.Sprintf("%s:2380", m.metricaddr))
	if err != nil {
		fmt.Printf("connect metrics svr (%s) failed\n", m.metricaddr)
		return nil
	}
	m.conn = c
	return nil
}

func GetNetStructSize(data interface{}) int {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return len(buf.Bytes())
}

func (m *MetricsNetCli) reset() {
	m.stackedPkg = 0
	m.lastPkgIn = time.Unix(32500886400, 0) //默认设置一个极大的年份
}

func (m *MetricsNetCli) realSend() {
	_, err := m.conn.Write(m.buf.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	m.buf.Reset()
}
func (m *MetricsNetCli) sendPkg(v RecodeMetrics) {
	err := binary.Write(m.buf, binary.LittleEndian, v)
	if err != nil {
		fmt.Println(err)
	}
	m.stackedPkg++
	m.totalSendPkg++
	if m.stackedPkg == 1 {
		m.lastPkgIn = time.Now()
	}
	if m.stackedPkg >= MaxLatencyPkg {
		m.realSend()
		m.reset()
	}
}

// 客户端循环发包，要注意buf 和socket 是否是线程安全的 。 待测
func (m *MetricsNetCli) ThreadSend( /*metrics2send chan []RequestMetrics */ ) {
	//
	go func() {
		for {
			select {
			case v, ok := <-m.metrics2send:
				if !ok {
					fmt.Println("net message channel closed")
					return
				}
				m.sendPkg(v)
				break
			default:
				if time.Since(m.lastPkgIn) > time.Duration(time.Millisecond*MaxLatency) {
					m.realSend()
					m.reset()
				}
				break
			}
		}
	}()
}

func (m *MetricsNetCli) UploadStatics(metrics RecodeMetrics) {
	m.metrics2send <- metrics
}

func (m *MetricsNetCli) Exit() {
	err := m.conn.Close()
	if err != nil {
		fmt.Println("%s connect err, %s", m.conn.RemoteAddr().String(), err)
		return
	}
}
