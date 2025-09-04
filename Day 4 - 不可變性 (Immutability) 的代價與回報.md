# Day 4: 不可變性 (Immutability) 的代價與回報

## 前言：看不見的敵人——隱藏的狀態變更

在昨天的文章中，我們利用泛型打造了類型安全的 Map, Filter, Reduce 等強大的 functional programming 工具。
我們體驗了將簡單函數組合起來，以聲明式的方式解決問題的快感，但是，還有一個地雷在我們的程式碼中。思考一下這個情境：

```golang
func main() {
    users := []User{
        {ID: 1, Name: "Alice", Role: "admin"},
        {ID: 2, Name: "Bob", Role: "user"},
    }

    // Map 函數的目的是回傳一個新的 []string
    // 但如果傳入的函數做了“壞事”呢？
    names := Map(users, func(u User) string {
        if u.Role == "admin" {
            u.Name = u.Name + " (Admin)" // 對 user 做後天改值的動作
        }
        return u.Name
    })
    
    fmt.Println("Names:", names)
    // 原始的 users slice 發生了什麼？
    fmt.Println("Original Users:", users) 
}
```

`users` 會被修改嗎？ 這取決於 User 的定義以及 Go 的傳遞語義。如果 User 是一個包含指標的複雜結構，或者我們傳遞的是 []*User，那麼原始資料就很可能在我們意想不到的地方被汙染了。
這就是**可變狀態 (Mutable State)** 帶來的混亂。當程式中任何地方都可以隨意修改一個共享的資料時，程式的行為就變得極難預測。

今天，我們就來探討 functional programming 的核心概念之一：**不可變性 (Immutability)**，來探討在 Golang 中，追求不可變性需要付出什麼代價，又能收穫怎樣巨大的回報。
什麼是不可變性？

## 不可變性的定義

> 一個物件或資料結構，一旦被創建，其狀態就不能再被改變。

如果我們需要「修改」一個不可變的物件，我們不能在原地修改它。相反，我們必須創建一個新的物件，這個新物件包含了我們想要的變更，而原始物件則保持原封不動。
讓我們用一個簡單的例子對比一下：

```golang
// 可變 (Mutable) 的做法
type MutablePoint struct {
    X, Y int
}
func (p *MutablePoint) Move(dx, dy int) {
    p.X += dx // 直接修改內部狀態
    p.Y += dy
}

// 不可變 (Immutable) 的做法
type ImmutablePoint struct {
    X, Y int
}
// 注意：Move 回傳了一個新的 ImmutablePoint
func (p ImmutablePoint) Move(dx, dy int) ImmutablePoint {
    return ImmutablePoint{ // 創建並回傳一個新實例
        X: p.X + dx,
        Y: p.Y + dy,
    }
}

func main() {
    // 可變
    p1 := &MutablePoint{X: 1, Y: 1}
    p1.Move(2, 2)
    fmt.Println(p1) // &{3 3} - p1 本身被改變了

    // 不可變
    p2 := ImmutablePoint{X: 1, Y: 1}
    p3 := p2.Move(2, 2)
    fmt.Println(p2) // {1 1} - p2 保持不變
    fmt.Println(p3) // {3 3} - p3 是一個全新的點
}
```

## 不可變性的巨大回報

為什麼 functional programming 如此推崇不可變性？因為它能帶來的好處是根本性的：

- **可預測性 (Predictability)**: 這是最直接的好處。當你將一個不可變的物件傳遞給一個函數時，你完全不必擔心這個函數會悄悄地修改你的物件。函數的行為變得像數學一樣純粹，它的輸出只依賴於它的輸入，這種特性被稱為**引用透明性 (Referential Transparency)**，這消除了程式中最常見的一類 Bug。
- **併發安全(Concurrency Safety)**: 這是在 Golang中，不可變性最有價值的地方。如果資料不能被修改，那麼它就可以在任意數量的 Goroutines 之間被自由地共享，而完全不需要使用 Mutex 或其他任何鎖機制！ 資料競爭 (Data Race) 的根本原因就是「多個執行緒同時對一塊共享記憶體進行讀寫」。不可變性直接消除了「寫」這個操作，從而根絕了資料競爭。
- **簡化偵錯與推理 (Easier Reasoning & Debugging)**: 當一個 Bug 出現時，如果你的狀態是可變的，你需要追蹤這個物件的整個生命週期，找出到底是哪一步錯誤地修改了它的狀態。而對於不可變的物件，你只需要關心它被創建時的狀態，因為你知道它永遠不會變。這大大降低了心智負擔。

