package main

import (
	"github.com/schedule-rsreu/schedule-api/config"
	"github.com/schedule-rsreu/schedule-api/internal/app"
)

func main() {
	app.Run(config.Get())
}
