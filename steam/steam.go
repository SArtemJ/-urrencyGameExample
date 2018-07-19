package main

import (
	"github.com/SArtemJ/CurrencyGameExample/steam/libsteam"
)

func main() {
	app := libsteam.NewApplication()
	app.Configure()
	app.Run()
}
