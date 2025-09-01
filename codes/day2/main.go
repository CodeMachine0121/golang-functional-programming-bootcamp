package main

import (
	"fmt"
	"time"
)

var myFunc = func(name string) {
	fmt.Printf("Hello, %s\n", name)
}

func greet(greeter func(string)) {
	greeter("World")
}

func getMultiplier(fector int) func(int) int {
	return func(n int) int {
		return fector * n
	}
}

// Non-High-Order Function
func doSomeWork() {
	start := time.Now()
	// do some business logic
	fmt.Println("Do business logic")
	time.Sleep(1 * time.Second)
	// done business logic

	duration := time.Since(start)
	fmt.Printf("Work took: %v\n", duration)
}

// Hight-Order Function
func measureTime(f func()) {

	start := time.Now()
	// do some business logic
	f()
	// done business logic
	duration := time.Since(start)
	fmt.Printf("Work took: %v\n", duration)
}

// non-Currying
func add (a, b int) int {
	return a + b
}
// Currying
func curriedAdd(a int) func(int) int {
	return func(b int) int {
		return a+b
	}
}

// compose
func compose (f, g func(int) int) func(int) int {
	return func(x int) int {
		return g(f(x))
	}
}


func main() {
	myFunc("Gopher")

	greet(myFunc)

	result := getMultiplier(2)(5)
	fmt.Println(result)

	// use non-high-order function
	doSomeWork()
	// use high-order function
	measureTime(func() {
		fmt.Println("Do business logic with High-Order Function")
		time.Sleep(1 * time.Second)
	})

	// 正常呼叫
    sum1 := add(3, 4) // 7
	fmt.Printf("sum1: %d\n", sum1)

    // 柯里化呼叫
    add3 := curriedAdd(3) // 得到一個「加 3」的函數
    sum2 := add3(4)       // 7
	fmt.Printf("sum2: %d\n", sum2)
    
    // 也可以這樣鏈式呼叫
    sum3 := curriedAdd(3)(4) // 7
	fmt.Printf("sum3: %d\n", sum3)

	// Compose
	 // 建立兩個單參數函數
    add5 := curriedAdd(5)
    multiplyBy2 := getMultiplier(2)

    // 現在，我們可以像組合積木一樣組合它們
    // 建立一個新函數：先加 5，再乘以 2
    add5AndMultiplyBy2 := compose(add5, multiplyBy2)

    composeResult := add5AndMultiplyBy2(10) // g(f(10)) => multiplyBy2(add5(10)) => multiplyBy2(15) => 30
    fmt.Printf("Compose Result: %d\n", composeResult)
}
