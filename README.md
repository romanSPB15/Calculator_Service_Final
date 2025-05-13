# Многопользовательский веб-сервис для вычисления арифметических выражений на Golang

## Описание

Программа представляет собой API-сервис для конкурентного вычисления выражений.<br>
Для связи агента(сервиса вычисляющего задачи как 2+8, 10-7) и оркестратора(разбивающий (2+8)-7 на задачи) используется GRPC на порту 8008.

## Предназначение

Пользователь хочет считать арифметические выражения. Он вводит строку `2 + 2 * 2` и хочет получить в ответ `6`.<br>
Но наши операции сложения и умножения (также деления и вычитания) выполняются "очень-очень" долго.<br>
Поэтому вариант, при котором пользователь делает *HTTP-запрос* и получает в качетсве ответа результат, невозможна.<br>
Более того, вычисление каждой такой операции в нашей "альтернативной реальности" занимает *"гигантские" вычислительные мощности*.<br>
Соответственно, каждое действие мы должны уметь выполнять отдельно и масштабировать эту систему можем добавлением вычислительных мощностей в нашу систему в виде новых "машин".<br>
Поэтому пользователь может с какой-то периодичностью уточнять у сервера "не посчиталость ли выражение"?<br>
Если выражение наконец будет вычислено - то он получит результат.<br>

## Установка

 - Для установки нужно выбрать директорию, где будет проект проекта:
```bash
cd <your_dir>
```
 - Потом необходимо выполнить эту команду:
```bash
git clone https://github.com/romanSPB15/Calculator_Service_Final
```
 - В выбранной папке появится папка ```Calculator_Service_Final``` c проектом.

## Работа с API
### Переменные среды
Сначала необходимо открыть файл ```./config/.env``` и установить параметры:

 - **TIME_ADDITION_MS** - время вычисления сложения(в миллисекундах);

 - **TIME_SUBTRACTION_MS** - время вычисления вычитания;

 - **TIME_MULTIPLICATIONS_MS** - время вычисления умножения;

 - **TIME_DIVISIONS_MS** - время вычисления деления;

 - **COMPUTING_POWER** - максмальное количество *worker*'ов, которые параллельно выполняют арифметические действия.

 - **HOST** - хост сервера(по умолчанию localhost)

 - **PORT** - порт сервера(по умолчанию 8080)

    - **ВАЖНО** - Порт не должен быть равен 8008 - на нём работает GRPC сервер!

 - **DEBUG** - отладка(вывод событий в лог)

 - **WEB** - включение веб-интерфейса, по умолчанию включен

### Запуск
 - Для запуска API необходимо выбрать директорию проекта:
```
cd <путь к папке Calculator_Service_Final>
```
 - Далее надо запустить файл ```cmd/main.go```:
```
go run cmd/main.go
```

### Управление

Для работы через терминал на **Curl** настоятельно рекомендую использовать **Git Bash**.<br> Но можно использовать и **Postman** - приложение для отправки *HTTP-запросов*, по-моему в нём работать проще.


#### Тесты
Для запуска тестов нужно выполнить такую команду:
```
go test ./internal/application
```

#### Регистрация и вход
##### Регистрация

###### Curl
- Здесь и далее предполагается, что **HOST**=localhost, а **PORT**=8080 - по умолчанию
```
curl --location 'http://localhost:8080/api/v1/register' \ // 
--header 'Content-Type: application/json' \
--data '{
  "login": <логин>,
  "password": <пароль>,
}'
```

###### Postman

 - **URL** http://localhost:8080/api/v1/register;
 - Запрос **POST**;
 - **Body** **RAW** { "login": <логин>, "password": <пароль>};
 - Нажать на ****SEND****.

--------------------------------------------------

###### Как работает

Пользователь отправляет запрос **POST** на ```/api/v1/register``` с телом:
```
{ "login": <логин>, "password": <пароль>}
```

В ответ получает 200 OK (в случае успеха), а 422 в случае ошибки:

- **Неправильный метод**(должен быть только **POST**): ```method not allowed```.
- **Неправильное тело запроса:** ```invalid body```.
- **Такой пользователь уже существует:** ```user already exists```.
- **Короткий пароль** (меньше *5 символов*): ```short password```.

Ошибка будет в теле ответа.

##### Вход

###### Curl
```
curl --location 'http://localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": <логин>,
  "password": <пароль>,
}'
```

