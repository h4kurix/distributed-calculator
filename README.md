# Calculator Web Service

## Описание проекта
Веб-сервис для вычисления арифметических выражений. Сервис принимает математические выражения через HTTP POST-запрос и возвращает их результат.

## Поддерживаемые операции
- Сложение (+)
- Вычитание (-)
- Умножение (*)
- Деление (/)
- Скобки для изменения порядка операций

## Ограничения
- Поддерживаются только целые и дробные числа
- Недопустимы символы, не относящиеся к цифрам и базовым арифметическим операциям


## Установка и запуск

### Клонирование репозитория
```bash
git clone https://github.com/h4kurix/calc-service.git
cd calc-service
```

### Запуск сервиса
```bash
go run ./cmd/calc_service/...
```

Сервис будет доступен по адресу `http://localhost:8080`

### Запуск тестов
```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с подробным выводом
go test ./... -v

# Запуск тестов с покрытием кода
go test ./... -cover
```

## Примеры curl-запросов

### Успешный расчет (200)
```shell
curl "http://localhost:8080/api/v1/calculate" -Method POST -ContentType "application/json" -Body '{"expression": "2+2*2"}'
```

### Невалидное выражение(422)
```shell
curl "http://localhost:8080/api/v1/calculate" -Method POST -ContentType "application/json" -Body '{"expression": "2+2*d"}'
```

## Внутренная ошибка сервера(500)
```shell
curl "http://localhost:8080/api/v1/calculate" -Method POST -ContentType "application/json" -Body '{"expression": "2+2*("}'
```

