## Как развернуть локально

После запуска docker compose up, нужно остановить

1. Выполнить команды

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
docker run -it --rm -v gopher-bot_config:/etc/gopher-bot busybox sh -c 'cat << EOF >> /etc/gopher-bot/config.ini
[tokens]
"http://gitea:3000"=gopher-bot:[token]

[server]
DEBUG_MODE=true
SECRET=iNeydroTioUC'
```

3. Добавить webhook для репозитория
4. Указать SECRET
5. В качестве хоста указать `gopher-bot:8080/hooks`
6. Дать права на PR, PR Synchronize