## Golang 的挑戰：不可變性的代價

既然不可變性這麼好，為什麼 Golang 本身不強制執行呢？ 因為 Golang 的設計哲學是務實的，而不可變性在 Golang 中是有成本的。

- **性能開銷**： 這是最大的挑戰。在上面的 `ImmutablePoint` 例子中，每次 Move 操作都會在堆疊上（或堆上）分配一個新的 struct。如果我們處理的是一個包含一百萬個元素的大 slice，為了修改其中一個元素而完整地複製整個 slice，將會帶來巨大的記憶體分配和 GC 壓力。
- **語法慣例**： Go 的語法和標準庫在設計上更傾向於可變操作。例如，內建的 `append` 函數可能會在 slice 容量足夠時原地修改，也可能會分配一個新的底層陣列。`json.Unmarshal` 會直接修改傳入的結構指標，這些都是為了效率而做的務實選擇。

### Slice 和 Map 的陷阱

Go 中的 Slice 和 Map 都是引用類型（更準確地說，是包含指標的結構）。這意味著當你將一個 slice 賦值給新變數時，你只是複製了指向底層陣列的指標，而不是資料本身。

```golang
s1 := []int{1, 2, 3}
s2 := s1 // s2 和 s1 指向同一個底層陣列
s2[0] = 99
fmt.Println(s1) // [99 2 3] - s1 也被修改了！
```

## 在 Go 中實踐不可變性的策略

我們的目標不是盲目地讓所有東西都不可變，而是有策略地在最需要的地方應用它。

- **Mindset**： 團隊內部約定，某些核心的資料類型（如 Order, Event）應該被視為不可變的，任何需要修改的函數，都必須回傳一個新的實例。
- **善用值傳遞**： 對於體積不大的 struct，優先使用**值傳遞**而不是指標傳遞。這樣每次傳遞都會隱式地創建一個副本，保證了原始資料的安全。
- **防禦性複製 (Defensive Copying)**: 在處理 slice 和 map 時，這是最關鍵的技巧。在函數的入口和出口處，手動創建資料的副本。

```golang
// 正確地複製一個 slice
func cloneSlice[T any](original []T) []T {
    newSlice := make([]T, len(original))
    copy(newSlice, original)
    return newSlice
}
```

### 透過 API 設計強制不可變

將 struct 的欄位設為非導出（小寫字母開頭），只提供導出的 Getter 方法來訪問資料。
任何修改操作都透過一個回傳新實例的方法來完成。

```golang
type ImmutableUser struct {
    name string // 非導出
    age  int
}

func NewUser(name string, age int) ImmutableUser {
    return ImmutableUser{name: name, age: age}
}

func (u ImmutableUser) Name() string {
    return u.name
}

// SetAge 不會修改 u，而是回傳一個新的 ImmutableUser
func (u ImmutableUser) SetAge(newAge int) ImmutableUser {
    return ImmutableUser{
        name: u.name,
        age:  newAge,
    }
}
```

### 探索持久性資料結構 (Persistent Data Structures) 

對於性能極其敏感的場景，可以研究 **PDS**，這是一種聰明的資料結構，當你「修改」它時，它會最大限度地重用舊結構的記憶體，只為變更的部分創建新節點。這大大降低了複製成本，雖然 Go 標準庫沒有內建 PDS，但有一些第三方套件可供使用。

## 總結與明日預告

今天，我們探討了 functional programming 的核心原則——不可變性。

我們看到了它在提升程式碼可預測性和併發安全方面的巨大威力，同時也直面了它在 Golang 中帶來的性能挑戰和實現複雜性，結論依然是**務實**，但我們應該將不可變性視為我們的一個主動技能，在最合適的地方——核心業務邏輯、跨 Goroutine 共享的資料模型——精準地使用它，以換取程式整體的健壯性。

我們現在有了函數組合、泛型工具以及對純粹性和不可變性的理解，但還有一種常見的 functional programming 模式我們沒有觸及：**遞迴 (Recursion)**。
在很多 FP 語言中，遞迴是迴圈的替代品，然而，Golang 對遞迴的支援有其自身的限制。

明天，我們將探討如何在 Go 中安全地使用遞迴，理解其 Stack Overflow 的風險，並學習一種名為 **Trampolining** 的技術來克服這個限制。