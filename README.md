# 📦 DeliveryService — микросервисная система доставки

**DeliveryService** — это система, построенная по микросервисной архитектуре, реализующая функционал расчета стоимости доставки, управления посылками, авторизации, оплаты, мониторинга и сбора метрик. Система построена с использованием **gRPC, REST, Kafka, Prometheus, Grafana, PostgreSQL и MongoDB**.

---

## ⚙️ Стек технологий

| Категория           | Технологии                                |
|---------------------|--------------------------------------------|
| Язык                | Go (Golang)                                |
| Коммуникация        | gRPC, REST                                 |
| Брокер сообщений    | Apache Kafka                               |
| Базы данных         | PostgreSQL, MongoDB                        |
| Мониторинг          | Prometheus + Grafana                       |
| Контейнеризация     | Docker, Docker Compose                     |
| API Gateway         | Кастомный на Go с middleware               |
| Метрики             | Prometheus Exporters                       |

---
## 📁 Структура проекта
```bash
DeliveryService/
├── auth/               # Авторизация
├── calculator/         # Сервис расчета стоимости доставки
├── client/             # CLI-клиент
├── database/           # Работа с базой данных
├── gateway/            # API Gateway
├── payment/            # Платежный сервис
├── interface/          # Сайт для рабоыт с сервисом
├── proto/              # Протобуф-схемы
├── grafana/            # Dashboards + Provisioning
├── scripts/            # Тестовые скрипты
├── mongo-init/         # Данные для заполнения mongodb
├── docker-compose.yml
├── prometheus.yml
└── Makefile
```
---

## Быстрый старт

### 🔧 Предварительные требования
- Go 1.20+
- Docker и Docker Compose
- `protoc` (Protocol Buffers compiler)

### 🐳 Запуск всей системы
```bash
make up  # Поднимает все сервисы, БД и мониторинг
```

### Запускаемые компоненты:

* Zookeeper + Kafka

* MongoDB + PostgreSQL

* Prometheus + Grafana

* Kafka Exporter

* MongoDB Exporter

* Kafdrop (UI для Kafka)

* Все микросервисы

## 🛠️ Сборка и запуск вручную
```bash
make gateway    # API Gateway
make auth       # Сервис авторизации
make calculate  # Сервис расчета доставки
make payment    # Платежный сервис
make db         # Инициализация базы данных
make insert     # Вставка тестовых данных
```

## 🧪 Тестирование
```bash
make test
```

## 🔌 gRPC Клиенты (пример)
```go
authClient := grpcclient.NewAuthGRPCClient("localhost:50052")
calculatorClient := grpcclient.NewCalculatorClient("localhost:50051")
paymentClient := grpcclient.NewPaymentGRPCClient("localhost:50053")
```
---
## 📡 Метрики и мониторинг

| Компонент  | URL                      |
|------------|--------------------------|
| Prometheus | http://localhost:9090    |
| Grafana    | http://localhost:3033    |
| Kafdrop    | http://localhost:9003    |

**Grafana:** Логин/пароль: `admin/admin`

### Пример конфигурации Prometheus:

```yaml
- job_name: "api-gateway"
  metrics_path: "/metrics"
  static_configs:
    - targets: ["host.docker.internal:8228"]
```
---
## 🌐 API Маршруты

| Метод   | Путь                            | Защищен | Описание                          | Query параметры                            |
|---------|---------------------------------|---------|-----------------------------------|--------------------------------------------|
| POST    | `/api/register`                 | ❌      | Регистрация пользователя          | —                                           |
| POST    | `/api/register-moderator`       | ❌      | Регистрация модератора            | —                                           |
| POST    | `/api/login`                    | ❌      | Авторизация                       | —                                           |
| POST     | `/api/calculate`               | ✅      | Расчет стоимости доставки         | — (в теле JSON)                             |
| POST    | `/api/calculate-by-tariff`      | ✅      | Расчет по тарифу                  | — (в теле JSON)                             |
| GET     | `/api/tariffs`                  | ✅      | Получение всех тарифов            | —                                           |
| POST     | `/api/tariff`                  | ✅      | Создание тарифа                   | — (в теле JSON)                             |
| DELETE     | `/api/tariff`                | ✅      | Удаление тарифа                   | — (в теле JSON)                             |
| POST    | `/api/payment/confirm`          | ✅      | Подтверждение оплаты              | — (в теле JSON)                             |
| GET     | `/api/profile`                  | ✅      | Просмотр профиля пользователя     | —                                           |
| GET     | `/api/packages`                 | ✅      | Получение всех посылок            | `status`, `limit`, `offset`                 |
| GET     | `/api/packages/my`              | ✅      | Получение своих посылок           | `status`, `limit`, `offset`                 |
| POST    | `/api/packages`                 | ✅      | Создание посылки                  | — (в теле JSON)                             |
| POST    | `/api/packages/create`          | ✅      | Создание поссылки (Kafka producer)| — (в теле JSON)                             |
| PUT     | `/api/packages`                 | ✅      | Обновление посылки                | — (в теле JSON)                             |
| DELETE  | `/api/packages`                 | ✅      | Удаление посылки                  | `id`                                        |
| GET     | `/api/packages/status`          | ✅      | Получение статуса посылки         | `id`                                        |
| POST    | `/api/packages/cancel`          | ✅      | Отмена посылки                    | `id`                                        |
---
## 📜 Генерация Protobuf

```bash
make protocalc  # Протокол калькулятора
make protoauth  # Протокол авторизации
make protopay   # Протокол платежей
```
## 📬 Kafka

**Producer**  
Публикует события создания заказов в Kafka topics

**Consumer**  
Обрабатывает события из Kafka и:
- Обновляет статусы заказов в MongoDB
- Сохраняет историю платежей в PostgreSQL

## 🛡️ Middleware

| Middleware | Описание |
|------------|----------|
| `EnableCORS` | Поддержка кросс-доменных запросов |
| `AuthMiddleware` | JWT-авторизация запросов |
| `LogMiddleware` | Логирование всех HTTP-запросов |

## 📈 Grafana Dashboard

**Основные метрики:**
- 🚦 **HTTP-метрики**
  - Количество запросов по сервисам
  - Статус коды ответов
  - Время обработки запросов

- ⏱ **Производительность**  
  - Среднее время отклика API
  - 95-й перцентиль времени ответа

- 📦 **Бизнес-метрики**
  - Количество созданных заказов
  - Статистика по оплатам
  - Активные пользователи

- 🛠 **Системные метрики**
  - Состояние Kafka (лаг, потребление)
  - MongoDB производительность
  - Использование ресурсов сервисов

## 🏁 Заключение

**DeliveryService** представляет собой современное микросервисное решение для управления доставкой, объединяющее:

- ⚡️ Высокопроизводительные gRPC-сервисы
- 🚀 Асинхронную обработку через Kafka
- 🔒 Безопасную JWT-аутентификацию
- 📊 Полноценный мониторинг через Grafana/Prometheus
- 🐳 Простую развертку через Docker