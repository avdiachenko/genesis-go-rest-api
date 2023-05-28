package main


import (
    "io"
    "os"
    "fmt"
    "log"
    "regexp"
    "errors"
    "context"
    "strings"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "encoding/base64"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/option"
    "google.golang.org/api/gmail/v1"
)


type Email struct {
    name string
}


func getBitcoinPrice() string {
    resp, err := http.Get("https://api.coincap.io/v2/assets/bitcoin")
    if err != nil {
        fmt.Println(err)
        return ""
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return ""
    }
    var apiData map[string]any
    json.Unmarshal(body, &apiData)

    if responseBody, ok := apiData["data"].(map[string]any); ok {
        if responseString, ok := responseBody["priceUsd"].(string); ok {
            return responseString
        } else {
            return ""
        }
    } else {
        return ""
    }
}


func bitcoinPriceHandler(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "application/json")

    bitcoinPrice := getBitcoinPrice()

    if bitcoinPrice != "" {
        fmt.Fprintf(w, bitcoinPrice)
    } else {
        w.WriteHeader(400)
    }
}


func subscriptionHandler(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "application/json")

    emailFile := "dat1"

    r.ParseForm()
    newEmail := r.Form["email"][0]

    emailExists := false
    _, err := os.Stat(emailFile)
    if !errors.Is(err, os.ErrNotExist) {
        data, err := ioutil.ReadFile(emailFile)
        if err != nil {
            fmt.Println(err)
        }

        r, _ := regexp.Compile("(^|\n)" + newEmail + "\n")
        emailExists = r.MatchString(string(data))
    }

    if emailExists {
        w.WriteHeader(409)
        return
    }

    f, err := os.OpenFile(emailFile, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0644)
    if err != nil {
        fmt.Println(err)
    }
    defer f.Close()

    _, err = f.WriteString(r.Form["email"][0] + "\n")
    if err != nil {
        fmt.Println(err)
    }
    f.Sync()
}


func emailingPriceHandler(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "application/json")

    emailFile := "dat1"

    var emails []string

    _, err := os.Stat(emailFile)
    if !errors.Is(err, os.ErrNotExist) {
        data, err := ioutil.ReadFile(emailFile)
        if err != nil {
            fmt.Println(err)
        }

        emails = strings.Split(string(data), "\n")
    }

    subject := "Bitcoin (BTC) price update!"

    body := getBitcoinPrice()

    if len(emails) > 0 {
        sendMessage(emails, subject, body)
    }
}


func handleRequests() {
    http.HandleFunc("/rate", bitcoinPriceHandler)
    http.HandleFunc("/subscribe", subscriptionHandler)
    http.HandleFunc("/sendEmails", emailingPriceHandler)
    log.Fatal(http.ListenAndServe(":3000", nil))
}


func sendMessage(recipients []string, subject string, body string) {
    ctx := context.Background()
    b, err := os.ReadFile("credentials.json")
    if err != nil {
            log.Fatalf("Unable to read client secret file: %v", err)
    }

    // If modifying these scopes, delete your previously saved token.json.
    config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
    if err != nil {
            log.Fatalf("Unable to parse client secret file to config: %v", err)
    }
    client := getClient(config)

    srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
            log.Fatalf("Unable to retrieve Gmail client: %v", err)
    }

    // Compose the message
    var message gmail.Message

	messageStr := []byte(
		"From: youremail@gmail.com\r\n" +
			"To: " + strings.Join(recipients, ", ") + "\r\n" +
			"Subject: " + subject + "\r\n\r\n" +
			body)

	// Place messageStr into message.Raw in base64 encoded format
	message.Raw = base64.URLEncoding.EncodeToString(messageStr)

    _, err = srv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Message sent!")
	}
}


func main() {
    handleRequests()
}
