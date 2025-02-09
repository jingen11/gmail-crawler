package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"

	gmailapi "github.com/jingen11/gmail-crawler/internal/gmailApi"
)

const (
	aws       = "no-reply-aws@amazon.com"
	anthropic = "invoice+statements@mail.anthropic.com"
	boot      = "invoice+statements@boot.dev"
)

// Saves a token to a file path.
func SaveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func ProcessMessage(msg *gmail.Message, srv *gmail.Service, user, messageId string) {
	var sender string

	// Print message headers for context
	for _, header := range msg.Payload.Headers {
		if header.Name == "Subject" || header.Name == "From" {
			fmt.Printf("%s: %s\n", header.Name, header.Value)
		}
		if header.Name == "From" {
			sender = header.Value
		}
	}

	DigPart(msg.Payload, srv, user, messageId, sender)
}

func DigPart(payload *gmail.MessagePart, srv *gmail.Service, user, messageId, sender string) {
	fmt.Println(payload.MimeType)
	if len(payload.Parts) > 0 {
		for _, p := range payload.Parts {
			DigPart(p, srv, user, messageId, sender)
		}
	} else {
		if payload.MimeType == "text/plain" || payload.MimeType == "text/html" {
			// body, _ := base64.StdEncoding.DecodeString(payload.Body.Data)
			// fmt.Println(string(body))
		}
		if payload.MimeType == "application/pdf" {
			attachment, err := gmailapi.GetAttachment(srv, user, messageId, payload.Body.AttachmentId)
			if err != nil {
				log.Printf("Error getting attachment: %v", err)
				return
			}

			dec, err := base64.URLEncoding.DecodeString(attachment.Data)
			if err != nil {
				log.Printf("Error decoding attachment data: %v", err)
				return
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Printf("Error getting home directory: %v", err)
				return
			}

			// Create the directory path if it doesn't exist
			savePath := fmt.Sprintf("%s/Desktop/Jeveloper-Tech/Receipt", homeDir)
			if err := os.MkdirAll(savePath, 0755); err != nil {
				log.Printf("Error creating directories: %v", err)
				return
			}

			subDir := "Unclaimed"

			if strings.Contains(sender, anthropic) {
				subDir = "Anthropic"
			} else if strings.Contains(sender, aws) {
				subDir = "AWS"
			} else if strings.Contains(sender, boot) {
				subDir = "Boot.dev"
			}

			filePath := fmt.Sprintf("%s/%s/%s", savePath, subDir, payload.Filename)
			f, err := os.Create(filePath)
			if err != nil {
				log.Printf("Error creating file: %v", err)
				return
			}
			defer f.Close()

			if _, err := f.Write(dec); err != nil {
				log.Printf("Error writing file: %v", err)
				return
			}

			if err := f.Sync(); err != nil {
				log.Printf("Error syncing file: %v", err)
				return
			}

			fmt.Printf("Successfully saved PDF: %s\n", filePath)
		}
	}
}
