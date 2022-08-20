package realtime

import (
	"os"

	"mize.app/realtime/centrifugo"
)

var CentrifugoController centrifugo.CentrifugoController

func InitialiseCentrifugoController() {
	CentrifugoController = centrifugo.CentrifugoController{BaseUrl: os.Getenv("CENTRIFUGO_SOCKET_URL")}
}

func FetchDefaultChannels() DefaultChannelsType {
	return DefaultChannels
}
