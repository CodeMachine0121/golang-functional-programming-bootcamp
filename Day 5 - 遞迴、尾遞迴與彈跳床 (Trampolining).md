# Day 5: 遞迴、尾遞迴與彈跳床 (Trampolining)

## 前言：Functional Programming 中的迴圈

在過去的幾天裡，我們建立了 functional programming 的幾個核心：**高階函數**、**泛型**以及**不可變性**。同時也學會了如何使用 `Map`、`Filter`、`Reduce` 等工具來處理集合，以一種聲明式的方式取代了傳統的 for 迴圈。

然而，在 functional programming 的世界裡，還有一種更基礎、更數學化的方式來表達重複性的操作，那就是**遞迴 (Recursion)**，對於許多 FP 語言來說，遞迴不僅僅是一種演算法技巧，它就是迴圈的同義詞。
今天，我們將探討遞迴的 FP 觀點，直面它在 Golang 中的缺陷——**堆疊溢位 (Stack Overflow)**，解釋為什麼 Golang 缺乏 FP 語言中常見的**尾遞迴優化 (Tail Call Optimization)**，並學習一種巧妙的模式——**彈跳床 (Trampolining)**——來安全地在 Go 中實現深度遞迴。

## 遞迴：用自身來定義問題

遞迴的核心思想是將一個問題分解成一個規模更小的、結構相同的子問題，直到達到一個已知的基礎情況 (Base Case)。

階乘 (Factorial) 是最經典的例子：

> Factorial(n) 等於 n * Factorial(n-1)

基礎情況：Factorial(0) 等於 1
在 Go 中的實現直觀明瞭：

```golang
func Factorial(n uint64) uint64 {
    // 基礎情況 (Base Case)
    if n == 0 {
        return 1
    }
    // 遞迴步驟 (Recursive Step)
    return n * Factorial(n-1)
}
```

從 FP 的角度看，這種寫法非常優雅，它沒有像 for 迴圈那樣引入一個可變的計數器 i 和一個可變的累加器 result。整個函數是一個純粹的表達式，其結果完全由輸入 n 決定。

### 隱藏的危險：呼叫堆疊 (The Call Stack)

這種優雅是有代價的。每次我們呼叫一個函數時，程式都需要在一個名為**呼叫堆疊 (Call Stack)** 的記憶體區域中儲存一些資訊，比如函數的參數、局部變數以及返回地址。當函數執行完畢後，這些資訊會從堆疊中彈出。

對於 `Factorial(5)`，堆疊的變化過程大致如下：

- main 呼叫 Factorial(5)，Factorial(5) 的資訊入棧。
- Factorial(5) 呼叫 Factorial(4)，Factorial(4) 的資訊入棧。
- Factorial(4) 呼叫 Factorial(3)，Factorial(3) 的資訊入棧。
- ... 直到 Factorial(0)。
- Factorial(0) 回傳 1，Factorial(0) 出棧。
- Factorial(1) 拿到結果，計算 1 * 1 並回傳，Factorial(1) 出棧。
- ... 堆疊逐層解開，直到 main 拿到最終結果。

呼叫堆疊的空間是有限的。如果遞迴的深度太深，就會耗盡所有堆疊空間，導致程式崩潰，這就是堆疊溢位 (Stack Overflow)。

```golang
func main() {
    // 在大部分系統上，這個數字足以導致堆疊溢位
    // panic: runtime error: stack overflow
    fmt.Println(Factorial(100000)) 
}
```

## FP 的解決方案：尾遞迴優化 (TCO)

許多 functional programming 語言（如 Scheme, OCaml, Haskell）的編譯器能夠解決這個問題，它們實現了一種名為**尾遞迴優化 (Tail Call Optimization, TCO)** 的技術。

> 尾呼叫 (Tail Call) 是指一個函數返回前的最後一個動作是呼叫另一個函數（或它自身）。

讓我們將階乘改寫成尾遞迴的形式：

```golang
// acc 是累加器 (accumulator)
func FactorialTail(n uint64, acc uint64) uint64 {
    // 基礎情況
    if n == 0 {
        return acc
    }
    // 遞迴呼叫是最後的動作
    return FactorialTail(n-1, n*acc)
}
```

在 `FactorialTail` 中，`return FactorialTail(...)` 是函數的最後一步。它不需要在遞迴呼叫返回後再做任何計算（比如 `n * ...`）。

具備 TCO 的編譯器能識別出這種模式，它不會創建一個新的堆疊幀 (stack frame)，而是直接**重用當前的堆疊幀**。這樣一來，無論遞迴多少次，堆疊空間的佔用都是固定的 O(1)。尾遞迴實際上被編譯成了一個高效的 `goto` 或迴圈。

