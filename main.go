// Fun fact: This started life as a much simpler Python tool,
// but then I got tired of setting up a new virtualenv every time
// the python version goes up.
package main

import (
	"log"
	"os"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

// This gets overridden during build of the release version
var version = "(development)"

// Our configuration.
type NotifyRSSConfig struct {
	Mail struct {
		Host       string
		Port       int    `default:"993"`
		Connection string `default:"ssl"`
		User       string
		Pass       string
		Folder     string `default:"INBOX"`
	}
	Options struct {
		Format string `default:"atom"`
	}
	Files struct {
		Aoo string
	}
}

func main() {
	var err error
	var cfg NotifyRSSConfig

	cfgFile := "config.yaml"

	log.Printf("notifyrss %s starting up", version)

	if len(os.Args) == 2 {
		cfgFile = os.Args[1]
	}

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFlags:          true,
		AllowUnknownFields: false,
		Files:              []string{cfgFile},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})

	err = loader.Load()
	if err != nil {
		log.Fatalf("can't read config file: %v", err)
	}

	if cfg.Mail.Host == "" || cfg.Mail.User == "" || cfg.Mail.Pass == "" {
		log.Fatalf("IMAP connection is not configured enough to try it")
	}

	generateFeeds(cfg, acquireEmail(cfg))

	log.Printf("done, see you next time.")

}
