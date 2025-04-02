**Земляков Алексей**

[ТГ(для быстрой связи)](https://t.me/zemliakov_25)

[Исходное задание](https://docs.google.com/document/d/1NAL-T_ig4ajPvugDBxSO4xH58ptwhnNhAoFbjc55MVE/edit?tab=t.0)

## Сгенерировать сертификаты
```shell 
make gen-crt
```

### Для работы с https необходимо добавить сертификаты в систему, на linux:
```shell
sudo apt-get install -y ca-certificates
sudo cp ca.crt /usr/local/share/ca-certificates
sudo update-ca-certificates
```

## Старт проекта в докере
```shell
make start
```

## Старт проекта локально вне докера
```shell
make run
```