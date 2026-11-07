# Тестовые примеры языка Z

## 1. Корректная программа

См. также `examples/all.z`, `examples/grades.z`.

```z
fun greet(name: text): text {
    return "Hello, " + name
}

fun abs(x: num): num {
    if x < 0 {
        return -x
    } else {
        return x
    }
}

fun entry() {
    a: num = 10
    b := 3
    out(a + b)
    out("Z" + " language")
    out(greet("Nikita"))

    nums: []num = [1, 2, 3]
    out(nums[0])

    i := 0
    loop i < 3 {
        out(i)
        i = i + 1
    }

    for k from 1 to 3 {
        out(k)
    }

    out(abs(-5))
}
```

## 2. Ошибка: унарный минус для text

```z
fun entry() {
    s := "abc"
    out(-s)
}
```

Ожидается семантическая ошибка. Файл: `examples/bad_string.z`.

## 3. Ошибка: умножение text

```z
fun entry() {
    s := "abc"
    out(s * 2)
}
```

Файл: `examples/bad_mul_string.z`.

## 4. Ошибка типов при присваивании

```z
fun entry() {
    age: num = 5
    age = "five"
}
```

## 5. Ошибка условия

```z
fun entry() {
    n := 10
    if n {
        out(n)
    }
}
```

Условие `if` должно иметь тип `flag`.
