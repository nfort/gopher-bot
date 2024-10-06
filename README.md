# gopher-bot

![alt text](https://github.com/nfort/gopher-bot/blob/main/screenshot.png?raw=true)

## Разработка

Для доступа к хуку нужно добавить

```bash
docker run -it --rm -v gitea_gitea-config:/etc/gitea busybox sh -c 'cat << EOF >> /etc/gitea/app.ini
[webhook]
ALLOWED_HOST_LIST = *
EOF'
```

