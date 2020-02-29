package app

import (
	"os"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/squeed/geezo-bot/pkg/config"
	"github.com/squeed/geezo-bot/pkg/db"
	"github.com/squeed/geezo-bot/pkg/mail"
)

type App struct {
	config config.Config
	db     *db.Conn
}

func Init(c *config.Config) (*App, error) {
	a := App{
		config: *c,
	}

	var err error
	a.db, err = db.GetConn(c)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to DB")
	}

	return &a, nil
}

// get list of images, add to db, move mails
func (a *App) ReadMail() error {
	conn, err := mail.IMAPConnect(&a.config)
	if err != nil {
		return err
	}

	klog.Info("retrieving messages")

	getRes, err := conn.GetMessages()
	if err != nil {
		return err
	}

	klog.Infof("processing %d messages", len(getRes.Messages))
	for uid := range getRes.Messages {
		a.processMessage(getRes, uid)
	}

	klog.Info("moving messages")
	err = conn.MoveMessages(&a.config, getRes)
	if err != nil {
		// not fatal, continue
		klog.Error("failed to move messages ", err)
	}

	return nil
}

func (a *App) DownloadImages() error {
	os.MkdirAll(a.config.Main.WorkDir, 0755)

	images, err := a.db.GetUndownloadedImages()
	if err != nil {
		return err
	}

	for _, image := range images {
		_ = a.downloadImage(&image)

	}
	return nil
}

func (a *App) SendMessages() error {
	images, err := a.db.GetUnsentImages()
	if err != nil {
		return err
	}

	for _, image := range images {
		err = a.sendImage(&image)
		if err != nil {
			klog.Warning(err)
		}
	}
	return nil

}

func (a *App) Close() {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			klog.Error("error closing DB", err)
		}
	}
}
