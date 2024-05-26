package yokaman

// 协议序号
//const ProtoStart = 1 //註冊這個項目
const ProtoEnd = 2

const ProtoRequestMetrics = 10 //上傳Request數據
const ProtoTransMetrics = 11   //上傳事務數據

type Protohead struct {
	Protocolid int8 //协议id
	Len        int8 //
}

type NetReqMetrics struct {
	Transid int8 //后面扩展
	Start   int64
	End     int64 //微秒
	Code    int8
	//Count   int16
}

type NetMetrics struct {
	Protohead
	Testid int32 //测试id
	//Nodeid int8 //机器人id
	NetReqMetrics
}

type StartMetrics struct {
	Protohead
	prjId  [64]byte
	testId [10]byte
	nodeId [10]byte
	//测试id
	Nodeid int8 //机器人id
}
