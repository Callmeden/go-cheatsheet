# Дженерики (Generics)

**Главное:** дженерики позволяют писать функции и типы, которые работают с **разными типами**, сохраняя **строгую типизацию**. Появились в **Go 1.18**. Параметры типа указываются в **квадратных скобках** `[T any]` после имени функции или типа.

## Зачем нужны

**Проблема:** без дженериков — дублирование кода под каждый тип.

```go
func SumInts(a, b int) int { return a + b }
func SumFloats(a, b float64) float64 { return a + b }
// ... и так для каждого типа
```

**Решение:** одна функция с параметром типа.

```go
func Sum[T int | float64](a, b T) T { return a + b }

Sum(3, 4)        // 7
Sum(3.5, 4.2)    // 7.7
```

## Синтаксис

### Параметры типа

```go
func Identity[T any](value T) T {
    return value
}
```

- `T` — **параметр типа**.
- `any` — **constraint** (ограничение): «любой тип».

### Юнион-типы

```go
func Add[T int | float64](a, b T) T {
    return a + b
}
```

- `int | float64` — union: T может быть **int** или **float64**.
- `|` — оператор объединения типов.

### Дженерик-структуры

```go
type Stack[T any] struct {
    elements []T
}

func (s *Stack[T]) Push(value T) {
    s.elements = append(s.elements, value)
}

func (s *Stack[T]) Pop() T {
    if len(s.elements) == 0 {
        var zero T
        return zero
    }
    value := s.elements[len(s.elements)-1]
    s.elements = s.elements[:len(s.elements)-1]
    return value
}

s := Stack[int]{}
s.Push(42)
s.Pop()   // 42
```

**Ключевое:** `Stack[int]` и `Stack[string]` — **разные типы**, но описываются **одним** кодом.

## Встроенные constraints

| Constraint | Что включает |
|---|---|
| **`any`** | Любой тип (синоним `interface{}`) |
| **`comparable`** | Типы с `==` и `!=` (bool, числа, строки, указатели, каналы, массивы/структуры из сравнимых) |

### any

```go
func PrintValue[T any](value T) {
    fmt.Println("Value:", value)
}
```

⚠️ **Избыток `any` снижает типобезопасность.** Если знаешь тип — используй конкретный constraint.

### comparable

```go
func Equal[T comparable](l, r T) bool {
    return l == r
}
```

⚠️ **`comparable` — только как constraint.** Нельзя использовать как тип переменной.

**Где нужен `comparable`:** ключи map, проверки на равенство.

## Пакет `constraints`

`golang.org/x/exp/constraints` — готовые наборы ограничений.

| Constraint | Что включает |
|---|---|
| `constraints.Ordered` | Типы с `<`, `>`, `<=`, `>=` (числа + строки) |
| `constraints.Integer` | Все целые |
| `constraints.Float` | Все с плавающей точкой |

```go
import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](a, b T) T {
    if a > b { return a }
    return b
}

Max(10, 20)     // 20
Max("a", "b")   // b
```

## Тильда `~T`

`~T` — «**тип с базовым типом T**». Включает **не только сам T**, но и **кастомные типы на его основе**.

```go
func PrintIntAliases[T ~int](value T) {
    fmt.Println("Value:", value)
}

type MyInt int
type AnotherInt int

var a MyInt = 42
PrintIntAliases(a)         // ✅ работает

var b AnotherInt = 84
PrintIntAliases(b)         // ✅ работает
```

**Без `~`** — только `int`, а `MyInt` **не подойдёт**.

## `S ~[]E` — дженерик по срезу

Ограничение на **срез элементов** типа `E`.

```go
func SumSlice[S ~[]E, E int | float64](slice S) E {
    var sum E
    for _, v := range slice {
        sum += v
    }
    return sum
}

SumSlice([]int{1, 2, 3, 4})          // 10
SumSlice([]float64{1.1, 2.2, 3.3})   // 6.6
```

- `S ~[]E` — S — срез с элементами `E`.
- `E int | float64` — элементы `int` или `float64`.