**然而，Go 語言的編譯器出於設計上的考量，明確地不支持尾遞迴優化。** 主要原因是 TCO 會讓程式的堆疊追蹤 (stack trace) 變得不準確，這違背了 Go 追求簡潔、易於偵錯的哲學。

## Go 的務實 workaround：彈跳床 (Trampolining)

既然 Go 不會幫我們優化堆疊，我們就必須自己動手。**彈跳床 (Trampolining)** 是一種將深度遞迴從堆疊 (stack) 轉移到堆 (heap) 上執行的模式。

> 核心思想是： **不要直接進行遞迴呼叫，而是回傳一個代表「下一步計算」的函數 (Thunk)**。

然後，我們在一個簡單的迴圈中不斷執行這些 Thunk，直到計算完成，這個迴圈就像一個「彈跳床」，一次次地「彈起」並執行下一個計算步驟。

讓我們用這個模式來重寫一個安全的深度加總函數：

### **1. 定義我們的 Trampoline 結構**

我們需要一個能代表「計算的下一步」的結構，一個泛型函數類型是完美的選擇，它要麼回傳最終結果，要麼回傳下一個要執行的步驟。

```golang
// Trampoline 是一個函數，它執行一小步計算。
// 它回傳一個結果（如果計算完成）和下一個 Trampoline（如果還沒完成）。
type Trampoline[T any] func() (*T, Trampoline[T])
```

###2. 重寫遞迴函數以回傳 Trampoline
code
Go
// sumRecursive 不再直接遞迴，而是回傳一個描述下一步的 Trampoline
func sumRecursive[T constraints.Integer | constraints.Float](start, end T, acc T) Trampoline[T] {
	if start > end {
		// 基礎情況：計算完成，回傳最終結果和一個 nil 的 Trampoline
		return func() (*T, Trampoline[T]) {
			return &acc, nil
		}
	}
	// 遞迴步驟：回傳一個新的 Trampoline，它封裝了下一步的計算
	return func() (*T, Trampoline[T]) {
		return nil, sumRecursive(start+1, end, acc+start)
	}
}
3. 建立執行迴圈（彈跳床本體）
code
Go
func Run[T any](t Trampoline[T]) *T {
	for t != nil {
		result, nextT := t()
		if nextT == nil { // 計算完成
			return result
		}
		t = nextT // 準備執行下一步
	}
	return nil
}

func main() {
    // 建立初始的 Trampoline
	initialTrampoline := sumRecursive(1, 200000, 0)
    
    // 把它放進執行器裡執行
	sum := Run(initialTrampoline)

	fmt.Println(*sum) // 安全地計算出結果，不會堆疊溢位
}
透過這個模式，我們將原本會深嵌在呼叫堆疊中的遞迴鏈，轉換成了一個在堆上創建的、扁平的函數鏈。驅動迴圈 Run 在一個單一的堆疊幀中執行，有效地將堆疊空間 O(N) 的問題轉化為了堆空間 O(N) 和時間 O(N) 的問題，從而避免了堆疊溢位。
這值得嗎？
彈跳床模式無疑比直接的遞迴或一個簡單的 for 迴圈要複雜得多。它引入了閉包和額外的函數呼叫開銷。
那麼，什麼時候應該考慮使用它？
演算法的自然表達： 當一個問題（比如遍歷樹、解析巢狀結構）的遞迴定義遠比其迭代版本要清晰和自然時。
無法預測的遞迴深度： 當你在處理外部輸入，無法保證其遞迴深度在安全範圍內時。
堅持 Functional Style： 當你希望在整個程式碼庫中保持一種純粹的、無副作用的 functional programming 風格時。
對於 Go 中的絕大多數日常問題，一個清晰的 for 迴圈通常是更簡單、更高效、更符合 Go 語言習慣 (idiomatic) 的選擇。彈跳床是你工具箱裡一件用於特殊場景的精密儀器，而不是一把日常使用的錘子。
總結與明日預告
今天，我們深入了遞迴這個 functional programming 的核心概念。我們理解了它在 Go 中的堆疊溢位風險，了解了 Go 為何缺乏尾遞迴優化，並親手實現了彈跳床模式作為一個安全的替代方案。
我們現在已經掌握了 functional programming 的幾個關鍵武器：函數組合、泛型、不可變性以及安全的遞迴模式。然而，我們一直以來都在迴避 Go 程式設計中最常見、最無處不在的一個問題：錯誤處理。
明天，我們將把目光聚焦在 if err != nil 上。我們將分析它背後的設計哲學，並探討如何借鑑 functional programming 的思想，從 error 作為一個普通的值，演進到使用更強大的類型（如 Either 單子）來處理錯誤，從而寫出更流暢、更具組合性的錯誤處理程式碼。
