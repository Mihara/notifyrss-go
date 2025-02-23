package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"time"

	rss "github.com/gorilla/feeds"
)

type NotificationSource struct {
	Name   string                                 // Used in the log.
	From   string                                 // Email the notifications come from.
	Cfg    string                                 // Name of the config variable in Files section
	Parser func([]*NotificationEmail) []*rss.Item // Parser function.
	Title  string                                 // Feed title.
	Link   string                                 // Feed link.
}

// All the notifiers we support go in here:
var SupportedNotifiers = []NotificationSource{
	{
		Name:   "Archive of Our Own",
		From:   "do-not-reply@archiveofourown.org",
		Cfg:    "Aoo",
		Parser: parseAoo,
		Title:  "Archive of Our Own Story Updates",
		Link:   "https://archiveofourown.org/",
	},
}

// This function assembles the items a parser function produced into a feed.
func makeFeed(notifier NotificationSource, items []*rss.Item, format string) (string, error) {
	now := time.Now()

	slices.SortFunc(items, func(a *rss.Item, b *rss.Item) int {
		return -a.Created.Compare(b.Created)
	})

	feed := &rss.Feed{
		Title:   notifier.Title,
		Id:      notifier.Link,
		Link:    &rss.Link{Href: notifier.Link, Rel: "self"},
		Updated: now,
		Created: now,
		Items:   items,
	}

	log.Printf("resulting feed of %s updates contains %d items", notifier.Name, len(items))

	switch format {
	case "atom":
		return feed.ToAtom()
	case "rss":
		return feed.ToRss()
	case "json":
		return feed.ToJSON()
	}
	return "", fmt.Errorf("unsupported feed format '%s'", format)

}

// This goes through the list of notifiers we know, runs the parser functions for them,
// and feeds their output into the feed generator function above.
func generateFeeds(cfg NotifyRSSConfig, notifications []*NotificationEmail) {

	for _, notifier := range SupportedNotifiers {

		// The annoying part, we need to get the value of the config field using reflection,
		// because I can't just add a pointer to a config field.
		target := reflect.ValueOf(cfg.Files).FieldByName(notifier.Cfg).String()

		// If the filename is configured, we run the appropriate parser.
		if target != "" {

			items := notifier.Parser(
				slices.Collect(Filtered(notifications, func(n *NotificationEmail) bool {
					return n.From == notifier.From
				})))

			feedText, err := makeFeed(notifier, items, cfg.Options.Format)
			if err != nil {
				log.Printf("failed to produce a feed, this might be a bug: %v", err)
			}

			log.Printf("saving feed in %s format to %s", cfg.Options.Format, target)
			err = os.WriteFile(target, []byte(feedText), 0644)
			if err != nil {
				log.Fatalf("could not write target feed file %s: %v", target, err)
			}

		}
	}
}
