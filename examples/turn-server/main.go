package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/pion/logging"
	"github.com/pion/turn"
)

func createAuthHandler(usersMap map[string]string) turn.AuthHandler {
	return func(username string, srcAddr net.Addr) (string, bool) {
		if password, ok := usersMap[username]; ok {
			return password, true
		}
		return "", false
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	usersMap := map[string]string{}

	users := os.Getenv("USERS")
	if users == "" {
		users = "com=webrtcdemo"
		//log.Panic("USERS is a required environment variable")
	}
	for _, kv := range regexp.MustCompile(`(\w+)=(\w+)`).FindAllStringSubmatch(users, -1) {
		usersMap[kv[1]] = kv[2]
	}
	log.Printf("usersMap=%+v", usersMap)

	realm := os.Getenv("REALM")
	if realm == "" {
		realm = "test.tiga.wang"
		//log.Panic("REALM is a required environment variable")
	}
	log.Printf("Realm=%+v", realm)

	udpPortStr := os.Getenv("UDP_PORT")
	if udpPortStr == "" {
		udpPortStr = "3478"
	}
	udpPort, err := strconv.Atoi(udpPortStr)
	if err != nil {
		log.Panic(err)
	}

	var channelBindTimeout time.Duration
	channelBindTimeoutStr := os.Getenv("CHANNEL_BIND_TIMEOUT")
	if channelBindTimeoutStr != "" {
		channelBindTimeout, err = time.ParseDuration(channelBindTimeoutStr)
		if err != nil {
			log.Panicf("CHANNEL_BIND_TIMEOUT=%s is an invalid time Duration", channelBindTimeoutStr)
		}
	}

	logger := logging.NewDefaultLoggerFactory()
	logger.DefaultLogLevel.Set(logging.LogLevelTrace)
	s := turn.NewServer(&turn.ServerConfig{
		Realm:              realm,
		AuthHandler:        createAuthHandler(usersMap),
		ChannelBindTimeout: channelBindTimeout,
		ListeningPort:      udpPort,
		LoggerFactory:      logger,
		Software:           os.Getenv("SOFTWARE"),
	})

	err = s.Start()
	if err != nil {
		log.Panic(err)
	}

	<-sigs

	err = s.Close()
	if err != nil {
		log.Panic(err)
	}
}
