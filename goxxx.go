// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.

// Main package for the goxxx project
//
// For the details see the file goxxx.go, as godoc won't show the documentation for the main package.
package main

import (
	"flag"
	"fmt"
	"github.com/vaz-ar/cfgFlags"
	"github.com/vaz-ar/goxxx/core"
	"github.com/vaz-ar/goxxx/database"
	"github.com/vaz-ar/goxxx/help"
	"github.com/vaz-ar/goxxx/invoke"
	"github.com/vaz-ar/goxxx/memo"
	"github.com/vaz-ar/goxxx/pictures"
	"github.com/vaz-ar/goxxx/quote"
	"github.com/vaz-ar/goxxx/search"
	"github.com/vaz-ar/goxxx/webinfo"
	"github.com/vaz-ar/goxxx/xkcd"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Version and build time
var (
	GlobalVersion string
	BuildTime     string
)

const (
	// Equivalent to enums (cf. https://golang.org/ref/spec#Iota)
	flagsExit    = iota //  == 0
	flagsSuccess        //  == 1
	flagsFailure        //  == 2
	flagsAddUser        //  == 3
)

// Config struct
type configData struct {
	channel       string
	channelKey    string
	nick          string
	server        string
	modules       []string
	debug         bool
	emailServer   string
	emailPort     int
	emailSender   string
	emailAccount  string
	emailPassword string
	admins        []string
}

// getOptions processes the command line arguments
func getOptions() (config configData, returnCode int) {
	// IRC
	flag.StringVar(&config.channel, "channel", "", "IRC channel name")
	flag.StringVar(&config.channelKey, "key", "", "IRC channel key (optional)")
	flag.StringVar(&config.nick, "nick", "goxxx", "the bot's nickname (optional)")
	flag.StringVar(&config.server, "server", "chat.freenode.net:6697", "IRC_SERVER[:PORT] (optional)")
	admins := flag.String("admin", "", "Administrators nick (separated by commas)")
	modules := flag.String("modules", "memo,webinfo,invoke,search,xkcd,pictures,quote", "Modules to enable (separated by commas)")
	// Email
	flag.StringVar(&config.emailServer, "email_server", "", "SMTP server address")
	flag.IntVar(&config.emailPort, "email_port", 0, "SMTP server port")
	flag.StringVar(&config.emailSender, "email_sender", "", "Email address to use in the \"From\" part of the header")
	flag.StringVar(&config.emailAccount, "email_account", "", "Email address from which to send emails")
	flag.StringVar(&config.emailPassword, "email_pwd", "", "password for the SMTP server")
	// Application
	flag.BoolVar(&config.debug, "debug", false, "Debug mode")
	version := flag.Bool("version", false, "Display goxxx version")

	flag.Usage = func() {
		fmt.Println("Usage:", os.Args[0], "-channel CHANNEL [ARGUMENTS]")
		fmt.Println()
		fmt.Println("Arguments description:")
		flag.PrintDefaults()
		fmt.Println("\nCommands description:")
		fmt.Println("  add_user <nick> <email>: Add an user to the database\n")
	}

	// Hybrid config: use flags and INI file
	// Command line flags take precedence on INI values
	if err := cfgFlags.Parse(); err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	config.modules = strings.Split(*modules, ",")
	config.admins = strings.Split(*admins, ",")

	if *version {
		fmt.Printf("\nGoxxx version: %s\n\n", GlobalVersion)
		returnCode = flagsExit
		return
	}

	lenArgs := len(flag.Args())
	// add_user command
	if lenArgs > 0 && flag.Args()[0] == "add_user" {
		if lenArgs != 3 {
			flag.Usage()
			returnCode = flagsFailure
			return
		}
		returnCode = flagsAddUser
	} else if config.channel == "" {
		flag.Usage()
		returnCode = flagsFailure
	} else {
		returnCode = flagsSuccess
	}
	return
}

func main() {
	// Set log output to a file
	logFile, err := os.OpenFile("./goxxx_logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	config, returnCode := getOptions()
	if returnCode == flagsExit {
		return
	} else if returnCode == flagsFailure {
		log.Fatal("Initialisation failed (getOptions())")
	}
	if config.debug {
		// In debug mode we show the file name and the line from where the log come from
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create the database
	db := database.NewDatabase("", false)
	defer db.Close()

	// Process commands if necessary
	if returnCode == flagsAddUser {
		if err := database.AddUser(flag.Args()[1], flag.Args()[2]); err == nil {
			fmt.Println("User added to the database")
		} else {
			fmt.Printf("\nadd_user error: %s\n", err)
		}
		return
	}

	// Create the bot
	bot := core.NewBot(config.nick, config.server, config.channel, config.channelKey)

	// Initialise packages
	for _, module := range config.modules {
		switch strings.TrimSpace(module) {
		case "invoke":
			if !invoke.Init(db, config.emailSender, config.emailAccount, config.emailPassword, config.emailServer, config.emailPort) {
				log.Println("Error while initialising invoke package")
				continue
			}
			cmd := invoke.GetCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToNick)
			help.AddMessages(cmd)
			log.Println("invoke module loaded")

		case "memo":
			memo.Init(db)
			bot.AddMsgHandler(memo.SendMemo, bot.ReplyToNick)

			cmd := memo.GetMemoCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)

			cmd = memo.GetMemoStatCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToNick)
			help.AddMessages(cmd)
			log.Println("memo module loaded")

		case "search":
			cmd := search.GetDuckduckGoCmd()
			bot.AddCmdHandler(cmd, bot.Reply)
			help.AddMessages(cmd)

			cmd = search.GetWikipediaCmd()
			bot.AddCmdHandler(cmd, bot.Reply)
			help.AddMessages(cmd)

			cmd = search.GetWikipediaFRCmd()
			bot.AddCmdHandler(cmd, bot.Reply)
			help.AddMessages(cmd)

			cmd = search.GetUrbanDictionnaryCmd()
			bot.AddCmdHandler(cmd, bot.Reply)
			help.AddMessages(cmd)
			log.Println("search module loaded")

		case "webinfo":
			webinfo.Init(db)
			bot.AddMsgHandler(webinfo.HandleURLs, bot.ReplyToAll)

			cmd := webinfo.GetCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)
			log.Println("webinfo module loaded")

		case "xkcd":
			cmd := xkcd.GetCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)
			log.Println("xkcd module loaded")

		case "pictures":
			pictures.Init(db, config.admins)

			cmd := pictures.GetPicCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)

			cmd = pictures.GetAddPicCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)

			cmd = pictures.GetRmPicCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)
			log.Println("pictures module loaded")

		case "quote":
			quote.Init(db, config.admins)
			bot.AddMsgHandler(quote.HandleMessages, nil)

			cmd := quote.GetQuoteCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)

			cmd = quote.GetRmQuoteCommand()
			bot.AddCmdHandler(cmd, bot.ReplyToAll)
			help.AddMessages(cmd)
			log.Println("quote module loaded")

		default:
		}
	}
	bot.AddCmdHandler(help.GetCommand(), bot.ReplyToNick)

	log.Println("Goxxx started")

	// Go signal notification works by sending os.Signal values on a channel.
	// We'll create a channel to receive these notifications
	// (we'll also make one to notify us when the program can exit).
	interruptSignals := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// signal.Notify registers the given channel to receive notifications of the specified signals.
	signal.Notify(interruptSignals, syscall.SIGINT, syscall.SIGTERM)

	// This goroutine executes a blocking receive for signals.
	// When it gets one it'll print it out and then notify the program that it can finish.
	go func() {
		sig := <-interruptSignals
		log.Printf("System signal received: %s\n", sig)
		done <- true
	}()

	// Start the bot
	go bot.Run()

	// The current routine will be blocked here until done is true
	<-done

	// Close the bot connection and the database
	bot.Stop()
	db.Close()

	log.Println("Goxxx exiting")
}
