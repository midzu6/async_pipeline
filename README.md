# Async Pipeline

**Решение домашнего задания по Go: конкурентный пайплайн обработки писем (Spammer)**

## Оригинальное задание

Полное описание задачи можно посмотреть здесь:  
**[https://gitlab.com/vk-golang/lectures/-/blob/master/02_async/99_hw/spammer/README.md?ref_type=heads](https://gitlab.com/vk-golang/lectures/-/blob/master/02_async/99_hw/spammer/README.md?ref_type=heads)**

## Описание задачи

Нужно реализовать аналог Unix-пайплайна на Go с использованием каналов:

```bash
cat emails.txt | SelectUsers | SelectMessages | CheckSpam | CombineResults
```

Каждая функция-стадия получает данные через chan interface{} и передаёт результат в следующий канал.
Главная цель — максимальная параллельность при соблюдении жёстких ограничений:

GetUser() — можно вызывать параллельно (1 секунда на запрос)

GetMessages() — поддерживает батчинг до 2 пользователей за вызов

HasSpam() — не больше 5 одновременных запросов (антибрут)

Никаких data race, deadlock’ов и двойных close

## Что реализовано
RunPipeline() — запускает все стадии конкурентно и закрывает каналы

SelectUsers() — дедупликация пользователей учитывает alias’ы

SelectMessages() — батчинг по 2 пользователя + добавил семафор (буфер 30)

CheckSpam() — ограничение параллелизма (семафор на 5)

CombineResults() — сбор всех результатов, сортировка (спам сверху + по ID)

## Пройденные тесты
Все тесты проходят (включая -race)

## Запуск
Bashgo test -v -race

Репозиторий создан как демонстрация решения задания по конкурентности в Go (VK Go Lectures).
Код находится в spammer.go
