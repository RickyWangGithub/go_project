package project

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

/*
实现kv存储，切有数据变更的watch监听
包含服务端和客户端
*/
const (
	ServiceName = "KVStorage"
)

// 服务端代码
type SetReq struct {
	Key, Value string
}
type KvStroageInterface interface {
	Get(key string, string2 *string) error
	Set(setReq SetReq, str *string) error
	Watch(key string, keyChanged *string) error
}

func RegisterKVStorageService(svc KvStroageInterface) {
	rpc.RegisterName(ServiceName, svc)
}

type KVStorage struct {
	data   map[string]string
	filter map[string]func(key string)
	lock   sync.Mutex
}

func NewKVStroage() *KVStorage {
	return &KVStorage{
		data:   make(map[string]string),
		filter: make(map[string]func(key string)),
		lock:   sync.Mutex{},
	}
}
func (receiver *KVStorage) Set(setReq SetReq, str *string) error {
	receiver.lock.Lock()
	defer receiver.lock.Unlock()
	fmt.Printf("req:%+v,%s\n", setReq, *str)
	if value := receiver.data[setReq.Key]; value != setReq.Value {
		for id, fn := range receiver.filter {
			fmt.Printf("called:id %s\n", id)
			fn(setReq.Key)
		}
	}
	receiver.data[setReq.Key] = setReq.Value
	*str = strconv.Itoa(len(receiver.data))
	fmt.Printf("req:%+v,%s\n", setReq, *str)
	return nil
}
func (receiver *KVStorage) Get(key string, value *string) error {
	receiver.lock.Lock()
	defer receiver.lock.Unlock()
	*value = receiver.data[key]
	return nil
}
func (receiver *KVStorage) Watch(key string, keyChanged *string) error {
	ch := make(chan string, 10)
	id := fmt.Sprintf("%s_%3d", time.Now().Format(""), rand.Int())
	receiver.lock.Lock()

	receiver.filter[id] = func(key string) {
		if len(ch) == cap(ch) {
			return
		}
		ch <- key
	}
	receiver.lock.Unlock()
	timeout :=  time.After(time.Second*10)
	for {
		select {
		case val := <-ch:
			*keyChanged += val + ","
			fmt.Printf("select val:%s\n",*keyChanged)
		case <-timeout:
			fmt.Printf("timeout\n")
			return nil
		}
	}
	return nil
}

func StartKVStorageServer() {
	RegisterKVStorageService(NewKVStroage())
	//rpc.RegisterName()
	listener, err := net.Listen(NetworkTCP, ":1234")
	PanicErrorr(err)

	for true {
		conn, err := listener.Accept()
		PanicErrorr(err)
		go rpc.ServeConn(conn)
	}
}

// client start
type KVStroageClient struct {
	*rpc.Client
}

func NewKVStroageClient() *KVStroageClient {
	client, err := rpc.Dial(NetworkTCP, "localhost:1234")
	PanicErrorr(err)
	return &KVStroageClient{client}
}
func (c *KVStroageClient) Get(key string, string2 *string) error {
	return c.Client.Call(ServiceName+".Get", key, string2)
}
func (c *KVStroageClient) Set(setReq SetReq, str *string) error {
	return c.Client.Call(ServiceName+".Set", setReq, str)
}
func (c *KVStroageClient) Watch(key string, keyChanged *string) error {
	return c.Client.Call(ServiceName+".Watch", key, keyChanged)

}
func StartClient() {
	client := NewKVStroageClient()
	go func() {
		//for {
			var resp string
			err := client.Watch("", &resp)
			PanicErrorr(err)
			fmt.Printf("keyChanged:%s\n\n", resp)
		//}
	}()
	time.Sleep(time.Second*3)
	k := "key"
	v := "value"
	for i := 0; i < 20; i++ {
		err := client.Get(k, &v)
		PanicErrorr(err)
		fmt.Printf("get resp:%s\n\n", v)
		k += strconv.Itoa(i)
		setReq := SetReq{
			Key:   k,
			Value: strconv.Itoa(i),
		}
		err = client.Set(setReq, &v)
		PanicErrorr(err)
		fmt.Printf("set resp:%s\n\n", v)
		time.Sleep(time.Second * 1)
	}
}
func PanicErrorr(err error) {

	if err != nil {
		fmt.Printf("err:%+v\n", err)
		panic(err)
	}
}
