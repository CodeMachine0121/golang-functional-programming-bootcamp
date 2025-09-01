# Day 2 - 超越一等公民-高階函數的組合與 Currying 的模擬

## 前言：函數不僅是值，更是積木

在昨天的文章中，我們探討了 Golang 的務實哲學與 pure functional programming 理想之間的張力，並確立了我們務實融合的路線。我的目標是借用 FP 的來增強 Golang，而非取代它。

今天，將深入**函數**。在 Golang 中，我們常說「函數是一等公民」(First-Class Citizens)，這意味著它們可以被存儲在變數中、作為參數傳遞、或作為回傳值，但這個定義僅僅是個起點。
真正的威力，來自於將函數視為可以被轉換、組合與創造的積木。我們將透過**高階函數 (Higher-Order Functions)** 來解鎖這種能力，並學習兩種強大的技術—— **部分應用 (Partial Application)** 與 **柯里化 (Currying)**

## 複習：函數作為一等公民

快速回顧一下「一等公民」的具體含義，這是一切的基礎。

```golang
package main

import "fmt"

// 1. 函數可以被賦值給變數
var myFunc = func(name string) {
    fmt.Printf("Hello, %s\n", name)
}

// 2. 函數可以作為參數傳遞
func greet(greeter func(string)) {
    greeter("World")
}

// 3. 函數可以作為回傳值
func getMultiplier(factor int) func(int) int {
    // 這個回傳的函數是一個閉包
    return func(n int) int {
        return n * factor
    }
}

func main() {
    myFunc("Gopher") // 直接呼叫變數

    greet(myFunc) // 傳遞函數作為參數

    multiplyByTwo := getMultiplier(2) // 接收回傳的函數
    result := multiplyByTwo(5)      // 呼叫它
    fmt.Println(result) // Output: 10
}
```

## 高階函數 (HOFs): 抽象行為的引擎

上面的 `greet` 和 `getMultiplier` 其實都已經是 **高階函數 (Higher-Order Functions, HOFs)** 了，其定義很簡單：

> 一個高階函數，是指一個接收函數作為參數，或回傳一個函數的函數。

HOFs 的核心價值在於抽象化行為，它們讓我們能將「做什麼」與「何時/如何做」分離開來。

思考一下一個常見的需求：計算一段程式碼的執行時間。

```golang
// 一個沒有 HOFs 的做法
func doSomeWork() {
    start := time.Now()
    // --- 核心邏輯開始 ---
    fmt.Println("Doing important work...")
    time.Sleep(1 * time.Second)
    // --- 核心邏輯結束 ---
    duration := time.Since(start)
    fmt.Printf("Work took: %v\n", duration)
}
```

如果有很多函數都需要計時，我們就得在每個函數裡重複 start 和 time.Since 的程式碼。這很繁瑣且容易出錯。
現在，讓我們用 HOF 來抽象這個「計時」的行為：

```golang
// measureTime 是一個 HOF，它接收一個函數 f
func measureTime(f func()) {
    start := time.Now()
    f() // 執行傳入的函數
    duration := time.Since(start)
    fmt.Printf("Execution took: %v\n", duration)
}

func main() {
    work := func() {
        fmt.Println("Doing important work...")
        time.Sleep(1 * time.Second)
    }
    measureTime(work)
}
```

看，我們成功地將計時邏輯與核心工作邏輯分離了，`measureTime` 不關心 `work` 具體做了什麼，它只負責執行並計時。

>  高階函數：它們是建立可重用、可組合行為的基礎。

## 部分應用 (Partial Application)

以上面的例子來說，`getMultiplier` 函數其實已經向我們展示了一種更強大的模式，當我們呼叫 `multiplyByTwo := getMultiplier(2)` 時，我們實際上做了一件有趣的事：我們取了一個需要兩個參數 (factor 和 n) 的邏輯，並預先固定了第一個參數 (factor = 2)，從而得到了一個只需要一個參數 (n) 的新函數 `multiplyByTwo`。

這個過程就叫做部分應用 (Partial Application)，

> 部分應用: 固定一個多參數函數的一部分參數，並產生一個更少參數的新函數的過程

這在 Go 中通常透過閉包 (Closures) 來實現，閉包會「記住」其創建時所在環境的變數，在我們的例子中，回傳的匿名函數就記住了 factor 的值。

