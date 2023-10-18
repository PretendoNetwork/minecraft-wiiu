package nex

import (
	secureconnection "github.com/PretendoNetwork/nex-protocols-common-go/secure-connection"
	matchmaking "github.com/PretendoNetwork/nex-protocols-common-go/matchmaking"
	matchmakingext "github.com/PretendoNetwork/nex-protocols-common-go/matchmaking-ext"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-common-go/matchmake-extension"
	nattraversal "github.com/PretendoNetwork/nex-protocols-common-go/nat-traversal"
	"github.com/PretendoNetwork/minecraft-wiiu/globals"
	match_making_types "github.com/PretendoNetwork/nex-protocols-go/match-making/types"

	"fmt"
)

func cleanupSearchMatchmakeSessionHandler(matchmakeSession *match_making_types.MatchmakeSession){
	//matchmakeSession.Attributes[2] = 0
	matchmakeSession.MatchmakeParam = match_making_types.NewMatchmakeParam()
	matchmakeSession.ApplicationData = make([]byte, 0)
	fmt.Println(matchmakeSession.String())
}

func cleanupMatchmakeSessionSearchCriteriaHandler(searchCriteria []*match_making_types.MatchmakeSessionSearchCriteria){
}

func gameSpecificMatchmakeSessionSearchCriteriaChecksHandler(requestSearchCriteria, sessionSearchCriteria *match_making_types.MatchmakeSessionSearchCriteria) bool{
	return true;
}

func createReportDBRecordHandler(pid uint32, reportID uint32, reportData []byte) error {
	return nil
}

func registerCommonSecureServerProtocols() {
	commonSecureconnectionProtocol := secureconnection.NewCommonSecureConnectionProtocol(globals.SecureServer)
	commonSecureconnectionProtocol.CreateReportDBRecord(createReportDBRecordHandler)
	matchmaking.NewCommonMatchMakingProtocol(globals.SecureServer)
	matchmakingext.NewCommonMatchMakingExtProtocol(globals.SecureServer)
	commonMatchmakeExtensionProtocol := matchmakeextension.NewCommonMatchmakeExtensionProtocol(globals.SecureServer)
	commonMatchmakeExtensionProtocol.CleanupSearchMatchmakeSession(cleanupSearchMatchmakeSessionHandler)
	commonMatchmakeExtensionProtocol.CleanupMatchmakeSessionSearchCriteria(cleanupMatchmakeSessionSearchCriteriaHandler)
	commonMatchmakeExtensionProtocol.GameSpecificMatchmakeSessionSearchCriteriaChecks(gameSpecificMatchmakeSessionSearchCriteriaChecksHandler)
	commonMatchmakeExtensionProtocol.DefaultProtocol.GetPlayingSession(GetPlayingSession)
	//commonMatchmakeExtensionProtocol.DefaultProtocol.AutoMatchmakeWithParam_Postpone(autoMatchmakeWithParam_Postpone)
	nattraversal.NewCommonNATTraversalProtocol(globals.SecureServer)
}