## Alias для дженерик-типов (Go 1.24+)

**До Go 1.24** — псевдонимы для дженериков **не работали**:

```go
type NumberList[T ~int | ~float64] = []T   // ❌ ошибка
```

**С Go 1.24** — работают:

```go
type NumberList[T ~int | ~float64] = []T   // ✅

func SumNumbers[T ~int | ~float64](nums NumberList[T]) T {
    var sum T
    for _, n := range nums {
        sum += n
    }
    return sum
}
```

**Когда полезно:** если в проекте часто повторяется один и тот же обобщённый тип — дай ему короткое имя.

## Ограничения дженериков в Go

| Ограничение | Что нельзя |
|---|---|
| **Нет перегрузки функций** | Два `Print` с разными типами — **нельзя**. Используй дженерик |
| **Слабая связь с `reflect`** | Дженерики работают **на этапе компиляции** |
| **Нет variance** | `[]Dog` нельзя передать туда, где ждут `[]Animal` |
| **Нет метаклассов** | Нельзя динамически создать тип |
| **Ограниченная композиция** | Нельзя `T Reader | Writer` — сделай свой интерфейс |

### Про variance

```go
type Animal struct{ Name string }
type Dog struct{ Animal }

func PrintAnimals(animals []Animal) { ... }

dogs := []Dog{{Animal{Name: "Rex"}}}
PrintAnimals(dogs)   // ❌ ошибка компиляции
```

**Почему:** в Go массивы и срезы **строго типизированы** — это защита от ошибок времени выполнения.

### Про `Reader | Writer`

```go
// ❌ нельзя
type ReaderWriter[T Reader | Writer] struct { ... }

// ✅ обход: свой интерфейс
type ReaderWriter interface {
    Reader
    Writer
}

func ProcessStream[T ReaderWriter](stream T) { ... }
```

## Примеры использования

### 1. Filter

```go
func Filter[T any](items []T, predicate func(T) bool) []T {
    var result []T
    for _, item := range items {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}

nums := []int{1, 2, 3, 4, 5}
even := Filter(nums, func(n int) bool { return n%2 == 0 })   // [2 4]

words := []string{"Go", "Rust", "Java"}
startsWithJ := Filter(words, func(w string) bool { return w[0] == 'J' })   // [Java]
```

### 2. Cache

```go
type Cache[K comparable, V any] struct {
    data map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] {
    return &Cache[K, V]{data: make(map[K]V)}
}

func (c *Cache[K, V]) Set(key K, value V) { c.data[key] = value }
func (c *Cache[K, V]) Get(key K) (V, bool) { value, ok := c.data[key]; return value, ok }

cache := NewCache[string, int]()
cache.Set("one", 1)
value, found := cache.Get("one")
```

**Почему `K comparable`:** ключ map **обязан** поддерживать `==`.

### 3. Decorate (декоратор)

```go
func Decorate[T any](fn func(T) T, before func(T), after func(T)) func(T) T {
    return func(input T) T {
        before(input)
        result := fn(input)
        after(result)
        return result
    }
}

addOne := func(n int) int { return n + 1 }
decorated := Decorate(
    addOne,
    func(input int) { fmt.Printf("Before: %d\n", input) },
    func(output int) { fmt.Printf("After: %d\n", output) },
)
decorated(5)   // Before: 5 / After: 6
```

**Обобщённый декоратор** — работает для **любого типа** `T`.

## Подводные камни

### 1. `any` — не «серебряная пуля»

Избыток `any` → **теряешь типизацию**, растут проверки. Если знаешь типы — используй конкретный constraint или union.

### 2. `comparable` — только constraint

```go
var x comparable   // ❌ нельзя как тип переменной
```

Только в `[T comparable]`.

### 3. Без `~` кастомный тип не подойдёт

```go
type MyInt int

func f[T int](v T) { ... }
f(MyInt(42))   // ❌ MyInt — не int (это отдельный тип)

func g[T ~int](v T) { ... }
g(MyInt(42))   // ✅
```

