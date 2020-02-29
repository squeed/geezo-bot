# Geezo-bot
A simple bot to scrape images from Tinybeans and send them as bare attachments.

It is intended to be run as a cronjob. It reads from, and writes to, a imap / smtp account. It stores some state in a sqlite DB.


## Usage

`geezo-bot --config=ZZZ`

## Config
A simple yaml file:

```yaml
imap:
  server: imap.example.com:123
  username: u
  password: "p"
smtp:
  server: smtp.example.com:456
  username: u
  password: "p"

main:
  to: frame@foo.com
  maxAttachments: 5
  dbFile: /path/to/sqlite.db
  doneFolder: geezo-bot-done

```


## Architecture
A periodic imap scraper that looks for messages in the inbox. If it finds them, it pulls images out, adds them to a database, and moves the message.

Then, all pending images are downloaded.
Then, all downloaded images are emailed.

