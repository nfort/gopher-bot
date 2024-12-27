<div align="center">
    <a href="README.ru.md">[russian]</a>
</div>

<div align="center">

# gopher-bot

Bot for Pull Request Checks in Gitea

<h4>
  <a href="#-установка">Install</a>
  ·
  <a href="#-разработка">Development</a>
</h4>

![alt text](https://github.com/nfort/gopher-bot/blob/main/screenshot.png?raw=true)

</div>

## ✨ Features

- Code compilation and build error checks
- Running a linter
- Running automated tests to check code functionality
- Code coverage analysis 
- Executing commands from Makefile (e.g., make build, lint, test)

## 📦 Install

Make sure that Gitea allows interaction with the bot. For this, the `ALLOWED_HOST_LIST` directive should be specified in the configuration with the host where gopher-bot is deployed:

```bash
cat << EOF >> /etc/gitea/app.ini
[webhook]
ALLOWED_HOST_LIST = *
EOF
```
1. For each repository where you want to use gopher-bot, open the webhook settings and create a new Gitea webhook:
    * Target URL: The URL where gopher-bot is deployed with the `/hook` segment (`http://gopher-bot:8080/hook`)
    * HTTP Method: `POST`
    * POST Content Type: `application/json`
    * Secret: the secret your config contains
    * Trigger On: Custom Events...
      * Pull Request Events
        * Pull Request
        * Pull Request Synchronized
    * Branch filter: `*`
    * Active: ✅
2. Add the gopher-bot user with a token and repository rights.
2. Add the gopher-bot user to the repository.
3. Install gopher-bot.

Installation can be done in two ways: via Docker or using a binary.

### Docker

On the machine, execute the commands, specifying gitea_host, the gopher-bot user token, and the secret:

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

### Binary

Build or download the binary from the release page:

```bash
CGO_ENABLED=0 GOOS=linux go build -o gopher-bot cmd/main.go
```

Add the configuration:

```bash
cat << EOF >> /etc/gopher-bot/config.ini
[tokens]
"http://[gitea_host]:[gitea_port]"=gopher-bot:[token]

[server]
DEBUG_MODE=true
SECRET=[secret]
```

Add the binary to the server and run it.
If your project dependencies include tools like golangci-lint, ensure they are installed on the server.

## 🚀 Development

After running `docker compose up`, you need to stop the container.

1. Execute the command:

```bash
docker run -it --rm -v gitea_gitea-config:/etc/gitea busybox sh -c 'cat << EOF >> /etc/gitea/app.ini
[webhook]
ALLOWED_HOST_LIST = *
EOF'
```

2. When setting up Gitea, specify gitea:3000 instead of localhost:3000 as the host. Add the gopher-bot user and configure its token for the repository.


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

3. Add the webhook for the repository.
4. Set SECRET.
5. Set URL to `gopher-bot:8080/hooks`.
6. Add rights fro PR, PR Synchronize.
7. Add the gopher-bot user to the repository.

## How to add gopher-bot to systemd

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

Reload systemd:

```bash
systemctl daemon-reload
```

Add the service to systemd:

```bash
systemctl enable gopher-bot
```
