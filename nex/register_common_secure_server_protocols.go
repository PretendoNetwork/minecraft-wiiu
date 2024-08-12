package nex

import (
	"github.com/PretendoNetwork/minecraft-wiiu/globals"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commonglobals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/database"
	commonnattraversal "github.com/PretendoNetwork/nex-protocols-common-go/v2/nat-traversal"
	commonsecure "github.com/PretendoNetwork/nex-protocols-common-go/v2/secure-connection"
	nattraversal "github.com/PretendoNetwork/nex-protocols-go/v2/nat-traversal"
	secure "github.com/PretendoNetwork/nex-protocols-go/v2/secure-connection"
	"os"
	"slices"

	commonmatchmaking "github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making"
	commonmatchmakingext "github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making-ext"
	commonmatchmakeextension "github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension"
	matchmaking "github.com/PretendoNetwork/nex-protocols-go/v2/match-making"
	matchmakingext "github.com/PretendoNetwork/nex-protocols-go/v2/match-making-ext"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"

	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
)

// Is this needed? -Ash
func cleanupSearchMatchmakeSessionHandler(matchmakeSession *matchmakingtypes.MatchmakeSession) {
	//_ = matchmakeSession.Attributes.SetIndex(2, types.NewPrimitiveU32(0))
	matchmakeSession.MatchmakeParam = matchmakingtypes.NewMatchmakeParam()
	matchmakeSession.ApplicationBuffer = types.NewBuffer(make([]byte, 0))
	globals.Logger.Info(matchmakeSession.String())
}

func CreateReportDBRecord(_ *types.PID, _ *types.PrimitiveU32, _ *types.QBuffer) error {
	return nil
}

// * Minecraft WiiU edition isn't always safe to play on public matches. To mitigate this, just claim there are no
// * public matches.
func stubBrowseMatchmakeSession(err error, packet nex.PacketInterface, callID uint32, _ *matchmakingtypes.MatchmakeSessionSearchCriteria, _ *types.ResultRange) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	lstGathering := types.NewList[*types.AnyDataHolder]()
	lstGathering.Type = types.NewAnyDataHolder()

	// * Don't include any sessions!
	//for _, session := range sessions {
	//	matchmakeSessionDataHolder := types.NewAnyDataHolder()
	//	matchmakeSessionDataHolder.TypeName = types.NewString("MatchmakeSession")
	//	matchmakeSessionDataHolder.ObjectData = session.GameMatchmakeSession.Copy()
	//
	//	lstGathering.Append(matchmakeSessionDataHolder)
	//}

	rmcResponseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())

	lstGathering.WriteTo(rmcResponseStream)

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCSuccess(endpoint, rmcResponseBody)
	rmcResponse.ProtocolID = matchmakeextension.ProtocolID
	rmcResponse.MethodID = matchmakeextension.MethodBrowseMatchmakeSession
	rmcResponse.CallID = callID

	return rmcResponse, nil
}

func gameSpecificCanJoinMatchmakeSession(manager *commonglobals.MatchmakingManager, pid *types.PID, session *matchmakingtypes.MatchmakeSession) *nex.Error {
	if !session.OpenParticipation.Value {
		return nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Gathering is not open to new participants")
	}

	isPublic := false
	attrib, err := session.Attributes.Get(0)
	if err == nil {
		// * I wish this was a joke. top 8 bits are GameMode
		// * This is the only difference between a public and a Friends match
		isPublic = (attrib.Value & 0xFFFFFF) == 0x30881
	}

	if isPublic && os.Getenv("PN_MINECRAFT_ALLOW_PUBLIC_MATCHMAKING") == "1" {
		//globals.Logger.Info("Game is public")
		return nil
	}

	host := session.OwnerPID
	hostFriends := manager.GetUserFriendPIDs(host.LegacyValue())
	if slices.Contains(hostFriends, pid.LegacyValue()) {
		//globals.Logger.Info("User is friend of host")
		return nil
	}

	isFriendsOfFriends := false
	if len(session.ApplicationBuffer.Value) > 0xc3 {
		isFriendsOfFriends = session.ApplicationBuffer.Value[0xc3] == 0x8F
	}

	if !isFriendsOfFriends {
		return nex.NewError(nex.ResultCodes.RendezVous.NotFriend, "User is not a friend of host")
	}

	// * Get the participants of this gathering so we don't have to check all 100whatever of host's friends
	_, _, participants, _, nerr := database.FindGatheringByID(manager, session.ID.Value)
	if err != nil {
		globals.Logger.Errorf("Can't find gathering for pariticpation check: %v", nerr)
		return nerr
	}

	for _, friend := range hostFriends {
		// * Make sure this friend is actually in-game
		// * This cast feels bad
		if slices.Contains(participants, uint64(friend)) {
			// * Are you a friend of the host's friend?
			friendsFriends := manager.GetUserFriendPIDs(friend)
			if slices.Contains(friendsFriends, pid.LegacyValue()) {
				//globals.Logger.Infof("User is friend of host's friend %v", friend)
				return nil
			}
		}
	}

	return nex.NewError(nex.ResultCodes.RendezVous.NotFriend, "User is not a friend of host's friends")
}

