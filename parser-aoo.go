package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	rss "github.com/gorilla/feeds"
)

// Archive of Our Own notification email parser.
func parseAoo(notifications []*NotificationEmail) []*rss.Item {
	items := []*rss.Item{}

	for _, notification := range notifications {

		// Other emails from AO3 will not contain this string in the subject.
		if strings.Contains(notification.Subj, "posted Chapter") {
			log.Printf("Subj: %s", notification.Subj)

			item := rss.Item{
				IsPermaLink: "true",
				Author:      &rss.Author{},
				Link: &rss.Link{
					Rel: "alternate",
				},
			}

			// Parse the HTML part of the message.
			soup, err := goquery.NewDocumentFromReader(bytes.NewBufferString(notification.Html))
			if err != nil {
				log.Printf("Message failed to parse, skipping...")
			}

			type ThingLink struct {
				Text string
				Url  string
			}

			var author ThingLink
			var work ThingLink
			var chapter ThingLink

			// Produce the information we're going to use in the feed item
			// -- here, by investigating every <a> tag
			// and picking out the ones we know the meaning of.
			soup.Find("a").Each(func(i int, s *goquery.Selection) {
				link, exists := s.Attr("href")
				if exists {
					// Abject idiocy: Apparently, sometimes they have "http" instead of "https" links.
					// Why?!
					link = strings.Replace(link, "http://", "https://", -1)

					switch {
					case strings.HasPrefix(link, "https://archiveofourown.org/users"):
						author.Text = s.Text()
						author.Url = link
					case strings.HasPrefix(link, "https://archiveofourown.org/works"):
						if strings.Contains(link, "chapters") {
							chapter.Url = link
							chapter.Text = s.Text()
						} else {
							work.Url = link
							work.Text = s.Text()
						}
					}
				}
			})

			if (author == ThingLink{} || chapter == ThingLink{} || work == ThingLink{}) {
				log.Print("failed to parse critical data out of the message, skipping...")
				continue
			}

			// We have the info we needed, so build the feed item
			item.Author.Name = author.Text
			item.Id = chapter.Url
			item.Link.Href = chapter.Url
			item.Title = fmt.Sprintf("%s - %s", work.Text, chapter.Text)
			item.Updated = notification.Date
			item.Created = notification.Date
			item.Description = fmt.Sprintf(`
			<p><a href="%s">%s</a> posted an update for story
			<a href="%s">%s</a>:
			<a href="%s">%s</a></p>
			`, author.Url, author.Text, work.Url, work.Text, chapter.Url, chapter.Text)

			items = append(items, &item)
		}
	}

	return items
}
