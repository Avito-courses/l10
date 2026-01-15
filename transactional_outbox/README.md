# Transactional Outbox Pattern

Демонстрация паттерна **Transactional Outbox** для гарантированной доставки сообщений в распределенных системах.

## 🎯 Что такое Transactional Outbox?

**Transactional Outbox** - это паттерн, который гарантирует атомарность операций с базой данных и отправки сообщений в брокер (Kafka, RabbitMQ и т.д.).

### Проблема

Классическая проблема распределенных систем:

```go
// НЕ АТОМАРНО!
// Если после сохранения заказа произойдет сбой,
// сообщение не отправится в Kafka
order := CreateOrder(...)           // Запись в БД
kafka.Publish("OrderCreated", order) // Отправка в Kafka
```

Возможные проблемы:
1. Заказ создан, но сообщение не отправлено (сбой сети/Kafka)
2. Сообщение отправлено, но заказ не создан (сбой БД)
3. Дублирование сообщений при retry

### Решение: Transactional Outbox

```go
// АТОМАРНО!
tx := db.Begin()

// 1. Сохраняем заказ
order := CreateOrder(tx, ...)

// 2. Сохраняем сообщение в outbox В ТОЙ ЖЕ ТРАНЗАКЦИИ
outbox := OutboxMessage{
    EventType: "OrderCreated",
    Payload: order,
}
SaveToOutbox(tx, outbox)

tx.Commit() // Либо оба успешно, либо оба откатятся

// 3. Отдельный процесс читает из outbox и публикует в Kafka
```

## 🏗️ Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                    Application                              │
│  OrderService.CreateOrder()                                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
          ┌──────────────────────────┐
          │   PostgreSQL Database    │
          │                          │
          │  ┌──────────────────┐    │
          │  │  orders table    │◄───┼──── 1. Запись заказа
          │  └──────────────────┘    │
          │                          │
          │  ┌──────────────────┐    │
          │  │  outbox table    │◄───┼──── 2. Запись события
          │  └──────────────────┘    │      (В ТОЙ ЖЕ ТРАНЗАКЦИИ!)
          └──────────┬───────────────┘
                     │
                     │ 3. Outbox Processor
                     │    (polling every 2s)
                     ▼
          ┌──────────────────────┐
          │   Apache Kafka       │
          │  Topic: order-events │◄────── 4. Публикация события
          └──────────────────────┘
                     │
                     ▼
          ┌──────────────────────┐
          │   Consumers          │
          │  (другие сервисы)    │
          └──────────────────────┘
```

## 📁 Структура проекта

```
transactional_outbox/
├── docker-compose.yml          # PostgreSQL, Kafka, Zookeeper
├── migrations/
│   └── 01_init.sql            # SQL миграции
├── models/
│   ├── order.go               # Модель заказа
│   ├── outbox.go              # Модель outbox
│   └── errors.go              # Ошибки
├── repository/
│   ├── order_repository.go    # Репозиторий заказов
│   └── outbox_repository.go   # Репозиторий outbox
├── service/
│   └── order_service.go       # Бизнес-логика (использует паттерн)
├── publisher/
│   ├── kafka_publisher.go     # Публикация в Kafka
│   └── outbox_processor.go    # Обработка outbox
├── main.go                    # Точка входа
├── go.mod
└── README.md
```

## 🚀 Быстрый старт

### 1. Запустить инфраструктуру (PostgreSQL, Kafka)

```bash
cd transactional_outbox

# Запустить Docker Compose
docker-compose up -d

# Проверить, что все запустилось
docker-compose ps
```

### 2. Установить зависимости

```bash
go mod download
```

### 3. Запустить приложение

```bash
go run main.go
```

## 📊 Что происходит при запуске

1. **Подключение к БД и Kafka**
   - Приложение подключается к PostgreSQL
   - Создается Kafka producer

2. **Запуск Outbox Processor**
   - Отдельная горутина каждые 2 секунды проверяет outbox
   - Находит необработанные сообщения
   - Публикует их в Kafka
   - Помечает как обработанные

3. **Создание заказов**
   - Создается 3 заказа
   - Каждый заказ = 1 запись в `orders` + 1 запись в `outbox`
   - **Обе записи в одной транзакции!**

4. **Публикация событий**
   - Outbox processor читает из `outbox`
   - Публикует в Kafka topic `order-events`
   - Помечает сообщения как обработанные

## 💡 Ключевые моменты паттерна

### 1. Атомарность
```go
tx := db.Begin()

