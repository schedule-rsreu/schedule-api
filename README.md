<p align="center">
    <img height="60px" width="60px" src="https://avatars.githubusercontent.com/u/163825083?s=100&v=4" />
    <h1 align="center">Schedule API</h1>
</p>

<p align="center">
    <a href="https://github.com/schedule-rsreu/schedule-api/actions/workflows/lint.yml"><img src="https://github.com/schedule-rsreu/schedule-api/actions/workflows/lint.yml/badge.svg" /></a>
    <a href="https://goreportcard.com/report/github.com/schedule-rsreu/schedule-api"><img src="https://goreportcard.com/badge/github.com/schedule-rsreu/schedule-api"/></a>
    <a href="https://img.shields.io/github/go-mod/go-version/schedule-rsreu/schedule-api"><img src="https://img.shields.io/github/go-mod/go-version/schedule-rsreu/schedule-api" /></a>
    <a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" /></a>
    <!-- <a href="https://codecov.io/github/schedule-rsreu/schedule-api"><img src="https://codecov.io/github/schedule-rsreu/schedule-api/graph/badge.svg?token=IFHLWELSNW" /></a> -->
    <a href="https://github.com/schedule-rsreu/schedule-api/actions/workflows/codeql.yml" title="Code quality workflow status"><img src="https://github.com/schedule-rsreu/schedule-api/actions/workflows/codeql.yml/badge.svg" /></a>
    <a href="https://github.com/schedule-rsreu/schedule-api/actions/workflows/dependabot/dependabot-updates"><img src="https://badgen.net/github/dependabot/schedule-rsreu/schedule-api" /></a>
</p>


<p align="center">
    API для <a href="https://t.me/schedule_rsreu_bot">бота</a> расписания занятий <a href="https://rsreu.ru/studentu/raspisanie-zanyatij">РГРТУ</a>.
<br>
<a href="https://api.rsreu-schedule.ru/docs/index.html">Swagger documentation</a>
</p>

## Запуск

Для запуска понадобиться `make` ([скачать](https://cmake.org/download/))
и `docker` ([инструкции по установке](https://docs.docker.com/engine/install/)).

Запуск локально, с поднятой базой данных отдельно. (Базу данных можно поднять в `docker`
выполнив `docker compose up postgres -d`):

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

- Исправление замечаний линтеров автоматически, если возможно:

```shell
make format
```

## Деплой в k3s

Workflow `.github/workflows/deploy.yml` публикует приватный image
`ghcr.io/schedule-rsreu/schedule-api:<git-sha>`, создаёт Kubernetes Secrets
через SSH и после первичной миграции обновляет Helm release в namespace
`schedule-api`.

Repository secrets:

- `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`, `DEPLOY_KNOWN_HOSTS`;
- `GHCR_USERNAME` и PAT `GHCR_PULL_TOKEN` с правом `read:packages`;
- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`;
- `POSTGRES_DSN` вида
  `postgres://USER:PASSWORD@schedule-api-postgres:5432/DB?sslmode=disable`.

PostgreSQL 18 хранит данные под `/var/lib/postgresql/18/docker`, поэтому PVC
монтируется в `/var/lib/postgresql`, а не в старый
`/var/lib/postgresql/data`.

Для первой миграции не создавайте repository variable
`K3S_DEPLOY_ENABLED`. После первого успешного workflow упакуйте chart и
установите только базы с baseline:

```bash
helm package charts/schedule-api --destination dist
scp dist/schedule-api-0.1.0.tgz \
  root@109.172.115.63:/tmp/schedule-api-bootstrap.tgz

helm upgrade --install schedule-api /tmp/schedule-api-bootstrap.tgz \
  --kubeconfig /etc/rancher/k3s/k3s.yaml \
  --namespace schedule-api \
  --create-namespace \
  --set-string image.tag='<git-sha>' \
  --set api.replicas=0 \
  --set ingress.enabled=false \
  --set migration.bootstrap=true \
  --wait --wait-for-jobs --timeout 10m
```

Восстановите и сравните PostgreSQL backup до включения API. Только
после проверки установите repository variable `K3S_DEPLOY_ENABLED=true`.
`parser2` должен оставаться остановленным до отдельной миграции в k3s.
