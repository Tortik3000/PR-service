

- прогнать линтеры
- описать нагрузочное тестирование
- написать доку

# Документация PR service

### Требования

---

Для работы необходимо установить:
- [Docker](https://www.docker.com/) (20.10+)
- [Docker Compose](https://docs.docker.com/compose/) (1.29+)
- [Make](https://www.gnu.org/software/make/) (4.3+) 

### Настройка переменных окружения

---

[.env](../.env.example) файл:

```
# Порты сервиса
REST_PORT=8080
METRICS_PORT=9000

# PostgreSQL
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=pr-service
POSTGRES_USER=ed
POSTGRES_PASSWORD=1234567

# Grafana & Prometheus
GRAFANA_PORT=3000
PROMETHEUS_PORT=9090
DS_PROMETHEUS=ds-prometheus-1
```

### Запуск через Docker Compose

---

```shell script

  docker-compose up -d
  
```

### Доступные сервисы

После успешного запуска доступны следующие интерфейсы:

| Сервис     | URL | Описание               |
|------------|-----|------------------------|
| REST API   | http://localhost:8080 | REST эндпоинты         |
| metrics    | http://localhost:9000/metrics | metrics                |
| Prometheus | http://localhost:9090 | Метрики                |
| Grafana    | http://localhost:3000 | Дашборды (admin/admin) |

### Запуск тестов

--- 

Для интеграционных тестов нужно поднять тестовую бд
```shell

sudo docker compose --env-file integration/pr-service/.env.test -f docker-compose.test.yml up -d

```

### Настройка переменных для подключения к тестовой бд

---

[.env.test](../integration/pr-service/.env.test) файл:

```
# Порты сервиса
REST_PORT=8080
METRICS_PORT=9000

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=2345
POSTGRES_DB=pr-service
POSTGRES_USER=ed
POSTGRES_PASSWORD=1234567
```

### Makefile

---

Для удобств локальной разработки сделан [`Makefile`](Makefile). Имеются следующие команды:

Установить необходимые зависимости
```bash

make bin-deps

```
Запустить полный цикл (линтер, тесты):

```bash 

make all

```

Запустить только тесты:

```bash

make test

``` 

Запустить линтер:

```bash

make lint

```

При разработке на Windows рекомендуется использовать [WSL](https://learn.microsoft.com/en-us/windows/wsl/install)
---

## Структура проекта

```
pr-service/
├── api/                    # OpenAPI спецификация
├── cmd/                    # Точки входа приложения
├── config/                 # Конфигурация
├── db/                     # Миграции и скрипты БД
├── docs/                   # Документация
├── generated/             # Сгенерированный код 
├── infra/                 # Конфигурация инфраструктуры
│   ├── grafana/          # Дашборды и datasources
│   ├── prometheus.yml    # Конфигурация Prometheus
├── integration/            # Интеграционные тесты
├── internal/              # Внутренний код приложения
│   ├── app.go 
│   ├── controller/       # HTTP хендлеры
│   ├── usecase/          # Бизнес-логика
│   ├── repository/       # Работа с БД
│   ├── metrics/          # Prometheus метрики
│   ├── models/          # Доменные модели и ошибки
│   └── middleware/        
│   
├── k6/                    # Нагрузочные тесты
└── ...
```

---

## Нагрузочное тестирование

### k6

Проект включает набор нагрузочных тестов, написанных с использованием [k6](https://k6.io/)

### Установка k6

Перед запуском тестов установите k6:


```shell

    sudo apt-get update
    sudo apt-get install k6
    
```

Для запуска тестов:
```shell

    k6 run k6/reassign_reviewer_load_test.js
    k6 run k6/load_test_without_rw.js
    
```
---

Допущения:
- Использовал поднятие миграций к БД в коде для простоты поднятия сервиса
- Добавил валидацию в OpenAPI спецификацию, так как эту валидацию, 
как правило, нужно делать на уровне хэндлеров, и удобно её получить из спецификации