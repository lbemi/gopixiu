package main

import "fmt"

func main() {
	var ch chan int
	fmt.Printf("var: the type of ch is %T \n", ch)
	fmt.Printf("var: the val of ch is %v \n", ch)
	if ch == nil {
		// 也可以用make声明一个channel，它返回的值是一个内存地址
		ch = make(chan int)
		fmt.Printf("make: the type of ch is %T \n", ch)
		fmt.Printf("make: the val of ch is %v \n", ch)
	}

	ch2 := make(chan string, 10)
	fmt.Printf("make: the type of ch2 is %T \n", ch2)
	fmt.Printf("make: the val of ch2 is %v \n", ch2)
}
