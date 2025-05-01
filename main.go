package main

import (
	"bufio"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
)

var (
	dg           *discordgo.Session
	err          error
	fileInputPtr string
	targetMap    map[string]Target // Putting each line of the csv in a list/array
	teams        []string          // Special list of the team number: Used for ...
	hostnameList []string
	osList       []string
)

type Config struct {
	BotToken    string `env:"BotToken,required"`
	ServerID    int    `env:"ServerID,required"`
	PwnboardURL int    `env:"PwnboardURL"`
}

type Target struct {
	ip         string
	teamstring string
	hostname   string
	ostype     string
}

type PwnBoard struct {
	IPs  string `json:"ip"`
	Type string `json:"type"`
}

func init() {
	flag.StringVar(&fileInputPtr, "f", "", "This csv should contains the list of targets: ip,team#,hostname,ostype")
	flag.Parse()

	if fileInputPtr == "" {
		log.Fatal("No file specified")
		os.Exit(0)
	}

	log.SetOutput(os.Stdout)
	log.Info("Target file: " + fileInputPtr)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("unable to load .env file: %e", err)
	}
	cfg := Config{}
	err = env.Parse(&cfg)
	if err != nil {
		log.Fatalf("unable to parse ennvironment variables: %e", err)
	}

	dg, err = discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Error("error creating Discord session,", err)
		return
	}
	log.Info("Bot Connected")
}

func newChannel(dg *discordgo.Session, channel *discordgo.ChannelCreate) {
	log.Infof("New channel created: ID=%s, Name=%s", channel.ID, channel.Name)
}

func main() {
	targetMap, teams, hostnameList, osList = parseCSV(fileInputPtr)

	dg.AddHandler(newChannel)

	err = dg.Open()
	if err != nil {
		return
	}

	// user, err := dg.User("@me")
	// if err != nil {
	// 	log.Fatalf("Failed to get bot user: %v", err)
	// }

	// Register slash commands
	// for _, v := range commands {
	// 	_, err := dg.ApplicationCommandCreate(user.ID, util.ServerID, v)
	// 	if err != nil {
	// 		log.Panicf("Cannot create '%v' command: %v", v.Name, err)
	// 	}
	// }

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// Parsing the input csv file and creates a list that will used in other parts of the code
func parseCSV(csvName string) (map[string]Target, []string, []string, []string) {
	var list = make(map[string]Target)
	f, err := os.Open(csvName)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	teamnum := []string{}
	hostnameList := []string{}
	osList := []string{}
	for s.Scan() {
		lineBuff := s.Bytes()
		v := strings.Split(string(lineBuff), ",")
		list[v[0]] = Target{
			ip:         v[0],
			teamstring: v[1],
			hostname:   v[2],
			ostype:     v[3],
		}
		teamnum = append(teamnum, v[1])
		hostnameList = append(hostnameList, v[2])
		osList = append(osList, v[3])
	}
	return list, teamnum, hostnameList, osList
}
