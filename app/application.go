package app

import (
	"bitbucket.org/dream_yun/hepaProxy/handler"
	"net/http"
)

type Application struct {
	port string
}

func New(port string) *Application {
	return &Application{port: port}
}

func (a *Application) Run() {
	h := handler.New()
	h.Regist()
	err := http.ListenAndServe(a.port, h.Router())
	if err != nil {
		panic(err)
	}
}