**`~T` = «включая производные типы».**

### 4. `[]Dog` ≠ `[]Animal`

Нет variance. Срезы **инвариантны**. Хочешь обобщённо — делай дженерик-функцию с constraint.

### 5. `T Reader | Writer` — не работает

Нельзя объединять интерфейсы через `|`. Создай **свой** интерфейс-объединение.

### 6. Дженерики **не для всего**

Не оборачивай в дженерик то, что **не нужно обобщать**. Если тип **один** — пиши конкретную функцию. Простота важнее «универсальности ради универсальности».

## Запомнить

- **Дженерики** — с **Go 1.18**. Параметры типа в `[T any]`.
- **`any`** — любой тип. **`comparable`** — типы с `==`/`!=` (нужны для ключей map).
- **Юнион-типы** — `int | float64` через `|`.
- **`~T`** — «тип с базовым T», включает **кастомные типы** на его основе.
- **`S ~[]E`** — дженерик по срезу.
- **`constraints`** из `golang.org/x/exp/constraints`: `Ordered`, `Integer`, `Float`.
- **Alias для дженериков** — с **Go 1.24**.
- **Нет перегрузки функций** — компенсируется дженериками.
- **Нет variance** — `[]Dog` ≠ `[]Animal`.
- **Нет `T Reader | Writer`** — делай свой интерфейс.
- **Дженерики работают на компиляции** — быстрее, чем `interface{}` + type assertion.
- **Не оборачивай всё в дженерики.** Тип один — пиши конкретную функцию.

# Рекурсивные ограничения типов (Go 1.26+)

**Главное:** с **Go 1.26** constraint может **ссылаться на сам обобщённый тип**. Это позволяет описывать контракты вида «тип, который умеет работать с самим собой» — **без** потери типизации и без `any`. Паттерн известен как **F-bounded polymorphism**.

## Что было запрещено до Go 1.26

```go
type T[P T[P]] struct{}   // ❌ invalid recursive type: T refers to itself
```

**Почему запрещено:** constraint не мог **прямо или косвенно** ссылаться на сам обобщённый тип.

**Что изменилось:** теперь такие конструкции **допустимы**. Система дженериков стала выразительнее.

## Классический пример — `Ordered`

```go
type Ordered[T Ordered[T]] interface {
    Less(T) bool
}
```

**Что здесь:**
- `Ordered[T]` — интерфейс, требующий метод `Less(T) bool`.
- Параметр `T` **обязан** удовлетворять `Ordered[T]`.
- Смысл: **«тип T умеет сравнивать себя с таким же T»**.

**Где применить:**

```go
type Tree[T Ordered[T]] struct {
    nodes []T
}

// netip.Addr имеет метод Less(netip.Addr) bool → автоматически подходит
t := Tree[netip.Addr]{}
```

**Ключевое:** constraint ссылается на **параметризованный интерфейс** с тем же типом, который ограничивает.

## Зачем это нужно

| Было (до 1.26) | Стало (1.26+) |
|---|---|
| Только `constraints.Ordered` или внешний comparator | Тип **сам знает**, как себя сравнивать |
| Приведение к `any` | Строгая типизация сохраняется |
| Копипаста под каждый тип | Одна generic-абстракция |

**Практическая ценность:**
- Более строгие и выразительные абстракции.
- Контейнеры и алгоритмы, требующие операций «над самим собой».
- Без потери конкретного типа и без `any`.

## Паттерны использования

### 1. Самотипизированные сравнения

```go
type Comparable[T Comparable[T]] interface {
    Equal(T) bool
}

type Set[T Comparable[T]] struct {
    values []T
}

func (s *Set[T]) Contains(v T) bool {
    for _, x := range s.values {
        if x.Equal(v) {
            return true
        }
    }
    return false
}
```

**Плюс:** `Equal` принимает **тот же самый тип**, а не `any`.

### 2. Fluent API (билдеры)

