# notifyrss-go

## What is it?

Imagine you have a [Miniflux](https://miniflux.app/) installation and you're tracking stories on [Archive of Our Own](https://archiveofourown.org/). And while that latter sends you email notifications about stories updating, you want them in your Miniflux instead. Or tinyRSS. Or Feedly. Or any other RSS/Atom/JSONFeed feed reader you like.

This tool was born to solve this particular problem, and so far, no others.

In the future, I plan to extend it to do other, similar jobs of transmuting email notifications into a feed of story updates (Major forums like Spacebattles and Sufficient Velocity come to mind), and there is code in there to make it easier, but right now, that's all it is doing.

## Installation

This is a [Go](https://go.dev/) program, so to build from source you just:

```shell
go install github.com/Mihara/notifyrss-go@latest
```

Or you can grab one of the binaries on the releases page. This is pure Go, and should work on any platform Go can compile for. Once you have an executable, it's on you to run it at regular intervals, or whenever a new email message comes in, in whatever way seems more expedient.

```shell
notifyrss [configuration file]
```

If you don't supply the configuration file parameter, it looks for `config.yaml` in the current directory.

You will also need to get the resulting static RSS/Atom/JSONFeed file to a web server, so that your Miniflux/tinyRSS/Feedly can pick it up. If you run your own feed aggregator, you probably already have a web server, or don't really have a problem with setting one up. If you don't, it shouldn't be difficult to set up Github Pages or Neocities or any other free static hoster to serve it, as long as you can run `notifyrss-go` at regular intervals to update your feed file.

Ideally, you want a separate email account to collect notifications and set up a forwarding scheme from your primary account that you used to register with *Archive of Our Own* where you actually receive notification emails. *(That's what I did.)* This is because your configuration file will inevitably contain the password to access this account, in plain text. Having a separate write-only account for the job is inherently more secure. While I have this account on my own email server, which runs on the same machine as my Miniflux installation, it can be anywhere, the only real requirement is to offer IMAP access and accept password login.

## Configuration

The configuration is a [YAML](https://en.wikipedia.org/wiki/YAML) file:

```yaml
mail:
  host: example.com
  port: 993
  connection: ssl
  user: notifier
  pass: verysecret
  folder: INBOX
options:
  format: atom
files:
  aoo: ao3-notifications.atom.xml
```

+ **mail**: Section pertaining to setting up the email where it will be picking up notifications from.
  + **host**: hostname of the email server. Required.
  + **port**: port of the IMAP server. Default is `993`.
  + **connection**: Connection type. Valid types are `plain`, `ssl`, `starttls`, default is `ssl`.
  + **user**: Username used for logging in. Required.
  + **pass**: Password. Required.
  + **folder**: IMAP folder to check. Default is `INBOX`, which is the primary inbox. You can use some other folder, e.g. set up your email to sort all notifications about story updates into a separate folder, and use that, although I still recommend a separate account. Ideally, there should be no extraneous emails in this folder, although if there are any, they will be ignored.
+ **options**: Section for general options.
  + **format**: Format of the feed to generate. Valid formats are `atom`, `rss`, `json`. Default is `atom`.
+ **files**: Files to be generated.
  + **aoo**: The filename for the Archive of Our Own feed. If not given, the feed will not be generated at all.

## How it works

Given the configuration file, `notifyrss-go` logs into the IMAP account, acquires every *unread* email in the given mailbox which it recognizes as coming from Archive of Our Own email notifier (or potentially, other such notifiers, once I get around to making them) and parses it to make a plausible feed item telling you that a story has a new chapter to read. The feed is then saved to a file. That's it. With Miniflux in particular, you can even configure it to fetch the actual chapter text, which is quite convenient.

It's important to note two things:

+ The emails will *stay* unread. It's on you to decide when you want to mark them read if at all.
+ Only the emails currently present in the mailbox and still unread will appear in the generated feed file as feed entries.

In practice, it will take you years of active reading to rack up enough notifications for the feed generation to start taking more than a second.

## Development

If you wish to preempt me and write a parser for some other kind of email notification, I'm open to pull requests -- take a look at `feed.go` and `parser-aoo.go` where comments should make what you need to do reasonably obvious. There's no reason this tool shouldn't be able to handle any reasonable email notifier service under the sun.

To build release binaries, you may want to use [Task](https://taskfile.dev/), although there's nothing particularly special about what it is doing here, and simple `go build` will build:

```shell
task build
```

## License

This program is released under the terms of MIT license. See the full text in [LICENSE](LICENSE)