## 柯里化 (Currying)

柯里化 (Currying) 是另一個與**部分應用**密切相關但概念上略有不同的技術，它的名字源於數學家 Haskell Curry。

> Currying: 將一個接收多個參數的函數，轉換成一系列只接收單一參數的函數鏈的過程。

讓我們比較一下 `add(a, b)` 這個函數的正常版本和柯里化版本：

```golang
// 正常版本：接收兩個參數
func add(a, b int) int {
    return a + b
}

// 柯里化版本：
// 接收第一個參數 a，回傳一個新函數
// 這個新函數接收第二個參數 b
func curriedAdd(a int) func(int) int {
    return func(b int) int {
        return a + b
    }
}

func main() {
    // 正常呼叫
    sum1 := add(3, 4) // 7

    // 柯里化呼叫
    add3 := curriedAdd(3) // 得到一個「加 3」的函數
    sum2 := add3(4)       // 7
    
    // 也可以這樣鏈式呼叫
    sum3 := curriedAdd(3)(4) // 7
}
```

## Currying vs. Partial Application 的區別

**Currying** 是一種嚴格的轉換，它是將 `f(a, b, c)` 轉換成 `f(a)(b)(c)` 的形式，結果是一連串的單參數函數。

**Partial Application** 是一種應用，它可以一次固定任意數量的參數。例如，將 `f(a, b, c)` 的 `b` 參數固定，得到 `g(a, c)`。

在 Go 這種沒有原生 Currying 語法的語言中，我們通常使用閉包來模擬它，所以兩者在實現上看起來很相似，但理解其意圖上的區別很重要。

> Currying 的目標是為了創造出統一的、易於組合的單參數函數接口。

## 為什麼要這麼麻煩？

現在到了最關鍵的問題：為什麼我們要用 Currying 或 Partial Application 這麼繞的方式來呼叫函數？
答案是：為了組合。

在數學中，函數組合 `(g ∘ f)(x)` 的意思是 `g(f(x))`，但在程式中，這意味著將一個函數的輸出作為另一個函數的輸入，像流水線一樣將它們串聯起來。
柯里化後的函數，因為其「接收一個值，回傳一個值（或下一個函數）」的統一接口，使得組合變得異常簡單和優雅。

讓我們來建立一個通用的 compose 函數：

```golang
// 為了簡單起見，這裡我們處理 func(int) int 類型
// 明天我們將用泛型來讓它變得通用
func compose(f, g func(int) int) func(int) int {
    return func(x int) int {
        return g(f(x)) // 先執行 f，再執行 g
    }
}

func main() {
    // 建立兩個單參數函數
    add5 := curriedAdd(5)
    multiplyBy2 := getMultiplier(2)

    // 現在，我們可以像組合積木一樣組合它們
    // 建立一個新函數：先加 5，再乘以 2
    add5AndMultiplyBy2 := compose(add5, multiplyBy2)

    result := add5AndMultiplyBy2(10) // g(f(10)) => multiplyBy2(add5(10)) => multiplyBy2(15) => 30
    fmt.Println(result)
}
```

看到這種組合的威力了嗎？我們不是寫下 `multiplyBy2(add5(10))` 這種嵌套的、從內到外閱讀的指令式程式碼，而是以一種宣告的方式定義了一個全新的業務邏輯 `add5AndMultiplyBy2`。這個新函數是完全獨立的，可以被傳遞、儲存和重用。

## 總結與明日預告

今天，我們從函數的「一等公民」身份出發，真正開始把它們當作可以操作和組合的積木。

- 高階函數 (HOFs) 是我們的工具，它們透過抽象行為來建立可重用的程式碼。
- 部分應用 (Partial Application) 和 柯里化 (Currying) 是我們準備積木的技術，它們透過固定參數來產生更專用、更易於組合的新函數。
- 函數組合 (Function Composition) 是我們的目標，它讓我們能以聲明式的方式將簡單功能串聯成複雜的業務邏輯。

這個「組合式」的思維模式，是 functional programming 的核心基石。
然而，你可能已經注意到了，我們的 compose 函數還不夠通用，它只能處理 func(int) int。明天，我們將探討 Go 1.18 引入的殺手級特性——泛型 (Generics)，看看它是如何徹底解放 HOFs，讓他能夠建立出真正類型安全、可重用的 functional programming 工具箱。