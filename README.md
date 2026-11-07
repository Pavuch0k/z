# Z

Язык программирования Z и транслятор `zc`.

Спецификация: https://pavuch0k.github.io/z/

## Назначение

Z — императивный язык со статической типизацией. Транслятор выполняет лексический, синтаксический и семантический анализ и обеспечивает исполнение программы либо построение исполняемого файла.

## Отличительные черты языка

| Конструкция | В языке Z |
| ----------- | --------- |
| Точка входа | `fun entry()` |
| Функции | `fun` |
| Целый тип | `num` |
| Вещественный тип | `real` |
| Строковый тип | `text` |
| Логический тип | `flag` (`yes` / `no`) |
| Вывод | `out(...)` |
| Ввод | `in<тип>()` |
| Цикл с условием | `loop` |
| Цикл по диапазону | `for … from … to …` |
| Строковые операции | только конкатенация `+` |

## Установка

| Платформа | Файл |
| --------- | ---- |
| Linux amd64 | `bin/zc-0.1.0-linux-amd64` |
| Windows amd64 | `bin/zc-0.1.0-windows-amd64.exe` |

## Использование

```bash
./bin/zc-0.1.0-linux-amd64 version
./bin/zc-0.1.0-linux-amd64 run examples/hello.z
./bin/zc-0.1.0-linux-amd64 run examples/grades.z
./bin/zc-0.1.0-linux-amd64 build examples/grades.z -o grades
./grades
```

Windows (Wine):

```bash
wine64 bin/zc-0.1.0-windows-amd64.exe run examples/hello.z
```

## Пример

```z
fun add(a: num, b: num): num {
    return a + b
}

fun entry() {
    out(add(2, 3))
    out("Z" + " language")
    i := 0
    loop i < 3 {
        out(i)
        i = i + 1
    }
}
```

## Сборка транслятора

```bash
./scripts/build-release.sh
./scripts/test.sh
```

## Структура проекта

```text
bin/          сборки транслятора
cmd/zc/       интерфейс командной строки
internal/     этапы трансляции
docs/         спецификация
examples/     примеры программ
scripts/      сборка и испытания
```

## Документация

1. [Описание языка](docs/01-description.md)
2. [Синтаксическая спецификация (РБНФ)](docs/02-syntax.md)
3. [Семантическая спецификация](docs/03-semantics.md)
4. [Схема транслятора](docs/04-translator.md)
5. [Тестовые примеры](docs/05-examples.md)
6. [Архитектура транслятора](docs/06-architecture.md)
7. [Лексический анализ](docs/07-lexical.md)
8. [Система типов](docs/08-types.md)
