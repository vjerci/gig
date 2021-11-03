package settings

import "os"

type RabbitSettings struct {
	Address string
	Queue   string
}

type HttpSettings struct {
	Address string
}

var Rabbit RabbitSettings
var Http HttpSettings

func Load() {
	Rabbit = RabbitSettings{
		Address: os.Getenv("RABBIT"),
		Queue:   os.Getenv("RABBIT_QUEUE"),
	}

	Http = HttpSettings{
		Address: ":" + os.Getenv("HTTP_PORT"),
	}
}