func registerCommonSecureServerProtocols() {
	secureProtocol := secure.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(secureProtocol)
	commonSecureProtocol := commonsecure.NewCommonProtocol(secureProtocol)

	commonSecureProtocol.CreateReportDBRecord = CreateReportDBRecord

	natTraversalProtocol := nattraversal.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(natTraversalProtocol)
	commonnattraversal.NewCommonProtocol(natTraversalProtocol)

	matchMakingProtocol := matchmaking.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMakingProtocol)
	commonMatchMakingProtocol := commonmatchmaking.NewCommonProtocol(matchMakingProtocol)
	commonMatchMakingProtocol.SetManager(globals.MatchmakingManager)

	matchMakingExtProtocol := matchmakingext.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMakingExtProtocol)
	commonMatchMakingExtProtocol := commonmatchmakingext.NewCommonProtocol(matchMakingExtProtocol)
	commonMatchMakingExtProtocol.SetManager(globals.MatchmakingManager)

	matchmakeExtensionProtocol := matchmakeextension.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchmakeExtensionProtocol)
	commonMatchmakeExtensionProtocol := commonmatchmakeextension.NewCommonProtocol(matchmakeExtensionProtocol)
	commonMatchmakeExtensionProtocol.SetManager(globals.MatchmakingManager)

	globals.MatchmakingManager.GetUserFriendPIDs = globals.GetUserFriendPIDs
	globals.MatchmakingManager.CanJoinMatchmakeSession = gameSpecificCanJoinMatchmakeSession

	commonMatchmakeExtensionProtocol.CleanupSearchMatchmakeSession = cleanupSearchMatchmakeSessionHandler
	if os.Getenv("PN_MINECRAFT_ALLOW_PUBLIC_MATCHMAKING") != "1" {
		globals.Logger.Warning("Public minigames are disabled for safety reasons.")
		globals.Logger.Warning("To enable public matches, set PN_MINECRAFT_ALLOW_PUBLIC_MATCHMAKING=1.")
		matchmakeExtensionProtocol.SetHandlerBrowseMatchmakeSession(stubBrowseMatchmakeSession)

		// * Make sure any unused MM protocols aren't able to show sessions
		matchmakeExtensionProtocol.SetHandlerAutoMatchmakePostpone(nil)
		matchmakeExtensionProtocol.SetHandlerAutoMatchmakeWithParamPostpone(nil)
		matchmakeExtensionProtocol.SetHandlerAutoMatchmakeWithSearchCriteriaPostpone(nil)
		matchMakingProtocol.SetHandlerFindByType(nil)
		matchMakingProtocol.SetHandlerFindByDescription(nil)
		matchMakingProtocol.SetHandlerFindByDescriptionRegex(nil)
		matchMakingProtocol.SetHandlerFindByID(nil)
		//matchMakingProtocol.SetHandlerFindBySingleID(nil) Used for friends matchmaking
		matchMakingProtocol.SetHandlerFindByOwner(nil)
		matchMakingProtocol.SetHandlerFindByParticipants(nil)
		matchMakingProtocol.SetHandlerFindInvitations(nil)
		matchMakingProtocol.SetHandlerFindBySQLQuery(nil)
	}
}
