# PubSub Service

** PubSub Service** — это демонстрационный проект на Go, реализующий простую Publish–Subscribe шину событий с gRPC API.

## 📋 Как собрать и запустить

1. **Склонируйте репозиторий и перейдите в папку:**

   ```bash
   git clone https://github.com/techn4r/PubSub-Service.git
   cd VKIntern
   ```

2. **Сгенерируйте код из protobuf-схемы:**

   ```bash
   protoc --go_out=. --go-grpc_out=. proto/pubsub.proto
   ```

3. **Установите зависимости и выполните тесты пакета subpub:**

   ```bash
   go mod download
   go test ./subpub
   ```

4. **Соберите бинарь gRPC-сервера:**

   ```bash
   go build -o bin/pubsub-server ./cmd/pubsub-server
   ```

5. **Запустите сервер:**

   ```bash
   ./bin/pubsub-server -port 50051
   ```

   * Параметр `-port` (по умолчанию `50051`) задает порт для gRPC.

6. **Проверьте работу:**

   * **Publish:**

     ```bash
     grpcurl -plaintext -d '{"key":"topic1","data":"hello"}' localhost:50051 pubsub.PubSub/Publish
     ```

   * **Subscribe:**

     ```bash
     grpcurl -plaintext -d '{"key":"topic1"}' localhost:50051 pubsub.PubSub/Subscribe
     ```

## 🔧 Конфигурация

Сервис настраивается через флаги:

| Флаг    | Описание          | Значение по умолчанию |
| ------- | ----------------- | --------------------- |
| `-port` | Порт gRPC-сервера | `50051`               |

## 🛠 Тесты

* Пакет `subpub` покрыт unit-тестами, чтобы проверить доставку, порядок FIFO, отписку и корректное завершение:

  ```bash
  go test ./subpub
  ```

---

*Документация по проекту, сборке и запуску сервиса PubSub.*
