package main

// The most arcane part of the whole thing, because apparently,
// to parse email with golang you need to actually know the IMAP standard
// by heart, because none of this is properly documented.
//
// Oh well.

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"golang.org/x/net/html/charset"
)

// Because the actual types are truly mindbending, we consolidate the interesting
// parts of the email into a simpler structure before handing that over to the parsers.
type NotificationEmail struct {
	From string
	Subj string
	Date time.Time
	Text string
	Html string
}

// The annoying part: taking the message apart into its html and txt bodies.
// It is my intuition, (the documentation is exceedingly lacking)
// that the BodySection[] map always contains exactly one element
// when the message was downloaded with Collect() as above.
// Whose key is a struct.
// And bizarrely, it's not a nil struct.
// So we have to curse to high heavens and loop through sections.
func parseEmail(message *imapclient.FetchMessageBuffer) *NotificationEmail {

	notification := NotificationEmail{
		From: message.Envelope.From[0].Addr(),
		Subj: message.Envelope.Subject,
		Date: message.Envelope.Date,
	}

	for _, bodyPart := range message.BodySection {

		mr, err := mail.CreateReader(bytes.NewReader(bodyPart))
		if err != nil {
			// Looking at the createreader code, if there was
			// an error, it's probably a borked message anyway.
			return nil
		}

	partLoop:
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("failed to read message part: %v", err)
				return nil
			}

			// We ignore attachments and stuff...
			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				chunkBytes, _ := io.ReadAll(p.Body)
				contentType, contentTypeParams, err := h.ContentType()
				if err != nil {
					continue partLoop
				}

				// In case we got utf-8, that's where it ends.
				text := string(chunkBytes)

				// But encodings that are not utf-8 should be converted to utf-8.
				// We're assuming they didn't lie to us. (they can)
				cs := contentTypeParams["charset"]
				if strings.ToUpper(cs) != "UTF-8" {
					enc, _, certain := charset.DetermineEncoding(chunkBytes, h.Get("Content-Type"))
					if certain && enc != nil {
						decodedText, err := enc.NewDecoder().Bytes(chunkBytes)
						if err != nil {
							log.Printf("failed to decode message, skipping.")
							continue partLoop
						}
						text = string(decodedText)
					}
				}

				switch contentType {
				case "text/plain":
					notification.Text = text
				case "text/html":
					notification.Html = text
				}

			}
		}
	}
	return &notification
}

// Reach into the IMAP server and ask it for all unread emails matching a certain From address.
func fetchMail(client *imapclient.Client, fromAddress string) []*NotificationEmail {

	searchResult, err := client.Search(
		&imap.SearchCriteria{
			Header: []imap.SearchCriteriaHeaderField{
				{Key: "From", Value: fromAddress},
			},
			NotFlag: []imap.Flag{
				"\\Seen",
			}}, &imap.SearchOptions{}).Wait()
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	fetchOptions := &imap.FetchOptions{
		Flags:         true,
		Envelope:      true,
		BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
		BodySection:   []*imap.FetchItemBodySection{{}},
	}

	messages, err := client.Fetch(searchResult.All, fetchOptions).Collect()
	if err != nil {
		log.Fatalf("failed to fetch: %v", err)
	}

	log.Printf("unread messages from %s: %d", fromAddress, len(messages))

	parsedMessages := make([]*NotificationEmail, 0, len(messages))

	for _, message := range messages {
		result := parseEmail(message)
		if result != nil {
			parsedMessages = append(parsedMessages, result)
		}
	}
	return parsedMessages
}

// Log into the IMAP server, fetch all interesting emails,
// and produce a slice of structures containing the important parts we're looking for.
func acquireEmail(cfg NotifyRSSConfig) []*NotificationEmail {
	var err error

	dialTone := fmt.Sprintf("%s:%d", cfg.Mail.Host, cfg.Mail.Port)
	var client *imapclient.Client
	switch strings.ToLower(cfg.Mail.Connection) {
	case "ssl":
		client, err = imapclient.DialTLS(dialTone, nil)
	case "starttls":
		client, err = imapclient.DialStartTLS(dialTone, nil)
	case "plain":
		client, err = imapclient.DialInsecure(dialTone, nil)
	default:
		log.Fatalf("ssl parameter must be one of 'plain', 'ssl', 'starttls'")
	}
	if err != nil {
		log.Fatalf("connection failure: %v", err)
	}
	defer client.Close()

	if err := client.Login(cfg.Mail.User, cfg.Mail.Pass).Wait(); err != nil {
		log.Fatalf("failed to login: %v", err)
	}

	mailbox, err := client.Select(cfg.Mail.Folder, &imap.SelectOptions{ReadOnly: true}).Wait()
	if err != nil {
		log.Fatalf("failed to reach %s: %v", cfg.Mail.Folder, err)
	}

	log.Printf("%s contains %v messages", cfg.Mail.Folder, mailbox.NumMessages)

	var notifications []*NotificationEmail

	if mailbox.NumMessages > 0 {
		for _, notifier := range SupportedNotifiers {
			notifications = append(notifications, fetchMail(client, notifier.From)...)
		}
	}

	if err := client.Logout().Wait(); err != nil {
		log.Fatalf("failed to logout: %v", err)
	}

	return notifications
}