###### Postman

 - **URL** http://localhost:8080/api/v1/login;
 - Запрос **POST**;
 - **Body** **RAW** { "login": <логин>, "password": <пароль>};
 - Нажать на ****SEND****.

-------------------------------------------------------------

###### Как работает


Пользователь отправляет запрос **POST** на ```/api/v1/login``` с телом:
```
{ "login": <логин>, "password":<пароль> }
```
- Тело ответа:
```
{"access_token": <токен>}
```

В ответ получает 200 OK и **JWT-токен для последующей авторизации**, а 422 в случае ошибки.

- **Неправильный метод**(должен быть **только POST**): ```method not allowed```
- **Неправильное тело запроса:** ```invalid body```
- **Пользователь не найден:** ```invalid login or password```

Ошибка будет в теле ответа.


---------------------------------------------------------------------
Для всех операций, кроме **login** и **register**, нуже заголовок заголовок Authorization:
```
Authorization: Bearer <полученный при login токен>
```

 - Если заголовок не найден, или отсутствует слово Bearer, то возвращает ошибку ```invalid header```.
 - Если токен неправильный, то возвращает ошибку ```invalid token```.

 #### Добавление арифметического выражения для вычисления

Добавление выражения для вычисления в **API**.

##### Curl
```
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": <выражение>
}'
```
##### Postman
 - **URL** http://localhost:8080/api/v1/calculate;
 - Запрос **POST**;
 - **Body** **RAW** {"expression": <выражение>};
 - Нажать на ****SEND****.

##### Коды ответа: 
 - 201 - выражение принято для вычисления
 - 422 - невалидные данные
 - 500 - что-то пошло не так

##### Тело ответа

```
{
    "id": <уникальный идентификатор выражения> // его ID
}
```
#### Получение списка выражений
##### Curl
```
curl --location 'http://localhost:8080/api/v1/expressions'
```
##### Postman
 - **URL** http://localhost:8080/api/v1/expressions;
 - Запрос **GET**;
 - **Body** **NONE**;
 - Нажать на ****SEND****.

##### Тело ответа:

Получение всех сохранённых выражений(**ID** не нужен).

```
{
    "expressions": [
        {
            "id": 2352351,
            "status": "OK",  // готово
            "result": 3
            "error": ""
        },
        
            "id": 5372342,
            "status": "Calculation",  // в процессе
            "result": 3
            "error": ""
        },
        {
            "id": 8251431,
            "status": "error",  // статус ошибки
            "result": 0
            "error": "someting error"
        },
        {
            "id": 34942763,
            "status": "Wait", // ждёт выполнения
            "result": 0
            "error": ""
        }
    ]
}
```
##### Коды ответа:
 - 200 - успешно получен список выражений
 - 500 - что-то пошло не так

#### Получение выражения по его идентификатору

Получение выражения по его идентификатору.

*Примечание:* Для того, чтобы получить выражение по его ID, необходимо сохранить полученный при **Добавление вычисления арифметического выражения** индитефикатор.

##### Curl

```
curl --location 'http://localhost:8080/api/v1/expressions/<id выражения>'
```

##### Postman:
 - **URL** http://localhost:8080/api/v1/expressions/<id выражения>;
 - Запрос **GET**;
 - Тело **NONE**;
 - Нажать на ****SEND****.

##### Тело ответа:

```
{
    "expression":
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
}
```

##### Коды ответа:
 - 200 - успешно получено выражение
 - 404 - нет такого выражения
 - 500 - что-то пошло не так

## Примеры работы с API

### Простой пример

#### Делаем запрос на регистрацию

##### Curl

```
curl --location 'http://localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "user0",
  "password": "user0_password"
}'
```

##### Postman

 - **URL** http://localhost:8080/api/v1/register;
 - Запрос **POST**;
 - **Body** **RAW** ```{"login": user0,"password": user0_password}```;
 - Нажать на ****SEND****.

#### Делаем запрос на вxод

##### Curl

```
curl --location 'http://localhost:8080/api/v1/login' \ // 
--header 'Content-Type: application/json' \
--data '{
  "login": "user0",
  "password": "user0_password"
}'
```

##### Postman

 - **URL** localhost:8080/api/v1/login;
 - Запрос **POST**;
 - **Body** **RAW** ```{"login": user0,"password": user0_password}```;
 - Нажать на ****SEND****.

--------------------------------------------------------

- Считываем из тела ответа токен.

#### Делаем запрос на вычисление выражения

