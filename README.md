<p align="center">
    <img height="60px" width="60px" src="https://avatars.githubusercontent.com/u/163825083?s=100&v=4" />
    <h1 align="center">Schedule API</h1>
</p>

<p align="center">
    <a href="https://github.com/schedule-rsreu/schedule-api/actions/workflows/lint.yml"><img src="https://github.com/schedule-rsreu/schedule-api/actions/workflows/lint.yml/badge.svg" /></a>
    <a href="https://goreportcard.com/report/github.com/schedule-rsreu/schedule-api"><img src="https://goreportcard.com/badge/github.com/schedule-rsreu/schedule-api"/></a>
    <a href="https://img.shields.io/github/go-mod/go-version/schedule-rsreu/schedule-api"><img src="https://img.shields.io/github/go-mod/go-version/schedule-rsreu/schedule-api" /></a>
    <a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" /></a>
    <a href="https://codecov.io/github/schedule-rsreu/schedule-api"><img src="https://codecov.io/github/schedule-rsreu/schedule-api/graph/badge.svg?token=IFHLWELSNW" /></a>
    <a href="https://github.com/schedule-rsreu/schedule-api/actions/workflows/codeql.yml" title="Code quality workflow status"><img src="https://github.com/schedule-rsreu/schedule-api/actions/workflows/codeql.yml/badge.svg" /></a>
    <a href="https://github.com/schedule-rsreu/schedule-api/actions/workflows/dependabot/dependabot-updates"><img src="https://badgen.net/github/dependabot/schedule-rsreu/schedule-api" /></a>
</p>


<p align="center">
    API для <a href="https://t.me/schedule_rsreu_bot">бота</a> расписания занятий <a href="https://rsreu.ru/studentu/raspisanie-zanyatij">РГРТУ</a>.
<br>
<a href="https://api.rsreu-schedule.ru/docs/index.html">Swagger documentation</a>
</p>

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
