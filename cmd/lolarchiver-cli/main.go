package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/user1/lolarchiver-cli/pkg/api"
	"github.com/user1/lolarchiver-cli/pkg/config"
)

const (
	version = "1.0.0"
)

func main() {
	creditsCmd := flag.NewFlagSet("credits", flag.ExitOnError)
	youtubeCmd := flag.NewFlagSet("youtube", flag.ExitOnError)
	twitterCmd := flag.NewFlagSet("twitter", flag.ExitOnError)
	twitchCmd := flag.NewFlagSet("twitch", flag.ExitOnError)
	kickCmd := flag.NewFlagSet("kick", flag.ExitOnError)
	reverseCmd := flag.NewFlagSet("reverse", flag.ExitOnError)
	databaseCmd := flag.NewFlagSet("database", flag.ExitOnError)
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "credits":
		creditsCmd.Parse(os.Args[2:])
		handleCredits(creditsCmd)
	case "youtube":
		youtubeCmd.Parse(os.Args[2:])
		handleYouTube(youtubeCmd)
	case "twitter":
		handle := twitterCmd.String("handle", "", "Twitter handle")
		id := twitterCmd.Int64("id", 0, "Twitter user ID")
		byOld := twitterCmd.Bool("by-old", false, "Search by old usernames")
		
		if err := twitterCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			twitterCmd.PrintDefaults()
			os.Exit(1)
		}

		if *handle == "" && *id == 0 {
			fmt.Println("Error: Either handle or id must be provided")
			twitterCmd.PrintDefaults()
			os.Exit(1)
		}

		done := make(chan bool)
		go func() {
			startTime := time.Now()
			for {
				select {
				case <-done:
					return
				default:
					elapsed := time.Since(startTime).Round(time.Second)
					fmt.Fprintf(os.Stderr, "\rProcessing (%v)...", elapsed)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()

		client, err := getClient()
		if err != nil {
			done <- true
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.TwitterHistoryLookup(*handle, *id, *byOld)
		if err != nil {
			done <- true
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}

		done <- true
		fmt.Fprint(os.Stderr, "\r")

		if len(resp.Body) == 0 {
			fmt.Println("No data found")
		} else {
			fmt.Println(string(resp.Body))
		}
	case "twitch":
		handleTwitch(twitchCmd)
	case "kick":
		handleKick(kickCmd)
	case "reverse":
		reverseCmd.Parse(os.Args[2:])
		handleReverse(reverseCmd)
	case "database":
		handleDatabase(databaseCmd)
	case "config":
		configCmd.Parse(os.Args[2:])
		handleConfig(configCmd)
	case "version":
		fmt.Printf("LoLArchiver CLI v%s\n", version)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("LoLArchiver CLI - A command-line interface for LoLArchiver API")
	fmt.Println("\nUsage:")
	fmt.Println("  lolarchiver-cli [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  credits     Check remaining API credits")
	fmt.Println("  youtube     YouTube-related operations")
	fmt.Println("  twitter     Twitter-related operations")
	fmt.Println("  twitch      Twitch-related operations")
	fmt.Println("  kick        Kick-related operations")
	fmt.Println("  reverse     Reverse lookup operations (email/phone)")
	fmt.Println("  database    Database search operations")
	fmt.Println("  config      Configuration operations")
	fmt.Println("  version     Show version information")
	fmt.Println("  help        Show this help message")
	fmt.Println("\nUse 'lolarchiver-cli [command] --help' for more information about a command")
}

func getClient() (*api.Client, error) {
	apiKey, err := config.GetAPIKey()
	if err != nil {
		return nil, err
	}
	return api.NewClient(apiKey), nil
}

func handleCredits(cmd *flag.FlagSet) {
	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.CheckCredits()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleYouTube(cmd *flag.FlagSet) {
	commentsCmd := flag.NewFlagSet("comments", flag.ExitOnError)
	repliesCmd := flag.NewFlagSet("replies", flag.ExitOnError)

	if len(os.Args) < 3 {
		fmt.Println("Expected 'comments' or 'replies' subcommand")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "comments":
		commentsCmd.Parse(os.Args[3:])
		handleYouTubeComments(commentsCmd)
	case "replies":
		repliesCmd.Parse(os.Args[3:])
		handleYouTubeReplies(repliesCmd)
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func handleYouTubeComments(cmd *flag.FlagSet) {
	userID := cmd.String("user-id", "", "YouTube user ID")
	handle := cmd.String("handle", "", "YouTube handle")
	channelID := cmd.String("channel-id", "", "YouTube channel ID")
	offset := cmd.Int("offset", 0, "Pagination offset")

	if *userID == "" && *handle == "" && *channelID == "" {
		fmt.Println("Error: At least one of user-id, handle, or channel-id must be provided")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.YouTubeUserComments(*userID, *handle, *channelID, *offset)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleYouTubeReplies(cmd *flag.FlagSet) {
	commentID := cmd.String("comment-id", "", "YouTube comment ID")
	if *commentID == "" {
		fmt.Println("Error: comment-id is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.YouTubeCommentReplies(*commentID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleTwitch(cmd *flag.FlagSet) {
	messagesCmd := flag.NewFlagSet("messages", flag.ExitOnError)
	timeoutsCmd := flag.NewFlagSet("timeouts", flag.ExitOnError)
	historyCmd := flag.NewFlagSet("history", flag.ExitOnError)
	followageCmd := flag.NewFlagSet("followage", flag.ExitOnError)
	followersCmd := flag.NewFlagSet("followers", flag.ExitOnError)

	if len(os.Args) < 3 {
		fmt.Println("Expected 'messages', 'timeouts', 'history', 'followage', or 'followers' subcommand")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "messages":
		handleTwitchMessages(messagesCmd)
	case "timeouts":
		handleTwitchTimeouts(timeoutsCmd)
	case "history":
		handleTwitchHistory(historyCmd)
	case "followage":
		handleTwitchFollowage(followageCmd)
	case "followers":
		handleTwitchFollowers(followersCmd)
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func handleTwitchMessages(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Twitch username")
	server := cmd.String("server", "superserver2", "Server (superserver2 or main)")
	offset := cmd.Int("offset", 0, "Pagination offset")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.TwitchUserMessages(*username, *server, *offset)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleTwitchTimeouts(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Twitch username")
	offset := cmd.Int("offset", 0, "Pagination offset")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.TwitchUserTimeouts(*username, *offset)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleTwitchHistory(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Twitch username")
	mode := cmd.String("mode", "", "Mode (username, utype, or btype)")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.TwitchUserHistory(*username, *mode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleTwitchFollowage(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Twitch username")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.TwitchFollowage(*username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleTwitchFollowers(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Twitch username")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.TwitchFollowers(*username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleKick(cmd *flag.FlagSet) {
	messagesCmd := flag.NewFlagSet("messages", flag.ExitOnError)
	timeoutsCmd := flag.NewFlagSet("timeouts", flag.ExitOnError)
	modsCmd := flag.NewFlagSet("mods", flag.ExitOnError)
	subscribersCmd := flag.NewFlagSet("subscribers", flag.ExitOnError)

	if len(os.Args) < 3 {
		fmt.Println("Expected 'messages', 'timeouts', 'mods', or 'subscribers' subcommand")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "messages":
		handleKickMessages(messagesCmd)
	case "timeouts":
		handleKickTimeouts(timeoutsCmd)
	case "mods":
		handleKickMods(modsCmd)
	case "subscribers":
		handleKickSubscribers(subscribersCmd)
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func handleKickMessages(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Kick username")
	offset := cmd.Int("offset", 0, "Pagination offset")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.KickUserMessages(*username, *offset)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleKickTimeouts(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Kick username")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.KickUserTimeouts(*username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleKickMods(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Kick username")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.KickUserModChannels(*username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleKickSubscribers(cmd *flag.FlagSet) {
	username := cmd.String("username", "", "Kick username")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Println("Error: username is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.KickUserSubscribers(*username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleReverse(cmd *flag.FlagSet) {
	phoneCmd := flag.NewFlagSet("phone", flag.ExitOnError)
	emailCmd := flag.NewFlagSet("email", flag.ExitOnError)

	if len(os.Args) < 3 {
		fmt.Println("Expected 'phone' or 'email' subcommand")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "phone":
		phoneCmd.Parse(os.Args[3:])
		handleReversePhone(phoneCmd)
	case "email":
		emailCmd.Parse(os.Args[3:])
		handleReverseEmail(emailCmd)
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func handleReversePhone(cmd *flag.FlagSet) {
	phone := cmd.String("phone", "", "Phone number")
	insecureMode := cmd.Bool("insecure", false, "Use insecure mode")

	if err := cmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *phone == "" && len(cmd.Args()) > 0 {
		*phone = cmd.Args()[0]
	}

	if *phone == "" {
		fmt.Println("Error: phone is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	done := make(chan bool)
	go func() {
		startTime := time.Now()
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(startTime).Round(time.Second)
				fmt.Fprintf(os.Stderr, "\rProcessing (%v)...", elapsed)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Check if API key exists in config
	config, err := config.Load()
	if err != nil {
		done <- true
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		done <- true
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.ReversePhoneLookup(*phone, *insecureMode)
	if err != nil {
		done <- true
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	done <- true
	fmt.Fprint(os.Stderr, "\r")

	switch resp.StatusCode {
	case 200:
		if len(resp.Body) == 0 || string(resp.Body) == "[]" {
			fmt.Println("No data found for this phone number")
		} else {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, resp.Body, "", "  "); err == nil {
				fmt.Println(prettyJSON.String())
			} else {
				fmt.Println(string(resp.Body))
			}
		}
	case 401:
		if config.APIKey == "" {
			fmt.Println("Error: Unauthorized - Please set your API key using:")
			fmt.Println("  ./lolarchiver-cli config set-api-key YOUR_API_KEY")
		} else {
			fmt.Println("Error: Phone lookup is only available through the web interface or you exceeded rate limit for today/this month.")
			fmt.Println("Please visit https://lolarchiver.com to use this feature.")
		}
	case 402:
		fmt.Println("Error: Phone lookup is only available through the web interface or you exceeded rate limit for today/this month.")
		fmt.Println("Please visit https://lolarchiver.com to use this feature.")
	case 403:
		fmt.Println("Error: Your current plan does not support phone lookup or you exceeded rate limit for today/this month.")
		fmt.Println("Please upgrade your plan or use the web interface at https://lolarchiver.com")
	case 404:
		fmt.Println("Error: No results found for this phone number")
	case 405:
		fmt.Println("Error: Phone number is too long")
	case 406:
		fmt.Println("Error: Phone number format is incorrect")
	case 415:
		fmt.Println("Error: Owner requested these results to be hidden")
	case 416:
		fmt.Println("Error: You have exhausted all credits. Credits refresh in 24 hours")
	case 500:
		fmt.Println("Error: Internal server error")
	default:
		fmt.Printf("Error: Unexpected response (Status %d)\n", resp.StatusCode)
		if len(resp.Body) > 0 {
			fmt.Println(string(resp.Body))
		}
	}
}

func handleReverseEmail(cmd *flag.FlagSet) {
	email := cmd.String("email", "", "Email address")
	insecureMode := cmd.Bool("insecure", false, "Use insecure mode")

	if *email == "" {
		fmt.Println("Error: email is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.ReverseEmailLookup(*email, *insecureMode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resp.Body))
}

func handleDatabase(cmd *flag.FlagSet) {
	query := cmd.String("query", "", "Search query")
	exact := cmd.Bool("exact", false, "Exact match")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *query == "" && len(cmd.Args()) > 0 {
		*query = cmd.Args()[0]
	}

	if *query == "" {
		fmt.Println("Error: query is required")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	done := make(chan bool)
	go func() {
		startTime := time.Now()
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(startTime).Round(time.Second)
				fmt.Fprintf(os.Stderr, "\rProcessing (%v)...", elapsed)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	config, err := config.Load()
	if err != nil {
		done <- true
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	client, err := getClient()
	if err != nil {
		done <- true
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.DatabaseLookup(*query, *exact)
	if err != nil {
		done <- true
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	done <- true
	fmt.Fprint(os.Stderr, "\r")

	switch resp.StatusCode {
	case 200:
		if len(resp.Body) == 0 || string(resp.Body) == "[]" {
			fmt.Println("No data found for this query")
		} else {
			// Try to pretty print the JSON if possible
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, resp.Body, "", "  "); err == nil {
				fmt.Println(prettyJSON.String())
			} else {
				fmt.Println(string(resp.Body))
			}
		}
	case 401:
		if config.APIKey == "" {
			fmt.Println("Error: Unauthorized - Please set your API key using:")
			fmt.Println("  ./lolarchiver-cli config set-api-key YOUR_API_KEY")
		} else {
			fmt.Println("Error: Database lookup is only available through the web interface or you exceeded rate limit for today/this month.")
			fmt.Println("Please visit https://lolarchiver.com to use this feature.")
		}
	case 402:
		fmt.Println("Error: Database lookup is only available through the web interface or you exceeded rate limit for today/this month.")
		fmt.Println("Please visit https://lolarchiver.com to use this feature.")
	case 403:
		fmt.Println("Error: Your current plan does not support database lookup or you exceeded rate limit for today/this month.")
		fmt.Println("Please upgrade your plan or use the web interface at https://lolarchiver.com")
	case 404:
		fmt.Println("Error: No results found for this query")
	case 416:
		fmt.Println("Error: You have exhausted all credits. Credits refresh in 24 hours")
	case 500:
		fmt.Println("Error: Internal server error")
	default:
		fmt.Printf("Error: Unexpected response (Status %d)\n", resp.StatusCode)
		if len(resp.Body) > 0 {
			fmt.Println(string(resp.Body))
		}
	}
}

func handleConfig(cmd *flag.FlagSet) {
	setAPIKeyCmd := flag.NewFlagSet("set-api-key", flag.ExitOnError)

	if len(os.Args) < 3 {
		fmt.Println("Expected 'set-api-key' subcommand")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "set-api-key":
		setAPIKeyCmd.Parse(os.Args[3:])
		apiKey := setAPIKeyCmd.Arg(0)
		if apiKey == "" {
			fmt.Println("Error: API key is required")
			setAPIKeyCmd.PrintDefaults()
			os.Exit(1)
		}

		if err := config.SetAPIKey(apiKey); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("API key set successfully")
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
} 