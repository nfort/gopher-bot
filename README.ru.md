<div align="center">

# gopher-bot

Бот для проверки Pull Request's в Gitea

<h4>
  <a href="#-установка">Установка</a>
  ·
  <a href="#-разработка">Разработка</a>
</h4>

![alt text](https://github.com/nfort/gopher-bot/blob/main/screenshot.png?raw=true)

</div>

## ✨ Возможности

- Компиляция кода и проверка на ошибки сборки
- Запуск линтера
- Запуск автоматизированных тестов для проверки работоспособности кода
- Анализ покрытия кода тестами 
- Может выполнять команды из Makefile (make build, lint, test)

## 📦 Установка

Для начала настройте Gitea

Убедитесь что Gitea позволят взаймодействовать с ботом, 
для этого в конфиге должна быть прописана директива `ALLOWED_HOST_LIST` с хостом, на котором развернут gopher-bot.

```bash
cat << EOF >> /etc/gitea/app.ini
[webhook]
ALLOWED_HOST_LIST = *
EOF
```
1. Для всех репозиторий, где вы хотите использовать gopher-bot, откройте настройки webhook репозитория и создайте новый `Gitea` webhook (Trigger On and Branch filter depend on what you would like to use, of course)
    * Target URL: URL на котором развернут gopher-bot, вместе `/hook` сегментом (`http://gopher-bot:8080/hook`)
    * HTTP Method: `POST`
    * POST Content Type: `application/json`
    * Secret: the secret your config contains
    * Trigger On: Custom Events...
      * Pull Request Events
        * Pull Request
        * Pull Request Synchronized
    * Branch filter: `*`
    * Active: ✅
2. Добавить пользователя gopher-bot c токеном c правами на repo
2. Добавить пользователя gopher-bot в репозиторий
3. Установите gopher-bot

Можно выполнить установка двумя способа: Docker или бинарник

### Docker

На машине выполнить команды, указав gitea_host, token пользователя gopher-bot и secret

```bash
docker volume create gopher-bot_config
docker volume create gopher-bot_var
docker run -it --rm -v gopher-bot_config:/etc/gopher-bot busybox sh -c 'cat << EOF >> /etc/gopher-bot/config.ini
[tokens]
"[gitea_host]"=gopher-bot:[token]

[server]
DEBUG_MODE=true
SECRET=[SECRET]'
docker run --restart always -p 8080:8080 -v gopher-bot_config:/etc/gopher-bot -v gopher-bot_var:/var/gopher-bot --name gopher-bot nfort/gopher-bot:1.0.0
```

### Бинарник

Соберите или загрузите бинарник из релиза

```bash
CGO_ENABLED=0 GOOS=linux go build -o gopher-bot cmd/main.go
```

Добавьте конфиг файл

```bash
cat << EOF >> /etc/gopher-bot/config.ini
[tokens]
"http://[gitea_host]:[gitea_port]"=gopher-bot:[token]

[server]
DEBUG_MODE=true
SECRET=[secret]
```

Добавьте бинарник на сервер и запустите.
Если вы используете golangci-lint или другие инструменты в качестве зависимостей проекта, их также следует установить на сервер.

## 🚀 Разработка

После запуска `docker compose up`, нужно остановить.

1. Выполнить команду

```bash
docker run -it --rm -v gitea_gitea-config:/etc/gitea busybox sh -c 'cat << EOF >> /etc/gitea/app.ini
[webhook]
ALLOWED_HOST_LIST = *
EOF'
```

2. При настройки Gitea указать `gitea:3000` вместо `localhost:3000` в качестве хоста.
Добавить пользотеля `gopher-bot`, добавить токен с правами на repo.
Скопировать токен

```bash
docker volume create gopher-bot_config
docker volume create gopher-bot_var
docker run -it --rm -v gopher-bot_config:/etc/gopher-bot busybox sh -c 'cat << EOF >> /etc/gopher-bot/config.ini
[tokens]
"http://gitea:3000"=gopher-bot:[token]

[server]
DEBUG_MODE=true
SECRET=iNeydroTioUC'
docker run --restart always -p 8080:8080 -v gopher-bot_config:/etc/gopher-bot -v gopher-bot_var:/var/gopher-bot --name gopher-bot nfort/gopher-bot:1.0.0
```

3. Добавить webhook для репозитория
4. Указать SECRET
5. В качестве хоста указать `gopher-bot:8080/hooks`
6. Дать права на PR, PR Synchronize
7. Добавить пользователя gopher-bot в репозиторий

## Как добавить gopher-bot в systemd

```bash
cat << EOF >> /etc/systemd/system/gopher-bot.service
[Unit]
Description=gopher-bot

[Service]
Environment="HOME=/root"
Environment="GOPATH=/root/.go"
Environment="GOCACHE=/root/.go-cache"
ExecStart=/opt/gopher-bot/gopher-bot
Restart=always

StandardOutput=append:/var/log/gopher-bot.log
StandardError=append:/var/log/gopher-bot.log

[Install]
WantedBy=multi-user.target
EOF
```

Выполнить перезагрузку systemd 

```bash
systemctl daemon-reload
```

Добавить сервис в systemd

```bash
systemctl enable gopher-bot
```
