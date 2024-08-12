package main

import (
	"database/sql"
	"fmt"
	pbfriends "github.com/PretendoNetwork/grpc-go/friends"
	"os"
	"strconv"
	"strings"

	pb "github.com/PretendoNetwork/grpc-go/account"
	"github.com/PretendoNetwork/minecraft-wiiu/globals"
	"github.com/PretendoNetwork/plogger-go"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func init() {
	globals.Logger = plogger.NewLogger()

	var err error

	err = godotenv.Load()
	if err != nil {
		globals.Logger.Warning("Error loading .env file")
	}

	kerberosPassword := os.Getenv("PN_MINECRAFT_KERBEROS_PASSWORD")
	authenticationServerPort := os.Getenv("PN_MINECRAFT_AUTHENTICATION_SERVER_PORT")
	secureServerHost := os.Getenv("PN_MINECRAFT_SECURE_SERVER_HOST")
	secureServerPort := os.Getenv("PN_MINECRAFT_SECURE_SERVER_PORT")
	accountGRPCHost := os.Getenv("PN_MINECRAFT_ACCOUNT_GRPC_HOST")
	accountGRPCPort := os.Getenv("PN_MINECRAFT_ACCOUNT_GRPC_PORT")
	accountGRPCAPIKey := os.Getenv("PN_MINECRAFT_ACCOUNT_GRPC_API_KEY")
	friendsGRPCHost := os.Getenv("PN_MINECRAFT_FRIENDS_GRPC_HOST")
	friendsGRPCPort := os.Getenv("PN_MINECRAFT_FRIENDS_GRPC_PORT")
	friendsGRPCAPIKey := os.Getenv("PN_MINECRAFT_FRIENDS_GRPC_API_KEY")

	if strings.TrimSpace(kerberosPassword) == "" {
		globals.Logger.Warningf("PN_MINECRAFT_KERBEROS_PASSWORD environment variable not set. Using default password: %q", globals.KerberosPassword)
	} else {
		globals.KerberosPassword = kerberosPassword
	}

	globals.InitAccounts()

	if strings.TrimSpace(authenticationServerPort) == "" {
		globals.Logger.Error("PN_MINECRAFT_AUTHENTICATION_SERVER_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(authenticationServerPort); err != nil {
		globals.Logger.Errorf("PN_MINECRAFT_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %s", authenticationServerPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_MINECRAFT_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %s", authenticationServerPort)
		os.Exit(0)
	}

	if strings.TrimSpace(secureServerHost) == "" {
		globals.Logger.Error("PN_MINECRAFT_SECURE_SERVER_HOST environment variable not set")
		os.Exit(0)
	}

	if strings.TrimSpace(secureServerPort) == "" {
		globals.Logger.Error("PN_MINECRAFT_SECURE_SERVER_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(secureServerPort); err != nil {
		globals.Logger.Errorf("PN_MINECRAFT_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %s", secureServerPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_MINECRAFT_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %s", secureServerPort)
		os.Exit(0)
	}

	if strings.TrimSpace(accountGRPCHost) == "" {
		globals.Logger.Error("PN_MINECRAFT_ACCOUNT_GRPC_HOST environment variable not set")
		os.Exit(0)
	}

	if strings.TrimSpace(accountGRPCPort) == "" {
		globals.Logger.Error("PN_MINECRAFT_ACCOUNT_GRPC_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(accountGRPCPort); err != nil {
		globals.Logger.Errorf("PN_MINECRAFT_ACCOUNT_GRPC_PORT is not a valid port. Expected 0-65535, got %s", accountGRPCPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_MINECRAFT_ACCOUNT_GRPC_PORT is not a valid port. Expected 0-65535, got %s", accountGRPCPort)
		os.Exit(0)
	}

	if strings.TrimSpace(accountGRPCAPIKey) == "" {
		globals.Logger.Warning("Insecure gRPC server detected. PN_MINECRAFT_ACCOUNT_GRPC_API_KEY environment variable not set")
	}

	globals.GRPCAccountClientConnection, err = grpc.Dial(fmt.Sprintf("%s:%s", accountGRPCHost, accountGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		globals.Logger.Criticalf("Failed to connect to account gRPC server: %v", err)
		os.Exit(0)
	}

	globals.GRPCAccountClient = pb.NewAccountClient(globals.GRPCAccountClientConnection)
	globals.GRPCAccountCommonMetadata = metadata.Pairs(
		"X-API-Key", accountGRPCAPIKey,
	)

	if strings.TrimSpace(friendsGRPCHost) == "" {
		globals.Logger.Error("PN_MINECRAFT_FRIENDS_GRPC_HOST environment variable not set")
		os.Exit(0)
	}

	if strings.TrimSpace(friendsGRPCPort) == "" {
		globals.Logger.Error("PN_MINECRAFT_FRIENDS_GRPC_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(friendsGRPCPort); err != nil {
		globals.Logger.Errorf("PN_MINECRAFT_FRIENDS_GRPC_PORT is not a valid port. Expected 0-65535, got %s", accountGRPCPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_MINECRAFT_FRIENDS_GRPC_PORT is not a valid port. Expected 0-65535, got %s", accountGRPCPort)
		os.Exit(0)
	}

	if strings.TrimSpace(friendsGRPCAPIKey) == "" {
		globals.Logger.Warning("Insecure gRPC server detected. PN_MINECRAFT_FRIENDS_GRPC_API_KEY environment variable not set")
	}

	globals.GRPCFriendsClientConnection, err = grpc.Dial(fmt.Sprintf("%s:%s", friendsGRPCHost, friendsGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		globals.Logger.Criticalf("Failed to connect to friends gRPC server: %v", err)
		os.Exit(0)
	}

	globals.GRPCFriendsClient = pbfriends.NewFriendsClient(globals.GRPCFriendsClientConnection)
	globals.GRPCFriendsCommonMetadata = metadata.Pairs(
		"X-API-Key", friendsGRPCAPIKey,
	)

	globals.Postgres, err = sql.Open("postgres", os.Getenv("PN_MINECRAFT_POSTGRES_URI"))
	if err != nil {
		globals.Logger.Critical(err.Error())
	}

	globals.Logger.Success("Connected to Postgres!")

}
