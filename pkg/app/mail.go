package app

import (
	"bytes"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	msgMail "github.com/emersion/go-message/mail"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/squeed/geezo-bot/pkg/db"
	"github.com/squeed/geezo-bot/pkg/mail"
	"github.com/squeed/geezo-bot/pkg/scrape"
)

// ExtractImages reads all desired images from the list of HTML
// contents and adds them to the database.
func (a *App) processMessage(res *mail.GetMessagesResult, uid uint32) error {

	msg := res.Messages[uid]

	urls, err := scrape.ScrapeHTML(msg.Content)
	if err != nil {
		klog.Error("failed to scrape HTML, failing message", err)
		res.FailMessage(uid)
		return nil
	}

	errors := []error{}

	for _, url := range urls {
		im := db.Image{Url: url}
		err := a.db.CreateImage(&im)
		if err != nil {
			klog.Error(err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		res.FailMessage(uid)
	}

	return nil
}

func (a *App) sendImage(im *db.Image) error {
	if im.DiskPath == "" {
		panic("coding error: undownloaded image")
	}

	klog.Infof("preparing message for image ID %d", im.ID)

	fp, err := os.Open(filepath.Join(a.config.Main.WorkDir, im.DiskPath))
	if err != nil {
		if os.IsNotExist(err) {
			klog.Warningf("weird: image %d doesn't seem to exist...", im.ID)

			im.DiskPath = ""
			a.db.UpdateImage(im)
			return nil
		}
		return errors.Wrapf(err, "failed to open image file %s", im.DiskPath)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(im.DiskPath))
	if mimeType == "" {
		klog.Warningf("weird: image %d has an unknown extension", im.ID)
		return nil
	}

	var b bytes.Buffer

	from := []*msgMail.Address{{Address: a.config.Main.From}}
	to := []*msgMail.Address{{Address: a.config.Main.To}}

	// Create our mail header
	var h msgMail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)
	h.SetSubject("This is a photo from geezo-bot")

	// Create a new mail writer
	mw, err := msgMail.CreateWriter(&b, h)
	if err != nil {
		return err
	}

	// Create a text part
	tw, err := mw.CreateInline()
	if err != nil {
		log.Fatal(err)
	}
	var th msgMail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, "This is an automatic message from the geezo-bot!")
	w.Close()
	tw.Close()

	// Create an attachment
	var ah msgMail.AttachmentHeader
	ah.Set("Content-Type", mimeType)
	ah.SetFilename(im.DiskPath)
	w, err = mw.CreateAttachment(ah)
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(w, fp)
	if err != nil {
		return errors.Wrap(err, "failed to create attachment")
	}
	w.Close()
	mw.Close()

	klog.Infof("sending message for image ID %d", im.ID)
	err = mail.SendMessage(&a.config, &b)
	if err == nil {
		im.Sent = true
		err2 := a.db.UpdateImage(im)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update image in DB")
		}
		return nil
	}
	return err
}
