package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/squeed/geezo-bot/pkg/db"
)

func (a *App) downloadImage(im *db.Image) error {
	path, err := a.getImage(im)
	if err != nil {
		klog.Error("failed to get image ", err)
		im.DownloadFailCount += 1
		if err2 := a.db.UpdateImage(im); err2 != nil {
			klog.Error("failed to update image after get failure ", err2)
		}
	}

	im.DiskPath = path
	if err := a.db.UpdateImage(im); err != nil {
		return errors.Wrapf(err, "failed to update image record %d", im.ID)
	}

	return nil
}

func imageFilename(im *db.Image) string {
	u, err := url.Parse(im.Url)
	if err != nil {
		klog.Errorf("could not parse url %v: %s", im.Url, err)
		return ""
	}

	ext := filepath.Ext(u.Path)

	return fmt.Sprintf("%d%s", im.ID, ext)
}

func (a *App) getImage(im *db.Image) (string, error) {
	if im.Url == "" {
		panic("coding error: no URL")
	}

	filename := imageFilename(im)
	if filename == "" {
		return "", errors.New("could not process image filename")
	}

	destPath := filepath.Join(a.config.Main.WorkDir, filename)
	tempPath := destPath + ".temp"

	klog.Infof("Downloading image %d (%s) to %s", im.ID, im.Url, destPath)

	if _, err := os.Stat(destPath); err == nil {
		klog.Info("image %s already exists, skipping", destPath)
		return destPath, nil
	}

	destFile, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	resp, err := http.Get(im.Url)
	if err != nil {
		destFile.Close()
		return "", errors.Wrapf(err, "failed to retrieve url for image %d (%s)", im.ID, im.Url)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(destFile, resp.Body); err != nil {
		destFile.Close()
		return "", errors.Wrapf(err, "failed to write image to file %s", tempPath)
	}

	if err := destFile.Close(); err != nil {
		return "", errors.Wrapf(err, "failed to close file %s", tempPath)
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		return "", errors.Wrap(err, "failed to rename downloaded file")
	}

	return filename, nil
}
