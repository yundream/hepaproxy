package main

import (
	"bitbucket.org/dream_yun/hepaProxy/app"
	"flag"
)

func main() {
	port := flag.String("bind", ":8080", "Bind Port")
	flag.Parse()
	proxy := app.New(*port)
	proxy.Run()
}
