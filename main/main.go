package main

import (
	"fmt"
	"geerpc"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
)

func startServer(addr chan string) {
	var foo Foo
	if err  := geerpc.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalln("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
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

func main() {
	//log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	client, _ := geerpc.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{i,i}
			var reply int
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error: ", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}


func reflectDemo() {
	var wg sync.WaitGroup
	typ := reflect.TypeOf(&wg)
	fmt.Printf("typ nummethod: %v\n", typ.NumMethod())
	for i := 0; i < typ.NumMethod();i++ {
		method := typ.Method(i)
		argv := make([]string,0, method.Type.NumIn())
		returns := make([]string,0, method.Type.NumOut())
		for j := 1;j < method.Type.NumIn();j++ {
			argv = append(argv, method.Type.In(j).Name())
		}
		for j := 0; j < method.Type.NumOut();j++ {
			returns = append(returns, method.Type.Out(j).Name())
		}
		log.Printf("func(w *%s) %s(%s) %s",
				typ.Elem().Name(),
				method.Name,
				strings.Join(argv, ","),
				strings.Join(returns, ","),
			)
	}
}