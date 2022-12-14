package realtime

import (
	"os"

	"mize.app/logger"
	"mize.app/realtime/centrifugo"
)

var CentrifugoController centrifugo.CentrifugoController

func InitialiseCentrifugoController() {
	CentrifugoController = centrifugo.CentrifugoController{BaseUrl: os.Getenv("CENTRIFUGO_SOCKET_URL")}
	logger.Info("realtime (centrifugo) - ONLINE")
}

func FetchDefaultChannels() DefaultChannelsType {
	return DefaultChannels
}
