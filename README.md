# Пример на Go

В данном проекте реализовано поднятие БД ClickHouse и скрипта на Go в Docker, запись данных в ClickHouse.


Для запуска проекта его неоходимо скачать командой

git clone https://github.com/gulyasmir/golang-example.git

зайти в директорию golang-example

и выполнить команду  sudo docker-compose up --build

docker exec -it clickhouse clickhouse-client --query "SHOW TABLES FROM test_db;"

Данная команда покажет таблицу logs

После просмотра желательно выполнить команду  sudo docker-compose down -v
чтобы удалить контейнер.