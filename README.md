# 28.11

## Запуск
```shell
git@github.com:AndreySirin/27.11.git
```
```shell
go run main.go
```
О проекте

Персистентность состояния обеспечивается за счёт встраиваемой базы данных (bbolt).

Реализована основная функциональность:
-создание запроса на проверку доступности интернет-ресурсов.
-получение отчета.

Используемые паттерны:
Producer–Consumer (очередь задач через канал для асинхронной обработки),
Worker Pool (пул воркеров для параллельной выполнения запросов и управления ресурсами),
Repository (абстракция работы с хранилищем Bbolt),
Graceful Shutdown (корректное завершение работы через context).

О работе сервиса:

## Использование API
## Создание запроса:
```shell
метод:POST
URI:http://localhost:8080/api/v1/tasks
 {
  "links": [
	"https://upload.wikimedia.org/wikipedia/commons/3/3f/Fronalpstock_big.jpg",
	"https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf",
	"https://upload.wikimedia.org/wikipedia/commons/6/6e/Golde33443.jpg"
  ]
 }
тело ответа:
{
[id]
}
код ответа:201
```
## Запрос на получение статуса.
```shell
метод:POST
URI:http://localhost:8080/api/v1/report
тело запрса:
 {
     "links_list":[id]
 }
тело ответа: .pdf
код ответа:200
```