```go
type Builder[T Builder[T]] interface {
    WithName(string) T
    WithAge(int) T
}

type UserBuilder struct {
    name string
    age  int
}

func (b UserBuilder) WithName(n string) UserBuilder { b.name = n; return b }
func (b UserBuilder) WithAge(a int) UserBuilder     { b.age = a; return b }
```

**Что даёт:** методы возвращают **конкретный тип** `T`, а не `any` или базовый интерфейс. Generic-функции работают с любым fluent-билдером, **не теряя тип**.

**Где особенно полезно:** SQL-билдеры, ORM, конфигураторы.

### 3. Клонируемые типы

```go
type Cloneable[T Cloneable[T]] interface {
    Clone() T
}

func Duplicate[T Cloneable[T]](v T) T {
    return v.Clone()
}
```

`Clone()` гарантированно возвращает **тот же тип** `T`.

**Применение:** immutable-структуры, AST-ноды, конфиги.

### 4. Алгебраические структуры (Monoid, Group)

```go
type Addable[T Addable[T]] interface {
    Add(T) T
    Zero() T
}

func Sum[T Addable[T]](values []T) T {
    if len(values) == 0 {
        var zero T
        return zero.Zero()
    }
    result := values[0]
    for _, v := range values[1:] {
        result = result.Add(v)
    }
    return result
}
```

**Строго типобезопасно** — без потери конкретного типа.

### 5. Узлы деревьев и графов

```go
type Node[T Node[T]] interface {
    Children() []T
}

func Traverse[T Node[T]](n T, visit func(T)) {
    visit(n)
    for _, child := range n.Children() {
        Traverse(child, visit)
    }
}
```

**Применение:** AST, DOM-подобные структуры, конфигурационные деревья.

### 6. Упорядочивание без built-in comparable

```go
type Ordered[T Ordered[T]] interface {
    Less(T) bool
}
```

Позволяет строить **generic-деревья, кучи, skip list, сортировки** — **без внешнего comparator** и **без ограничения на встроенные типы**.

### 7. Self-type для state-machine

```go
type State[T State[T]] interface {
    Next() T
}

func Run[T State[T]](initial T, steps int) T {
    current := initial
    for i := 0; i < steps; i++ {
        current = current.Next()
    }
    return current
}
```

**Идея:** состояние автомата возвращает **новый объект того же типа**.

## Подводные камни

### 1. Go 1.26+

Работает **только с Go 1.26 и выше**. На старых версиях — ошибка компиляции `invalid recursive type`.

### 2. Не путать с наследованием

`Ordered[T Ordered[T]]` — **не наследование**. Это **контракт**: «T реализует метод с сигнатурой, где T — тот же тип». Никакой иерархии типов здесь нет.

### 3. Усложняет чтение

F-bounded polymorphism — **мощный, но сложный** инструмент. Применяй, когда **реально нужно** — например, для библиотек контейнеров. Для обычного кода — **избыточно**.

### 4. Метод должен возвращать/принимать **тот же** тип

В `Builder[T Builder[T]]` метод `WithName(string) T` — **именно T**. Если вернёшь `any` или базовый тип — контракт **не выполнится**.

### 5. Не всё, что «звучит похоже», подходит

«Self-type» — не то же самое, что `Self` в Rust или Python. Это **конкретный паттерн** для дженериков, не «ссылка на свой тип».

## Запомнить

- **Рекурсивные constraints** — с **Go 1.26**. Constraint может ссылаться на сам обобщённый тип.
- До 1.26: `invalid recursive type` — **компилятор запрещал**.
- Паттерн называется **F-bounded polymorphism**.
- Типичная форма: `type X[T X[T]] interface { ... }`.
- Смысл: **«тип T работает с самим собой»** — без `any`, без потери типизации.
- **Применения:** `Ordered`, `Comparable`, `Cloneable`, билдеры, деревья/графы, state-machine, алгебраические структуры.
- **Где особенно полезно:** библиотеки контейнеров, SQL-билдеры, ORM.
- **Не путать с наследованием** — это **контракт**, не иерархия типов.
- **Не для всего.** Сложный инструмент — применять **точечно**.