##### Curl
```
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' \
--data '{
  "expression": "2+2/2"
}'
```
##### Postman
 - **URL** http://localhost:8080/api/v1/calculate;
 - Запрос **POST**;
 - Header **Authorization** установить в "Bearer <токен>"
 - **Body** **RAW** {"expression": "2+2/2"};
 - Нажать на ****SEND****.

##### Ответ
Статус 201(успешно создано);
```
{
    "id": 12345 // пример
}
```

#### Получаем наше выражение
##### Curl
```
curl --location 'localhost:8080/api/v1/expressions/12345'
```
##### Postman
 - **URL** localhost:8080/api/v1/expressions/12345;
 - Запрос **GET**;
 - Header **Authorization** установить в "Bearer <токен>"
 - Тело **NONE**;
 - Нажать на ****SEND****

##### Ответ
Статус 200(успешно получено);
```
{
    "expression":
        {
            "id": 12345,
            "status": "OK",
            "result": 321
        }
}
```

#### Получаем все выражения
##### Curl
```
curl --location 'http://localhost:8080/api/v1/expressions'
--header 'Authorization: Bearer <токен>'
```
##### Postman
 - **URL** http://localhost:8080/api/v1/expressions;
 - Запрос **GET**;
 - Header **Authorization** установить в "Bearer <токен>"
 - **Body** **NONE**;
 - Нажать на ****SEND****.

##### Ответ
Статус 200(успешно получены);
```
{
    "expressions": [
        {
            "id": 12345,
            "status": "OK",
            "result": 321
        },
    ]
}
```

### Пример с ошибкой в запросе №1

#### Делаем **неправильный** запрос на вычисление выражения

##### Curl

```
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' \
--data '{
  "radhgsags": "2+2/2"
}'
```

##### Postman
 - **URL** http://localhost:8080/api/v1/calculate;
 - Запрос **POST**;
 - Header **Authorization** установить в "Bearer <токен>"
 - **Body** **RAW** {"radhgsags": "2+2/2"};
 - Нажать на ****SEND****.

##### Ответ
Статус 422(**неправильный** запрос);


### Пример с ошибкой в запросе №2

#### Делаем **правильный** запрос на вычисление выражения

##### Curl
```
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' \
--data '{
  "expression": "2+2/2"
}'
```


##### Postman
 - **URL** http://localhost:8080/api/v1/calculate;
 - Запрос **POST**;
 - Header **Authorization** установить в "Bearer <токен>"
 - **Body** **RAW** {"expression": "2+2/2"};
 - Нажать на ****SEND****.

##### Ответ
Статус 201(успешно создано);
```
{
    "id": 12345 // пример
}
```
#### Далее получаем наше выражение(**неправильный** ID)
##### Curl
```
curl --location 'http://localhost:8080/api/v1/expressions/45362'
```

##### Postman:
 - **URL** http://localhost:8080/api/v1/expressions/45362;
 - Запрос **GET**;
 - Header **Authorization** установить в "Bearer <токен>"
 - Тело **NONE**;
 - Нажать на ****SEND****.

##### Ответ
Статус 404(не найдено);


### Пример с ошибкой в запросе №3

#### Делаем запрос с **некорректным** URL на вычисление выражения

##### Curl
```
curl --location 'http://localhost:8080/api/v1/abc' \
--header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' \
--data '{
  "expression": "121+2"
}'
```
##### Postman
 - **URL** http://localhost:8080/api/v1/abc;
 - Запрос **POST**;
 - Header **Authorization** установить в "Bearer <токен>"
 - **Body** **RAW** {"expression": "121+2"};
 - Нажать на ****SEND****.

##### Ответ
Статус 404(**NOT FOUND**);

### Пример с ошибкой в запросе №4

#### Register

##### Curl
```
curl --location 'http://localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "abc": "def"
}'
```

##### Postman
 - **URL** http://localhost:8080/api/v1/register;
 - Запрос **POST**;
 - **Body** **RAW** { "abc": "def"};
 - Нажать на ****SEND****.

##### Ответ
```
invalid body
```

#### Login

##### Curl
```
curl --location 'http://localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "abc": "def"
}'
```

##### Postman
 - **URL** http://localhost:8080/api/v1/login;
 - Запрос **POST**;
 - **Body** **RAW** { "abc": "def"};
 - Нажать на ****SEND****.

##### Ответ
```
invalid body
```

### Веб-интерфейс


 - [Главная страница](http://localhost:8080/api/v1/web).

 - Использует cookies.