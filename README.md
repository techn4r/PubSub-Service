# 📡 PubSubService — gRPC Pub/Sub сервер на Go

Пример реализации Pub/Sub‑сервиса с использованием gRPC на языке Go.
Проект демонстрирует работу подписок, публикаций сообщений и стриминга через gRPC.

---

## 📂 Структура проекта

```
pubsubservice/
├── grpc/
│   └── pubsub/               # Сгенерированные gRPC‑файлы (.pb.go)
├── go.mod                    # Go‑module
├── main.go                   # Точка входа (запуск сервера)
├── pubsub.proto              # Описание gRPC‑API (protobuf)
├── server.go                 # Реализация gRPC‑сервера
├── subpub.go                 # Реализация логики Pub/Sub
├── subpub_test.go            # Юнит‑тесты
```

---

## 🚀 Быстрый старт

### 1. Сгенерировать gRPC‑код

```bash
protoc \
  --proto_path=. \
  --go_out=paths=source_relative:./grpc/pubsub \
  --go-grpc_out=paths=source_relative:./grpc/pubsub \
  pubsub.proto
```

### 2. Установить зависимости

```bash
go mod tidy
```

### 3. Запустить сервер

```bash
go run .
```

Ожидаемый вывод:

```
gRPC server listening on :50051
```

---

## 🛠️ Проверка Publish через grpcurl

```bash
grpcurl -plaintext \
  -d '{"subject":"demo","message":"Hello from grpcurl"}' \
  localhost:50051 pubsub.PubSub/Publish
```

Ожидаемый ответ:

```json
{
  "success": true
}
```

---

## 🧪 Тесты

```bash
go test ./...
```

---

## 📡 Что внутри

* **Subscribe** — серверный стрим, через который клиенты получают сообщения выбранного subject.
* **Publish** — рассылает сообщение всем активным подписчикам subject.
* Внутренняя логика Pub/Sub реализована вручную (без внешних брокеров).

---

## ✅ Стек

* Go 1.21+
* gRPC
* Protocol Buffers

---

## 📝 Лицензия

MIT License