// Обе операции в одной транзакции
order := CreateOrder(tx, ...)
SaveToOutbox(tx, outboxMsg)

tx.Commit() // Либо обе успешны, либо обе откатятся
```

### 2. Гарантия доставки
- Если заказ создан → сообщение гарантированно в outbox
- Outbox processor будет retry до успешной публикации
- **At-least-once delivery**

### 3. Идемпотентность
- Возможны дубликаты (при retry)
- Consumers должны быть идемпотентными
- Используйте уникальные ID событий

### 4. Eventual Consistency
- Небольшая задержка (2 секунды в нашем примере)
- Настраивается через `pollInterval`

## 🔍 Мониторинг

### Kafka UI
Открой http://localhost:8080 для просмотра сообщений в Kafka

### PostgreSQL
```bash
# Подключиться к БД
docker exec -it outbox_postgres psql -U outbox_user -d outbox_db

# Посмотреть заказы
SELECT * FROM orders;

# Посмотреть outbox
SELECT * FROM outbox;

# Необработанные сообщения
SELECT * FROM outbox WHERE processed = false;

# Обработанные сообщения
SELECT * FROM outbox WHERE processed = true;
```

## 📈 Производительность

### Настройки Outbox Processor

```go
const (
    pollInterval = 2 * time.Second  // Как часто проверять outbox
    batchSize    = 10               // Сколько сообщений обрабатывать за раз
)
```

**Рекомендации:**
- **pollInterval**: 1-5 секунд для большинства случаев
- **batchSize**: 10-100 в зависимости от нагрузки
- Для высокой нагрузки: запускайте несколько processor'ов

## 🛠️ Расширения

### 1. Партиционирование outbox
Для высокой нагрузки разделите outbox по агрегатам:

```sql
CREATE TABLE outbox_orders (...);
CREATE TABLE outbox_payments (...);
```

### 2. Несколько processor'ов
Запустите несколько экземпляров для параллельной обработки:

```go
for i := 0; i < numWorkers; i++ {
    go outboxProcessor.Start(ctx)
}
```

### 3. Dead Letter Queue
Добавьте таблицу для сообщений, которые не удалось обработать:

```sql
CREATE TABLE outbox_dlq (
    id SERIAL PRIMARY KEY,
    original_message_id INT,
    error_message TEXT,
    retry_count INT
);
```

## 🎓 Когда использовать?

### Используйте Transactional Outbox когда:
- Нужна гарантия доставки сообщений
- Критична консистентность данных
- Используете event-driven архитектуру
- Есть микросервисы, которые должны узнавать о событиях

### Не используйте когда:
- Нужна синхронная обработка (используйте API)
- Низкая нагрузка и простая архитектура
- Не критична eventual consistency

## 🐛 Troubleshooting

### Сообщения не публикуются в Kafka
```bash
# Проверьте, что Kafka запущена
docker-compose ps kafka

# Проверьте логи Kafka
docker-compose logs kafka

# Проверьте outbox
docker exec -it outbox_postgres psql -U outbox_user -d outbox_db -c "SELECT * FROM outbox WHERE processed = false;"
```

### База данных не доступна
```bash
# Проверьте статус PostgreSQL
docker-compose ps postgres

# Перезапустите контейнер
docker-compose restart postgres
```

### Очистка данных
```bash
# Остановить все
docker-compose down

# Удалить volumes (очистит БД)
docker-compose down -v

# Запустить заново
docker-compose up -d
```

## 📚 Дополнительные материалы

- [Microservices Patterns by Chris Richardson](https://microservices.io/patterns/data/transactional-outbox.html)
- [Event-Driven Microservices](https://www.oreilly.com/library/view/building-event-driven-microservices/9781492057888/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)

## 🎯 Альтернативные подходы

1. **Change Data Capture (CDC)**
   - Debezium читает transaction log
   - Не требует outbox таблицу
   - Более сложная настройка

2. **Saga Pattern**
   - Для распределенных транзакций
   - Компенсирующие транзакции
   - Более сложная логика

3. **Two-Phase Commit (2PC)**
   - Распределенные транзакции
   - Плохая производительность
   - Не рекомендуется

## 🏁 Заключение

Transactional Outbox - это надежный способ гарантировать доставку сообщений в event-driven системах. 

**Ключевые преимущества:**
- Атомарность операций
- Гарантия доставки
- Простота реализации
- Надежность

Используйте этот паттерн в production системах для критичных событий!


