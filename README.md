# Workshop 10: Stability Patterns & Transactional Outbox

Коллекция примеров паттернов стабильности и паттерна Transactional Outbox на Go.

## Структура проекта

```
workshop10/
├── stability/              # Паттерны стабильности
│   ├── rate_limiter/      # Rate Limiting (4 алгоритма)
│   ├── circuit_breaker/   # Circuit Breaker
│   ├── retry/             # Retry с различными стратегиями
│   ├── timeout/           # Timeout
│   └── fallback/          # Fallback / Graceful Degradation
└── transactional_outbox/  # Transactional Outbox Pattern
```

## Паттерны стабильности (Stability Patterns)

### 1. Rate Limiter

Ограничивает количество запросов к сервису для защиты от перегрузок.

**Реализованные алгоритмы:**

- **Token Bucket** (`rate_limiter/token_bucket/`)
  - Ведро с токенами, пополняется с фиксированной скоростью
  - Позволяет всплески трафика в пределах размера ведра
  - Порт: 8081

- **Leaky Bucket** (`rate_limiter/leaky_bucket/`)
  - Запросы обрабатываются с постоянной скоростью
  - Очередь для входящих запросов
  - Порт: 8082

- **Fixed Window** (`rate_limiter/fixed_window/`)
  - Фиксированные временные окна
  - Счетчик сбрасывается в начале каждого окна
  - Порт: 8083

- **Sliding Window** (`rate_limiter/sliding_window/`)
  - Скользящее временное окно
  - Более точное ограничение, чем Fixed Window
  - Порт: 8084

**Запуск примера:**
```bash
cd stability/rate_limiter/token_bucket
go run .
curl http://localhost:8081/api
```

**Подробное описание алгоритмов:**
См. `stability/rate_limiter/ALGORITHMS.md`

### 2. Circuit Breaker

Предотвращает каскадные сбои, временно блокируя запросы к падающему сервису.

**Состояния:**
- **Closed** - нормальная работа, запросы проходят
- **Open** - сервис недоступен, запросы отклоняются
- Автоматический переход Open → Closed после timeout

**Запуск:**
```bash
cd stability/circuit_breaker
go run .
```

**Применение:**
- Защита от каскадных сбоев
- Быстрый fail для недоступных сервисов
- Автоматическое восстановление

### 3. Retry

Повторяет неудачные операции с различными стратегиями backoff.

**Стратегии:**
- **Exponential Backoff** - экспоненциальное увеличение задержки
- **Fixed Delay** - фиксированная задержка между попытками
- **Jitter** - добавление случайности к задержке
- **Full Jitter** - полностью случайная задержка в диапазоне

**Запуск:**
```bash
cd stability/retry
go run .
```

**Применение:**
- Временные сетевые сбои
- Rate limiting на стороне сервера
- Нестабильные внешние API

### 4. Timeout

Ограничивает время выполнения операций для предотвращения зависаний.

**Реализация:**
- Использует `context.WithTimeout`
- Отмена операции при превышении времени
- Освобождение ресурсов

**Запуск:**
```bash
cd stability/timeout
go run .
```

**Применение:**
- Защита от медленных операций
- Контроль времени выполнения запросов
- Предотвращение resource exhaustion

### 5. Fallback / Graceful Degradation

Обеспечивает деградацию функциональности при сбоях некритичных сервисов.

**Механизм:**
- Основной функционал (цены) - всегда работает
- Дополнительный функционал (рекомендации, отзывы) - с fallback
- Флаг `isGD` указывает на использование кэша/fallback данных

**Запуск:**
```bash
cd stability/fallback
go run .
```

**Применение:**
- Критичные операции должны работать всегда
- Некритичные функции могут деградировать
- Улучшение user experience при частичных сбоях

## Transactional Outbox Pattern

Гарантирует атомарность записи в базу данных и отправки событий в message broker.

**Компоненты:**
- PostgreSQL - хранение данных и outbox таблицы
- Kafka - message broker для событий
- Outbox Processor - фоновый процесс для публикации событий

**Гарантии:**
- Атомарность: запись в БД и outbox в одной транзакции
- At-least-once delivery
- Eventual consistency

**Запуск:**
```bash
cd transactional_outbox
docker-compose up -d
go run .
```

**Проверка:**
- Kafka UI: http://localhost:8080
- PostgreSQL: порт 5432

**Подробнее:**
См. `transactional_outbox/README.md`

## Быстрый старт

### Проверка всех примеров

**Rate Limiters:**
```bash
cd stability/rate_limiter/token_bucket && go run . &
cd stability/rate_limiter/leaky_bucket && go run . &
cd stability/rate_limiter/fixed_window && go run . &
cd stability/rate_limiter/sliding_window && go run . &
```

**Остальные паттерны:**
```bash
cd stability/circuit_breaker && go run .
cd stability/retry && go run .
cd stability/timeout && go run .
cd stability/fallback && go run .
```

**Transactional Outbox:**
```bash
cd transactional_outbox
docker-compose up -d
sleep 10
go run .
```

## Когда использовать какой паттерн

| Паттерн | Проблема | Решение |
|---------|----------|---------|
| **Rate Limiter** | Перегрузка сервиса | Ограничение входящих запросов |
| **Circuit Breaker** | Каскадные сбои | Временная блокировка запросов к падающему сервису |
| **Retry** | Временные сбои | Повторная попытка с backoff |
| **Timeout** | Зависшие операции | Ограничение времени выполнения |
| **Fallback** | Недоступность некритичных сервисов | Использование кэша или упрощенных данных |
| **Transactional Outbox** | Несогласованность БД и событий | Атомарная запись в БД и message broker |

## Комбинирование паттернов

Паттерны часто используются вместе:

```
HTTP Request → Rate Limiter → Circuit Breaker → Retry → Timeout → Service
                                                              ↓ (if fails)
                                                           Fallback
```

**Пример:**
1. Rate Limiter отсекает лишние запросы
2. Circuit Breaker проверяет доступность сервиса
3. Retry повторяет при временных сбоях
4. Timeout предотвращает зависания
5. Fallback предоставляет альтернативные данные

## Требования

- Go 1.21+
- Docker & Docker Compose (для transactional_outbox)
- curl (для тестирования HTTP endpoints)

## Дополнительная информация

- Rate Limiter алгоритмы: `stability/rate_limiter/ALGORITHMS.md`
- Transactional Outbox: `transactional_outbox/README.md`
