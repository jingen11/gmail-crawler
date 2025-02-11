package command

import (
	"context"
	"log"
	"os"
	"strings"

	gmailapi "github.com/jingen11/gmail-crawler/internal/gmailApi"
	"github.com/jingen11/gmail-crawler/internal/utils"
)

type Command struct {
	Name      string
	Arguments []string
}

func HandleAuth(cmd *Command) error {
	credentials, err := os.ReadFile("./credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return err
	}

	config, err := gmailapi.GetConfig(credentials)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return err
	}

	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "./token.json"
	_, err = gmailapi.TokenFromFile(tokFile)
	if err != nil {
		tok := gmailapi.GetTokenFromWeb(config)
		utils.SaveToken(tokFile, tok)
	}
	return nil
}

func HandleScrap(cmd *Command) error {
	length := len(cmd.Arguments)
	credentials, err := os.ReadFile("./credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return err
	}

	config, err := gmailapi.GetConfig(credentials)
	if err != nil {
		log.Fatalf("Unable to get config: %v", err)
		return err
	}
	tokFile := "./token.json"
	tok, err := gmailapi.TokenFromFile(tokFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return err
	}
	srv, err := gmailapi.InitService(config, context.Background(), tok)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	filters := "has:attachment from:no-reply-aws@amazon.com OR invoice+statements@mail.anthropic.com OR invoice+statements@boot.dev OR mongodb-account@mongodb.com"

	if length == 1 { // date
		filters = "has:attachment from:no-reply-aws@amazon.com OR invoice+statements@mail.anthropic.com OR invoice+statements@boot.dev OR mongodb-account@mongodb.com after:" + cmd.Arguments[0]
	}

	if length == 2 { // date // from
		emails := strings.Split(cmd.Arguments[1], ",")
		from := ""
		for index, email := range emails {
			if index == 0 {
				from += email
			} else {
				from += " OR " + email
			}
		}
		filters = "has:attachment after:" + cmd.Arguments[0] + " from:" + from
	}
	user := "me"
	messages, err := gmailapi.ListMessages(srv, filters, user)
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
		return err
	}
	for _, l := range messages {
		msg, err := gmailapi.GetMessage(srv, user, l.Id)
		if err != nil {
			log.Printf("Unable to retrieve message %v: %v", l.Id, err)
			continue
		}

		utils.ProcessMessage(msg, srv, user, l.Id)
	}
	return err
}
