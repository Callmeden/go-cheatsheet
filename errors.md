# Обработка ошибок в Go

**Главное:** в Go ошибки — это **обычные значения**, а не исключения. Любой тип с методом `Error() string` — это ошибка. Функции **возвращают ошибку вместе с результатом**, а вызывающий **явно проверяет** её. Это делает поток управления **прозрачным и предсказуемым**.

## error — встроенный интерфейс

```go
type error interface {
    Error() string
}
```

Любой тип с методом `Error() string` — **уже ошибка**. Никакого `implements`, как с любым интерфейсом в Go.

## Философия: ошибки, а не паники

| Принцип | Суть |
|---|---|
| Ошибки — обычные значения | Возвращаются вместе с результатом |
| Минимизация паник | `panic` — только для **необратимых** ситуаций |
| Явная обработка | Каждый вызов — либо обработка, либо **осознанный** игнор |

**Почему ошибки, а не исключения:**

| Причина | Смысл |
|---|---|
| Прозрачность | Видно в сигнатуре функции |
| Локализация | Обработка там, где возникла |
| Производительность | Дешевле, чем создание стека исключения |
| Предсказуемость | Код явно показывает, что может пойти не так |

## Создание ошибок

| Способ | Когда использовать |
|---|---|
| `errors.New("...")` | Простая ошибка без форматирования |
| `fmt.Errorf("...")` | С форматированием (`%d`, `%s`) |
| `fmt.Errorf("...: %w", err)` | **Оборачивание** — сохранить оригинал |
| Свой тип с `Error()` | Нужны **поля** (Code, Message) |

### errors.New

```go
var ErrNotFound = errors.New("item not found")
```

**Sentinel errors** — глобальные переменные, чтобы сравнивать через `errors.Is`.

### fmt.Errorf

```go
return fmt.Errorf("cannot divide %d by zero", a)
```

### Свой тип ошибки

```go
type MyError struct {
    Code    int
    Message string
}

func (e MyError) Error() string {
    return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func doSomething() error {
    return MyError{Code: 404, Message: "resource not found"}
}
```

**Когда:** нужны **данные** внутри ошибки (код, поля, детали).

## Оборачивание ошибок (wrapping)

```go
func openConfig() error {
    err := readFile("config.yaml")
    if err != nil {
        return fmt.Errorf("openConfig: %w", err)
    }
    return nil
}
```

**`%w` vs `%v`:**

| Формат | Что делает | errors.Is работает? |
|---|---|---|
| `%w` | **Оборачивает** — сохраняет оригинал | ✅ Да |
| `%v` | Просто подставляет текст | ❌ Нет |

⚠️ **Разница принципиальная.** Хочешь, чтобы `errors.Is` нашёл оригинал — **только `%w`**.

## Сравнение ошибок

### errors.Is — «это та ошибка?»

```go
if errors.Is(err, ErrNotFound) {
    fmt.Println("Item not found")
}
```

**Как работает:**
1. Сравнивает `err == target`.
2. Если нет — вызывает `Unwrap()` и повторяет.
3. Идёт **по всей цепочке**, пока не найдёт или не закончится.

⚠️ **НЕ используй `==`** для обёрнутых ошибок — вернёт `false`, потому что `fmt.Errorf` создал **новую** ошибку.

### errors.As — «вытащи ошибку типа T»

```go
var myErr *MyError
if errors.As(err, &myErr) {
    fmt.Println("Code:", myErr.Code)
}
```

**Когда:** нужно **достать данные** из ошибки конкретного типа.

⚠️ Второй аргумент — **указатель на переменную** (`&myErr`), чтобы Go мог записать туда найденную ошибку.

### errors.Unwrap — один уровень

```go
orig := os.ErrPermission
wrapped := fmt.Errorf("wrap1: %w", orig)
unwrapped := errors.Unwrap(wrapped)   // orig
```

**Разница:**
- `errors.Is` / `errors.As` — проходят **всю цепочку**.
- `errors.Unwrap` — **один уровень**.

### Свой метод Is (продвинутый)

```go
type HttpError struct {
    Code int
    Msg  string
}

func (e *HttpError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.Code, e.Msg)
}

func (e *HttpError) Is(target error) bool {
    t, ok := target.(*HttpError)
    if !ok {
        return false
    }
    return e.Code == t.Code
}
```

Теперь любая ошибка с `Code == 404` считается «равной» `Err404`. **errors.Is вызовет твой `Is`** и определит равенство по-своему.

## errors.AsType (Go 1.26+)

```go
// Было (go 1.13+)
var myErr *MyError
if errors.As(err, &myErr) { ... }

// Стало (go 1.26+)
if merr, ok := errors.AsType[*MyError](err); ok {
    fmt.Println(merr.Code)
}
```

| Плюс AsType | Что даёт |
|---|---|
| Type-safe | Проверка на этапе **компиляции** |
| Без рефлексии | **Быстрее** |
| Без паники | В рантайме не рванёт |
| Меньше аллокаций | Экономия памяти |
| Короче код | Без `var myErr *T` |

