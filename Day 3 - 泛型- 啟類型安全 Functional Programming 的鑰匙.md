# Day 3 - 泛型——開啟類型安全 Functional Programming 的鑰匙

## 前言：掙脫 interface{} 的枷鎖

昨天，我們體驗了函數組合 (Function Composition) 的優雅。我們建立了一個 compose 函數，成功地將 `add5`和 `multiplyBy2` 兩個小函數黏合成一個全新的業務邏輯。然而，我們也留下了一個巨大的痛點：

```golang
// 昨天的 compose 函數
func compose(f, g func(int) int) func(int) int {
    return func(x int) int {
        return g(f(x))
    }
}
```

這個函數被焊死在了 `func(int) int` 這個型別上。如果我們想組合操作字串或自訂結構的函數，就必須再寫一個 `composeString` 或 `composeMyStruct` 函數，這完全違背了我們追求抽象與可重用性的初衷。

今天，我們將迎來解決這個問題的利器，也是 Golang 相對重要的特性： **泛型 (Generics)**，我們將實驗泛型是如何將我們從 interface{} 的枷鎖中解放出來，讓我們能夠建立出真正類型安全、可重用且優雅的 functional programming 工具箱。

## 回憶：沒有泛型的黑暗時代 (interface{})

在 Go 1.18 之前，當我們想編寫一個能處理「任何類型」的函數時，唯一的選擇就是使用空介面 `interface{}` (在 Go 1.18 後成為 any 的別名)。
讓我們嘗試用 `interface{}` 來編寫一個通用的 Map 函數。Map 函數的作用是接收一個集合和一個函數，並將該函數應用於集合中的每一個元素，最後回傳一個新的集合。

```golang
// 使用 interface{} 的 Map，問題重重
func MapInterface(collection []interface{}, f func(interface{}) interface{}) []interface{} {
    result := make([]interface{}, len(collection))
    for i, item := range collection {
        result[i] = f(item)
    }
    return result
}

func main() {
    numbers := []interface{}{1, 2, 3, 4}
    
    // 我們期望將每個數字加倍
    doubled := MapInterface(numbers, func(item interface{}) interface{} {
        // 1. 危險的類型斷言
        num, ok := item.(int)
        if !ok {
            // 如果集合中混入了非 int 類型，這裡就會出錯
            return 0 
        }
        return num * 2
    })
    
    fmt.Println(doubled) // [2 4 6 8]

    // 2. 從結果中取值時，還需要一次類型斷言
    firstItem := doubled[0].(int)
    fmt.Println(firstItem + 1) // 3
}
```

這個 `MapInterface` 函數雖然能「運作」，但它充滿了問題：

- **喪失類型安全**： 編譯器完全不知道 collection 裡面裝的是什麼。我們可以在 `[]interface{}{1, "hello", 3}` 這樣的集合上呼叫它，編譯器不會報錯，但執行時期 `item.(int)` 就會失敗。
- **繁瑣的類型驗證**： 程式碼中充斥著 `.(int)` 這種運行時類型斷言，這既不美觀，也可能引發 panic。
- **可讀性極差**： 函數簽名 `func(interface{}) interface{}` 幾乎沒有提供任何有用的資訊。它沒有告訴我們輸入和輸出之間應該有什麼樣的關係。

## 曙光：泛型的登場

Go 1.18 引入的泛型，徹底改變了這一切。泛型允許我們在定義函數或類型時使用類型參數 (Type Parameters)。
讓我們用泛型重寫 Map 函數：

```golang
// Go 1.18+
// [A any, B any] 是類型參數列表
// A 是輸入集合的元素類型
// B 是輸出集合的元素類型
func Map[A any, B any](collection []A, f func(A) B) []B {
    result := make([]B, len(collection))
    for i, item := range collection {
        result[i] = f(item)
    }
    return result
}

func main() {
    numbers := []int{1, 2, 3, 4}

    // 我們期望將每個數字加倍
    // A 的類型被推斷為 int, B 的類型也被推斷為 int
    doubled := Map(numbers, func(n int) int {
        // 不需要任何類型斷言！
        return n * 2
    })

    fmt.Println(doubled) // [2 4 6 8]
    fmt.Println(doubled[0] + 1) // 3, doubled 的類型就是 []int

    // 另一個例子：將數字轉換為字串
    // A 的類型被推斷為 int, B 的類型被推斷為 string
    stringified := Map(numbers, func(n int) string {
        return "Number: " + strconv.Itoa(n)
    })
    fmt.Println(stringified) // [Number: 1 Number: 2 Number: 3 Number: 4]
}
```

對比一下，泛型版本的 Map 帶來了壓倒性的優勢：

