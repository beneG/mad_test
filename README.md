В начале необходимо установить необходимые модули коммандой:
go get -u github.com/labstack/echo
go get -u github.com/VividCortex/mysqlerr
go get -u github.com/go-sql-driver/mysql
go get -u github.com/jmoiron/sqlx

#### URI для логина
/api/v1/login
methods: POST
form data: {"login": "nurbek", "password": "password123"}
пример response: {"token": "8b018a17b266c618e15de91e8f53dffa", "errMsg": "OK", "errCode": 0}

При всех остальных запросах в заголовке http запроса должно быть поле ключ-значение:
"Authorization": "Bearer <токен который вы получили при логине>"

#### URI для манипуляции с пользователями (slug - login пользователя)
/api/v1/users/{slug}
methods: GET, PUT, POST, DELETE

#### URI для получения всего списка тасков
/api/v1/tasks
methods: GET

#### URI для манипуляций с тасками
/api/v1/tasks/{task_id}
methods: GET PUT POST DELETE

#### URI для комманд
/api/v1/tasks/{task_id}/{command}
methods:  POST
commands: acquire, finish, accept, close

поле command должно быть в POST-запросе и содерржать одно из вышеперечисленных значений