⚠️ **`T` и `*T` — разные типы.** Если ошибка возвращается как `*MyError` — проси `[*MyError]`, не `[MyError]`.

**Рекомендация из курса:** с Go 1.26 — **переходить на `AsType`** вместо `As`.

## errors.Join (Go 1.20+)

```go
err1 := errors.New("database disconnected")
err2 := errors.New("failed to close file")
err := errors.Join(err1, err2)
```

**Что даёт:**
- Все ошибки в одной.
- `errors.Is` / `errors.As` работают **по всем** вложенным.

**Правила:**
- `errors.Join(nil, err2)` — `nil` игнорируется.
- `Join(err)` — вернёт ошибку **как есть**.
- `Join()` без аргументов — вернёт `nil`.

**Где применять:**
- `defer a.Close()` + `defer b.Close()` — оба могут вернуть ошибку.
- Параллельная обработка — несколько горутин.
- Основная ошибка + уточнения (parse + validation).

## panic, recover, defer

### panic — только для необратимого

```go
func unsafeDivision(a, b int) int {
    if b == 0 {
        panic("division by zero")
    }
    return a / b
}
```

**Когда:** нарушение инвариантов, невозможность продолжения.

### recover — ловля паники

```go
func safeDivision(a, b int) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered:", r)
        }
    }()
    result := a / b
    fmt.Println("Result:", result)
}
```

**recover работает ТОЛЬКО внутри `defer`.** В обычном коде — вернёт `nil`.

### defer — освобождение ресурсов

```go
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()   // выполнится ВСЕГДА: при return, panic, ошибке
```

**Порядок срабатывания defer — LIFO (обратный):**

```go
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
// Вывод: 3, 2, 1
```

**Что защищает defer:** закрытие файлов, освобождение мьютексов, соединений с БД.

## Подводные камни

### 1. nil-интерфейс с nil-значением внутри

```go
func returnNil() error {
    var err *MyError = nil
    return err   // ← вернёт error, который НЕ nil
}

if err := returnNil(); err != nil {
    fmt.Println("Error is not nil")   // ← выполнится!
}
```

**Почему:** интерфейс `error` = пара `(тип, значение)`. Тип = `*MyError` (задан), значение = nil. Интерфейс **не пустой** → `err != nil`.

**Как проверить:** `errors.Is`, `errors.As`, или явная проверка типа.

**Это тот же баг**, что в теме интерфейсов — **nil-указатель внутри интерфейса**.

### 2. `==` вместо `errors.Is`

```go
if err == ErrNotFound { ... }   // ❌ не сработает для обёрнутых
if errors.Is(err, ErrNotFound) { ... }   // ✅ правильно
```

### 3. `%v` вместо `%w`

```go
return fmt.Errorf("context: %v", err)   // ❌ оригинал потерян
return fmt.Errorf("context: %w", err)   // ✅ цепочка сохранена
```

### 4. `recover` вне `defer`

```go
if r := recover(); r != nil { ... }   // ❌ всегда nil
```

### 5. panic для обычных ошибок

Не «файл не найден», не «неверный пароль». **Только** для необратимого.

## Лучшие практики

| № | Практика |
|---|---|
| 1 | Обрабатывай ошибку **сразу**, не откладывай |
| 2 | Сравнивай через `errors.Is` / `errors.As` (или `AsType` с Go 1.26) |
| 3 | **Добавляй контекст** через `fmt.Errorf("...: %w", err)` |
| 4 | `panic` — только для необратимого |
| 5 | Логируй **критические** ошибки |
| 6 | Документируй возвращаемые ошибки в комментариях |
| 7 | `defer` для освобождения ресурсов **всегда** |

## Запомнить

- `error` — **встроенный интерфейс** с методом `Error() string`.
- Ошибки — **значения**, не исключения. Возвращаются **вместе с результатом**.
- **`%w`** — оборачивание с сохранением оригинала. **`%v`** — просто текст.
- **`errors.Is`** — «есть ли в цепочке эта ошибка». **`errors.As`** — «вытащи тип».
- **`errors.Unwrap`** — один уровень. `Is`/`As` — всю цепочку.
- **`errors.AsType[E](err)`** (Go 1.26+) — type-safe замена `As`, быстрее, без рефлексии.
- **`errors.Join`** (Go 1.20+) — склеивает несколько ошибок.
- **`panic`** — только для необратимого. **`recover`** — только внутри **`defer`**.
- **`defer`** — выполняется **всегда** (return, panic). Порядок — **LIFO**.
- **Sentinel errors** (`var ErrX = errors.New(...)`) — сравнивай через `errors.Is`.
- **Свой тип ошибки** — когда нужны поля внутри.
- **Свой `Is(target error)`** — если хочешь кастомную логику равенства.
- **⚠️ nil-указатель внутри `error` ≠ nil.** Интерфейс смотрит на **пару (тип, значение)**.