- **完全的型別安全**： 編譯器知道 collection 是 []A，傳入的函數 f 必須是 func(A) B。所有類型不匹配的錯誤都會在編譯時期被捕獲。
- **無需類型驗證**： 程式碼變得乾淨、安全，且沒有運行時類型檢查的開銷。
- **極佳的可讀性**： 函數簽名 Map[A any, B any]([]A, func(A) B) []B 清晰地描述了函數的意圖：它取一個 A 類型的集合，透過一個從 A 到 B 的轉換函數，產生一個 B 類型的集合。
- **應用泛型**：重鑄我們的 compose 函數

現在，讓我們回到昨天的問題，用泛型來重鑄 compose 函數。一個通用的 compose(f, g) 函數需要處理三種類型：f 的輸入類型，f 的輸出類型（同時也是 g 的輸入類型），以及 g 的輸出類型。

```golang
// 通用的、類型安全的 compose 函數
func Compose[A any, B any, C any](f func(A) B, g func(B) C) func(A) C {
	return func(x A) C {
		// f 的輸出 (B) 正好是 g 的輸入 (B)
		return g(f(x))
	}
}

func main() {
    // 案例 1: int -> int -> int
    add5 := func(n int) int { return n + 5 }
    multiplyBy2 := func(n int) int { return n * 2 }
    add5AndMultiplyBy2 := Compose(add5, multiplyBy2)
    fmt.Println(add5AndMultiplyBy2(10)) // 30

    // 案例 2: int -> string -> string
    toString := func(n int) string { return strconv.Itoa(n) }
    addPrefix := func(s string) string { return "Value: " + s }
    formatNumber := Compose(toString, addPrefix)
    fmt.Println(formatNumber(100)) // "Value: 100"
}
```

看！只用一個 Compose 函數，我們就優雅地處理了完全不同的類型和業務邏輯。這就是泛型為 functional programming 帶來的革命性變化。它讓理論成為了工程實踐。

## 建立我們的泛型 FP 工具箱

有了泛型，我們就可以開始建立一些 functional programming 中最核心、最常用的工具了。

### Filter: 篩選出集合中符合條件的元素

```golang
func Filter[T any](collection []T, predicate func(T) bool) []T {
    var result []T
    for _, item := range collection {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}
```

### Reduce (或 Fold): 將集合中的所有元素「摺疊」或「歸約」成一個單一的值

```golang
// A 是集合元素的類型, B 是累加器的類型
func Reduce[A any, B any](collection []A, initialValue B, accumulator func(B, A) B) B {
    result := initialValue
    for _, item := range collection {
        result = accumulator(result, item)
    }
    return result
}
```

## 綜合實戰：告別 for 迴圈

假設我們有一個需求：計算一個整數陣列中，所有偶數的平方和。

### 傳統的指令式寫法

```golang
func imperativeStyle(numbers []int) int {
    var sum int
    for _, n := range numbers {
        if n%2 == 0 { // 篩選偶數
            square := n * n // 計算平方
            sum += square   // 累加
        }
    }
    return sum
}
```

## 用剛剛建立的 FP 工具來解決它

```golang
func functionalStyle(numbers []int) int {
    // 1. 篩選出所有偶數
    evens := Filter(numbers, func(n int) bool {
        return n % 2 == 0
    })

    // 2. 將所有偶數轉換為它們的平方
    squares := Map(evens, func(n int) int {
        return n * n
    })

    // 3. 將所有平方數累加起來
    sum := Reduce(squares, 0, func(acc, n int) int {
        return acc + n
    })

    return sum
}
```

甚至，我們可以將它們鏈式地寫在一起，這更符合 functional programming 的風格：

```golang
func functionalStyleChained(numbers []int) int {
    return Reduce(
        Map(
            Filter(numbers, func(n int) bool { return n % 2 == 0 }),
            func(n int) int { return n * n },
        ),
        0,
        func(acc, n int) int { return acc + n },
    )
}
```

這種風格的程式碼是聲明式的，它沒有告訴電腦「如何」一步步地去迴圈、判斷、賦值，而是直接描述了「是什麼」——結果是「對篩選過的數字進行映射後再進行歸約」，這種程式碼通常更易於理解和推理。

## 總結與明日預告

今天，我們見證了泛型如何成為 Golang 實踐 functional programming 的關鍵鑰匙，它讓我們擺脫了 interface{} 的類型不安全和繁瑣，能夠建立出真正通用、可重用且類型安全的 Map, Filter, Reduce, Compose 等核心工具。

我們也初步體驗了使用這些工具組合來解決問題的聲明式風格，但我們的工具箱還缺少重要的一環。
在今天的範例中，我們一直在傳遞資料、建立新的 slice，但我們沒有深入探討一個核心問題：**資料本身的可變性 (Mutability)**。

如果 Map 或 Filter 裡面的函數悄悄地修改了原始資料，會發生什麼？這將引導我們進入 functional programming 的下一個核心概念。
明天，我們將深入探討 **不可變性 (Immutability)**，分析它在 Go 中的實現代價與巨大回報，特別是在併發場景下。