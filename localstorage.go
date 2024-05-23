package yokaman

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	//"os/types"
)

const MaxWriteLatency = 1000 //100毫秒
const MaxLatencyWrite = 128  //最大延迟写入数量

const metricsdir = "./metricsdata"

var handler *Localstorage
var storage_once sync.Once

type Localstorage struct {
	filepath      string
	testid        int32
	file          *os.File
	writer        *csv.Writer
	metrics2write chan ReqMetrics
	lastPkgIn     time.Time //最近一次收到待发送业务包的时间戳，当时间超过MaxDelay立即发送
	stackedPkg    uint32    //堆积的业务包量
}

func NewStorage(testid int32) *Localstorage {
	storage_once.Do(func() {
		handler = &Localstorage{
			testid: testid,
		}

		var err error
		// 创建文件夹
		dir, err := os.Getwd()
		path := filepath.Join(dir, metricsdir)
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		handler.file, err = os.Create(fmt.Sprintf("%s/%d_metrics_%d.csv",
			metricsdir, testid, time.Now().UnixMilli()))
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}
		handler.writer = csv.NewWriter(handler.file)
		handler.metrics2write = make(chan ReqMetrics, 1024*1024)
		handler.lastPkgIn = time.Unix(32500886400, 0) //默认设置为3000年
	})
	return handler
}

func (l *Localstorage) Close() {
	err := l.file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (l *Localstorage) WriteHeader() {
	// 写入CSV数据
	defer l.writer.Flush()

	header := []string{"Start", "End", "Request", "Code"}
	err := l.writer.Write(header)
	if err != nil {
		log.Fatalf("failed writing record: %s", err)
	}

	log.Printf("backup metrics created successfully")
}

// 客户端循环发包，要注意buf 和socket 是否是线程安全的 。 待测
func (l *Localstorage) ThreadWrite( /*metrics2send chan []NetReqMetrics */ ) {
	//
	go func() {
		for {
			select {
			case v, ok := <-l.metrics2write:
				if !ok {
					fmt.Println("write message channel closed")
					return
				}
				l.Write(v)
				break
			default:
				if time.Since(l.lastPkgIn) > time.Duration(time.Millisecond*MaxWriteLatency) {
					l.writer.Flush()
					//fmt.Printf("【%v】 over time, flush csv, num = %d\n", time.Now().Format("2006-01-02 15:04:05.00"), l.stackedPkg)
					l.reset()
				} else {
					time.Sleep(time.Millisecond) //长期无数据，释放cpu
				}
				break
			}
		}
	}()
}

func (l *Localstorage) reset() {
	l.stackedPkg = 0
	l.lastPkgIn = time.Unix(32500886400, 0) //默认设置一个极大的年份
}

func (l *Localstorage) Write(metrics ReqMetrics) {
	l.stackedPkg++

	err := l.writer.Write([]string{
		strconv.FormatFloat(float64(metrics.Reqtime), 'f', -1, 64),
		strconv.FormatFloat(float64(metrics.Resptime), 'f', -1, 64),

		//strconv.FormatInt(int64(metrics.Resptime), 10),
		metrics.Trans,
		strconv.FormatInt(int64(metrics.Code), 10)})

	if err != nil {
		log.Fatalf("failed writing record: %s", err)
	}

	//fmt.Printf("【%v】 write metrics, num = %d\n", time.Now().Format("2006-01-02 15:04:05.00"), l.stackedPkg)

	if l.stackedPkg >= MaxLatencyWrite {
		l.writer.Flush()
		l.reset()
	}

	if l.stackedPkg == 1 {
		l.lastPkgIn = time.Now()
	}

}
func (l *Localstorage) WriteMetris(metrics ReqMetrics) {
	// 写入CSV数据
	l.metrics2write <- metrics
}
