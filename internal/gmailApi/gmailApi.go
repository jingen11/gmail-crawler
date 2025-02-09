package gmailapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func GetConfig(credentials []byte) (*oauth2.Config, error) {
	return google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope)
}

func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Request a token from the web, then returns the retrieved token.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Set up local server for OAuth callback
	config.RedirectURL = "http://localhost:8080/callback"

	ch := make(chan string)
	randState := "state-token"

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != randState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "Authorization successful! You can close this window.")
		ch <- code
	})

	// Start local server
	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Generate authorization URL and open it in browser
	authURL := config.AuthCodeURL(randState, oauth2.AccessTypeOffline)
	fmt.Printf("Opening browser for authorization: %v\n", authURL)

	err := exec.Command("open", authURL).Start()
	if err != nil {
		fmt.Printf("Failed to open browser automatically. Please open this URL manually:\n%v\n", authURL)
	}

	// Wait for callback
	code := <-ch

	// Shutdown server
	server.Shutdown(context.Background())

	// Exchange code for token
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return tok
}

func InitService(config *oauth2.Config, ctx context.Context, tok *oauth2.Token) (*gmail.Service, error) {
	client := config.Client(ctx, tok)

	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

func ListMessages(srv *gmail.Service, filters, user string) ([]*gmail.Message, error) {
	res, err := srv.Users.Messages.List(user).Q(filters).Do()
	if err != nil {
		return []*gmail.Message{}, err
	}
	if len(res.Messages) == 0 {
		fmt.Println("No messages found.")
		return []*gmail.Message{}, nil
	}
	return res.Messages, nil
}

func GetMessage(srv *gmail.Service, user, messageId string) (*gmail.Message, error) {
	return srv.Users.Messages.Get(user, messageId).Format("full").Do()
}

func GetAttachment(srv *gmail.Service, user, messageId, attachmentId string) (*gmail.MessagePartBody, error) {
	return srv.Users.Messages.Attachments.Get(user, messageId, attachmentId).Do()
}
