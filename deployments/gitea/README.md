Добавить для Gitea виртуальную сеть.

![image](https://user-images.githubusercontent.com/1002000/114235710-0f9c9b80-7e0a-11eb-9c6c-6f9a6b8c9e0b.png)

## Конфигурация

Для доступа к хуку нужно добавить

```bash
docker run -it --rm -v gitea_gitea-config:/etc/gitea busybox sh -c 'cat << EOF >> /etc/gitea/app.ini
[webhook]
ALLOWED_HOST_LIST = *
EOF'
```
