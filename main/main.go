package main

import (
	"context"
	"geerpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func startServer(addr chan string) {
	var foo Foo
	if err  := geerpc.Register(&foo); err != nil {
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


type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(args Args, reply  *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func call(addCh chan string) {
	client, _ := geerpc.DialHTTP("tcp", <-addCh)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{i,i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error: ", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
func main() {
	//log.SetFlags(0b1111111111)
	addr := make(chan string)
	go call(addr)
	startServer(addr)
}