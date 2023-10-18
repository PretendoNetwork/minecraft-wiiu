package nex

import (
	utility "github.com/PretendoNetwork/nex-protocols-go/utility"
	"github.com/PretendoNetwork/minecraft-wiiu/globals"
)

func registerSecureServerNEXProtocols() {
	_ = utility.NewProtocol(globals.SecureServer)

}
