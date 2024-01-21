package main

import (
	"myoidc/internal/app"
)

func main() {
	a := app.Setup()
	a.Run()
}
