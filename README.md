
---
Build сервиса:
```
make build
```

---
Запуск сервиса:
```
make launch
```

---
Запуск сервиса (docker-compose):
```
$ cd /tmp
$ git clone --branch develop git@github.com:spendmail/s3_previewer.git s3_previewer
$ cd previewer
$ make run
```

---
Проверка работы:
```
wget http://localhost:8888/resize/1024/0/bucket_name/key.jpg

