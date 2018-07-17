package main

import (
	"github.com/SArtemJ/CurrencyGameExample/currency/libcurrency"
)

func main() {
	app := libcurrency.NewApplication()
	app.Configure()
	app.Run()
}
