package main

import (
	"example.com/template/infra/startup"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := startup.Startup(); err != nil {
		logrus.Fatalln("Startup failed:", err.Error())
	}
}
