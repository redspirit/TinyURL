# TinyURL - Сервис сокращения ссылок

TinyURL - это легкий и эффективный сервис для создания коротких URL-адресов с поддержкой пользовательских алиасов, ограничением срока действия, статистикой использования и настраиваемой длиной кода.

![Go Version](https://img.shields.io/badge/go-1.24-blue)

## Быстрый старт с Docker

```bash
# Запуск TinyURL
docker run -d -p 8080:8080 -v tinyurl-data:/data --name tinyurl iwnmname/tinyurl
```

## Два способа создания коротких ссылок

### Способ 1: Использование API через curl

```bash
# Создание короткой ссылки
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'

# С пользовательским алиасом
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com", "alias":"mylink"}'

# С ограничением срока действия (7 дней)
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com", "ttl_days": 7}'

# С указанием длины кода (4 символа)
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com", "code_len": 4}'

# Получение статистики
curl http://localhost:8080/stats/mylink
```

### Способ 2: Использование CLI через Docker

```bash
# Создание короткой ссылки (для Linux)
docker run --rm -it --network host iwnmname/tinyurl ./tinyurl-cli -s http://localhost:8080 short https://example.com

# С пользовательским алиасом
docker run --rm -it --network host iwnmname/tinyurl ./tinyurl-cli -s http://localhost:8080 short https://example.com -a mylink

# С ограничением срока действия (7 дней)
docker run --rm -it --network host iwnmname/tinyurl ./tinyurl-cli -s http://localhost:8080 short https://example.com -t 7

# Получение статистики
docker run --rm -it --network host iwnmname/tinyurl ./tinyurl-cli -s http://localhost:8080 stats mylink
```

## Особенности

- ✂️ **Сокращение URL** - превращение длинных ссылок в короткие и удобные
- 🔠 **Пользовательские алиасы** - создание запоминающихся коротких кодов
- ⏱️ **Ограничение срока действия** - установка времени жизни ссылок (TTL)
- 📏 **Настраиваемая длина кода** - выбор длины короткого кода от 2 до 10 символов
- 📊 **Статистика переходов** - отслеживание количества кликов по ссылкам
- 🗄️ **SQLite хранилище** - простое хранение без внешних зависимостей

## API

### Создание короткой ссылки
```
POST /shorten
```

Запрос:
```json
{
  "url": "https://example.com",
  "alias": "example",
  "ttl_days": 7,
  "code_len": 6
}
```

Параметры:
- `url` (обязательный) — исходная ссылка
- `alias` (опционально) — пользовательский алиас
- `ttl_days` (опционально) — срок действия в днях
- `code_len` (опционально) — длина генерируемого кода, от 2 до 10 (по умолчанию 6)

Ответ:
```json
{
  "code": "example",
  "short_url": "http://localhost:8080/r/example"
}
```

### Переход по короткой ссылке
```
GET /r/{code}
```
Перенаправляет на оригинальный URL.

### Получение статистики
```
GET /stats/{code}
```

Ответ:
```json
{
  "url": "https://example.com",
  "created_at": "2025-08-19T19:05:32Z",
  "expires_at": "2025-08-26T19:05:32Z",
  "hit_count": 5
}
```

## Управление Docker-контейнером

```bash
# Просмотр логов
docker logs tinyurl

# Остановка контейнера
docker stop tinyurl

# Перезапуск контейнера
docker restart tinyurl

# Удаление контейнера
docker rm tinyurl

# Удаление тома с данными
docker volume rm tinyurl-data
```
