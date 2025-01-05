![lint workflow](https://github.com/schedule-rsreu/schedule-api/actions/workflows/lint.yml/badge.svg)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/schedule-rsreu/schedule-api)
[![GitHub License](https://img.shields.io/badge/license-MIT-blue)](/LICENSE)
[![codecov](https://codecov.io/github/schedule-rsreu/schedule-api/graph/badge.svg?token=IFHLWELSNW)](https://codecov.io/github/schedule-rsreu/schedule-api)

# Schedule API

API для [бота](https://t.me/schedule_rsreu_bot) расписания занятий [РГРТУ](https://rsreu.ru/studentu/raspisanie-zanyatij)

## Запуск

Запуск локально, с поднятой базой данных отдельно. (Базу данных можно поднять в `docker`
выполнив `docker compose up mongodb -d` )

```shell
make run
```

Запуск всего проекта с помощью `docker compose`

```shell
make d
```

## Локальная разработка

### Линтеры

- Установка

```shell
make install
```

- Запуск проверок

```shell
make lint
```

- Исправление замечаний линтреа автоматически, если возможно

```shell
make fix
```
