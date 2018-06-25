package main

import (
	"zmq4"
	"fmt"
)

func main() {
	context, _ := zmq4.NewContext()

	sub, _ := context.NewSocket(zmq4.SUB)

	sub.SetSubscribe("")

	sub.Connect("tcp://localhost:5555")


	for {
		msg, _ := sub.Recv(0)
		fmt.Println(msg)
	}


}
