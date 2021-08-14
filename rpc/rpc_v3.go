package project

import (
	"fmt"
	"net"
	"net/rpc"
)

const (
	NameSpaceSumService = "SumService"
	NetworkAddr         = "localhost:1234"
)

// 定义请求和响应格式
type SumRequest struct {
	X, Y int
}
type SumResponse struct {
	Z int
}

// 定义支持的方法
type MyServiceInterface = interface {
	Sum(sumRequest SumRequest, sumResponse *SumResponse) error
}

type SumService struct{}

func NewSumService() *SumService {
	return &SumService{}
}

// SumService 实现 MyServiceInterface
func (service *SumService) Sum(req SumRequest, resp *SumResponse) error {
	resp.Z = req.X + req.Y
	return nil
}

// 注册SumService方法
func RegisterSumService(svc MyServiceInterface) {
	rpc.RegisterName(NameSpaceSumService, svc)
}

// 启动server，包含注册，绑定，监听，提供服务 四部
func StartSumService() {
	RegisterSumService(NewSumService())
	listener, err := net.Listen(NetworkTCP, NetworkAddr)
	PanicErrorr(err)
	for {
		conn, err := listener.Accept()
		PanicErrorr(err)
		rpc.ServeConn(conn)
	}
}

// 客户端代码
type SumServiceClient struct {
	*rpc.Client
}

func (c SumServiceClient) Sum(req SumRequest, resp *SumResponse) error {
	return c.Call(NameSpaceSumService+".Sum", req, resp)
}

// 统一构建一次client
func NewSumServiceClient() *SumServiceClient {
	client, err := rpc.Dial(NetworkTCP, NetworkAddr)
	PanicErrorr(err)
	return &SumServiceClient{client}
}
func TestSumRpc() {
	client := NewSumServiceClient()
	req := SumRequest{
		X: 2,
		Y: 7,
	}
	var resp SumResponse
	err := client.Sum(req, &resp)
	PanicErrorr(err)
	fmt.Printf("req:%+v, resp:%+v", req, resp)
}
