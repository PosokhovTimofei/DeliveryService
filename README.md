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
├── consumer/           # Kafka consumer
├── database/           # Работа с базой данных
├── gateway/            # API Gateway
├── payment/            # Платежный сервис
├── producer/           # Kafka producer
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
make producer   # Kafka producer
make consumer   # Kafka consumer
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

| Метод | Путь                   | Защищен | Описание                          |
|-------|------------------------|---------|-----------------------------------|
| POST  | `/api/register`        | ❌      | Регистрация пользователя         |
| POST  | `/api/login`           | ❌      | Авторизация                      |
| GET   | `/api/calculate`       | ✅      | Расчет стоимости доставки        |
| POST | `/api/calculate-by-tariff` |  ✅      | Расчет по тарифу |
| GET | `/api/tariffs` |  ✅      | Получение всех тарифов |
| POST  | `/api/create`          | ✅      | Создание заказа (Kafka producer) |
| GET   | `/api/packages`        | ✅      | Получение всех посылок           |
| GET   | `/api/my/packages`     | ✅      | Получение своих посылок          |
| GET   | `/api/profile`         | ✅      | Просмотр профиля пользователя    |
| POST  | `/api/payment/confirm` | ✅      | Подтверждение оплаты             |
| GET   | `/api/packages/{packageID}/status` |  ✅      | Просмотр статуса доставки             |
| POST   | `/api/packages/{packageID}/cancel` |  ✅      | Отмена посылки |
| DELETE | `/api/packages/{packageID}` |  ✅      | Удаление посылки |
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
| `enableCORS` | Поддержка кросс-доменных запросов |
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