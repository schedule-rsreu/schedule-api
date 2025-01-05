# Schedule API

[![lint workflow](https://github.com/schedule-rsreu/schedule-api/actions/workflows/lint.yml/badge.svg)](https://github.com/schedule-rsreu/schedule-api/actions/workflows/lint.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/schedule-rsreu/schedule-api)
[![GitHub License](https://img.shields.io/badge/license-MIT-blue)](/LICENSE)
[![codecov](https://codecov.io/github/schedule-rsreu/schedule-api/graph/badge.svg?token=IFHLWELSNW)](https://codecov.io/github/schedule-rsreu/schedule-api)
[![CodeQL](https://github.com/schedule-rsreu/schedule-api/actions/workflows/codeql.yml/badge.svg)](https://github.com/schedule-rsreu/schedule-api/actions/workflows/codeql.yml "Code quality workflow status")


API для [бота](https://t.me/schedule_rsreu_bot) расписания
занятий [РГРТУ](https://rsreu.ru/studentu/raspisanie-zanyatij).

## Запуск

Для запуска понадобиться `make`
и `docker` ([инструкции по установке `docker`](https://docs.docker.com/engine/install/)).

Запуск локально, с поднятой базой данных отдельно. (Базу данных можно поднять в `docker`
выполнив `docker compose up mongodb -d`):

```shell
make run
```

Запуск всего проекта с помощью `docker compose`:

```shell
make d
```

## Локальная разработка

Для работы некоторых линтеров нужен diff. Для Windows его можно скачать
по [ссылке](https://deac-riga.dl.sourceforge.net/project/gnuwin32/diffutils/2.8.7-1/diffutils-2.8.7-1.exe?viasf=1).

### Линтеры

- Установка:

```shell
make install
```

- Запуск проверок:

```shell
make lint
```

- Исправление замечаний линтреа автоматически, если возможно:

```shell
make fix
```
