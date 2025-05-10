// В этом пакете описаны все типы, использующиеся в калькуляторе.
// Можно использовать для написания клиента.
package types

import "sync"

// Мап задач
type TasksMap map[TaskID]*Task

// Мап выражений пользователя
type ExpressionsMap map[ExpressionID]*Expression

// Статус задачи или выражения
type Status string

// Конкурентный мап задач
type ConcurrentTasksMap struct {
	m TasksMap
	*sync.Mutex
}

// Новый конкурентный мап задач
func NewConcurrentTasksMap() *ConcurrentTasksMap {
	return &ConcurrentTasksMap{make(map[TaskID]*Task), &sync.Mutex{}}
}

// Получение ссылки на задачу
func (cm *ConcurrentTasksMap) Get(id TaskID) *Task {
	cm.Lock()
	res, ok := cm.m[id]
	if !ok {
		t := &Task{}
		cm.m[id] = t
		cm.Unlock()
		return t
	}
	cm.Unlock()
	return res
}

// Добавление ссылки на задачу
func (cm *ConcurrentTasksMap) Add(id TaskID, t *Task) {
	cm.Lock()
	cm.m[id] = t
	cm.Unlock()
}

// Получение самого мапа
func (cm *ConcurrentTasksMap) Map() map[TaskID]*Task {
	return cm.m
}

type UserID = string

// Пользователь
type User struct {
	ID              UserID
	Login, Password string
	Expressions     ExpressionsMap
}

// Запрос на регистрацию/вход
type RegisterLoginRequest struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

// Выражение с ID и UserID
type ExpressionForUser struct {
	ExpressionWithID
	UserID UserID
}

// Выражение
type Expression struct {
	Data   string  `json:"data"`
	Status Status  `json:"status"`
	Result float64 `json:"result"`
}

// Выражение с ID
type ExpressionWithID struct {
	ID ExpressionID `json:"id"`
	Expression
}

// ID выражения
type ExpressionID = uint32

////////////////////////////////////////////////////////////////////////////////////////////////////
// Запросы/ответы

// Запрос обработчика CalculateHandler
type CalculateHandlerRequest struct {
	Expression string `json:"expression"`
}

// Ответ обработчика CalculateHandler
type CalculateHandlerResponse struct {
	ID ExpressionID `json:"id"`
}

// Ответ обработчика GetExpressionHandler
type GetExpressionHandlerResponse struct {
	Expression ExpressionWithID `json:"expression"`
}

// Ответ обработчика GetExpressionsHandler
type GetExpressionsHandlerResponse struct {
	Expressions []ExpressionWithID `json:"expressions"`
}

// Ответ обработчика LoginHandler
type LoginHandlerResponse struct {
	AccessToken string `json:"access_token"`
}

type AccountHandlerResponse struct {
	Username string `json:"username"`
}
