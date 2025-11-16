# Документация PR service

**Все команды запускаются из корня репозитория**

## Требования

Для работы необходимо установить:
- [Docker](https://www.docker.com/) (20.10+)
- [Docker Compose](https://docs.docker.com/compose/) (1.29+)
- [Make](https://www.gnu.org/software/make/) (4.3+) 

## Настройка переменных окружения


нужно создать [.env](../.env.example) файл:

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

## Запуск через Docker Compose


```shell script

make generate
docker-compose up -d
  
```

## Доступные сервисы

После успешного запуска доступны следующие интерфейсы:

| Сервис     | URL | Описание               |
|------------|-----|------------------------|
| REST API   | http://localhost:8080 | REST эндпоинты         |
| metrics    | http://localhost:9000/metrics | metrics                |
| Prometheus | http://localhost:9090 | Метрики                |
| Grafana    | http://localhost:3000 | Дашборды (admin/admin) |

## Настройка переменных для подключения к тестовой бд


[.env.test](../.env.test) файл:

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

## Запуск тестов


Для интеграционных тестов нужно поднять тестовую бд
```shell

docker compose --env-file .env.test -f docker-compose.test.yml up -d

```

Запуск тестов:
```bash

make all

```


## Makefile

---

Для удобств локальной разработки сделан [`Makefile`](Makefile). Имеются следующие команды:

Запустить полный цикл 
(нужно запустить чтобы сгенерировать файлы, сбилдить проект и скачать все зависимости):

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

build:

```bash

make build

```

Генерация файлов:

```bash

make generate

```

При разработке на Windows рекомендуется использовать [WSL](https://learn.microsoft.com/en-us/windows/wsl/install)


## Структура проекта

---

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


## Нагрузочное тестирование

---

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

docker compose up -d

```

```shell

    k6 run k6/reassign_reviewer_load_test.js
    k6 run k6/load_test_without_rw.js
    
```

### Результат запуска [reassign_reviewer_load_test.js](../k6/reassign_reviewer_load_test.js)
    15 комад
    20 человек в каждой команде
    каждый является автором одного PR

    Запрсы отправлются батчем из 10 штук
    после этого обновляется состояние в тесте
    rps: 100

    


  █ TOTAL RESULTS 

    HTTP
    http_req_duration..............: avg=1.23ms min=355.15µs med=1.21ms max=16.25ms p(90)=1.62ms p(95)=1.78ms
      { expected_response:true }...: avg=1.23ms min=564.04µs med=1.21ms max=16.25ms p(90)=1.62ms p(95)=1.78ms
    http_req_failed................: 0.79%  959 out of 120265
    http_reqs......................: 120265 1000.193118/s

    EXECUTION
    dropped_iterations.............: 5      0.041583/s
    iteration_duration.............: avg=2.89ms min=2.18ms   med=2.8ms  max=17.68ms p(90)=3.3ms  p(95)=3.6ms 
    iterations.....................: 11995  99.75734/s
    vus............................: 0      min=0             max=0
    vus_max........................: 1      min=1             max=1


Все http_req_failed это 409(reviewer is not assigned to this PR)
Происходит из-за параллельного изменения несколькими машанами ревьюеров в массиве тестов

### Результат запуска [reassign_reviewer_load_test.js](../k6/load_test_without_rw.js)
    10 комад
    10 человек в каждой команде
    каждый является автором одного PR

    Все операции отправляются параллельно 
    Каждая операция с 10 виртуальных машин
    rps с каждой машины: 100


  █ TOTAL RESULTS 

    HTTP
    http_req_duration..............: avg=1.22ms min=266.86µs med=1.02ms max=20.02ms p(90)=2.21ms p(95)=2.52ms
      { expected_response:true }...: avg=1.22ms min=266.86µs med=1.02ms max=20.02ms p(90)=2.21ms p(95)=2.52ms
    http_req_failed................: 0.00%  0 out of 82106
    http_reqs......................: 82106  575.171824/s

    EXECUTION
    iteration_duration.............: avg=1.17ms min=326.78µs med=1.01ms max=20.62ms p(90)=1.97ms p(95)=2.26ms
    iterations.....................: 72006  504.418951/s
    vus............................: 0      min=0          max=0 
    vus_max........................: 60     min=60         max=60

    NETWORK
    data_received..................: 115 MB 804 kB/s
    data_sent......................: 17 MB  117 kB/s


### Итог
Максимальное время ответа: 20ms 

Значит либо все работает хорошо, либо тесты плохие

Также надо было провести параллельный запуск reassignReviewerPR с конфликтующими операциями(setIsActive and mergePr), 
но тяжело

---

Допущения:
- Использовал поднятие миграций к БД в коде для простоты поднятия сервиса
- Добавил валидацию в OpenAPI спецификацию, так как эту валидацию, 
как правило, нужно делать на уровне хэндлеров, и удобно её получить из спецификации