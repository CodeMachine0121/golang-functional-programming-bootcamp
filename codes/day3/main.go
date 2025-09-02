package main

import "fmt"

func Map[A any, B any](collection []A, f func(n A) B) []B {

	result := make([]B, len(collection))

	for i, item := range collection {
		result[i] = f(item)
	}
	return result
}

func Filter[T any](collection []T, predicate func(T) bool) []T {

	var result []T

	for _, item := range collection {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

func Reduce[A any, B any](collection []A, initialValue B, accumulator func(B, A) B) B {
	result := initialValue

	for _, item := range collection {
		result = accumulator(result, item)
	}

	return result
}

// 計算一個整數陣列中，所有偶數的平方和。

func imperaticeStyle(numbers []int) int {
	var sum int

	for _, n := range numbers {
		if n%2 == 0 {
			square := n * n
			sum += square
		}

	}
	return sum
}

func functionalStyle(numbers []int) int {
	// 1. 篩選出所有偶數
	evens := Filter(numbers, func(n int) bool {
		return n%2 == 0
	})

	//2. 將所有偶數轉為平方
	squares := Map(evens, func(n int) int {
		return n * n
	})

	// 3. 將所有平方數累加起來
	sum := Reduce(squares, 0, func(acc, n int) int {
		return acc + n
	})

	return sum
}

func main() {
	collection := []int{1,2,3,4,5,6,7,8,9}
	fmt.Printf("使用傳統作法: %d\n", imperaticeStyle(collection))
	fmt.Printf("使用functional 作法: %d\n", functionalStyle(collection))
}