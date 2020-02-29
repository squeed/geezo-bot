package db

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Image struct {
	gorm.Model // id, dates

	Url      string `gorm:"unique"`
	DiskPath string `gorm:"type:text"`
	Sent     bool   `gorm:"DEFAULT:false"`

	DownloadFailCount int `gorm:"default:0"`
}

func (conn *Conn) CreateImage(image *Image) error {
	err := conn.db.Create(image).Error
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return nil
		}
		return errors.Wrapf(conn.db.Error, "failed to create image %s", image.Url)
	}

	return nil
}

func (conn *Conn) UpdateImage(image *Image) error {
	err := conn.db.Save(&image).Error
	if err != nil {
		return errors.Wrapf(conn.db.Error, "failed to update image %s", image.Url)
	}
	return nil
}

func (conn *Conn) GetUndownloadedImages() ([]Image, error) {
	images := []Image{}
	err := conn.db.Where("disk_path == ? OR disk_path IS NULL", "").Find(&images).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get downloaded images")
	}

	return images, nil
}

func (conn *Conn) GetUnsentImages() ([]Image, error) {
	images := []Image{}
	err := conn.db.Where("not sent").Find(&images).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get unsent images")
	}

	return images, nil

}
