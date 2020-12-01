package main

import (
	"context"
	"geerpc"
	"geerpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func startHttpServer(addr chan string) {
	var foo Foo
	if err := geerpc.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}

	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalln("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	//geerpc.Accept(l)
	geerpc.HandleHTTP()
	_ = http.Serve(l, nil)
}

func startServer(addr chan string) {
	var foo Foo
	server := geerpc.NewServer()
	if err := server.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalln("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	server.Accept(l)
}

type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServersDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	client := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{i, i}
			foo(client, context.Background(), "call", "Foo.Sum", &args)
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServersDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcaset", "Foo.Sum", &Args{i, i * i})
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{i, i * i})
		}(i)
	}
	wg.Wait()
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}

}
func main() {
	log.SetFlags(0b1111111111)
	addrCh1 := make(chan string)
	addrCh2 := make(chan string)
	go startServer(addrCh1)
	go startServer(addrCh2)
	addr1 := <-addrCh1
	addr2 := <-addrCh2
	time.Sleep(time.Second * 2)
	call(addr1, addr2)
	broadcast(addr1, addr2)
}
