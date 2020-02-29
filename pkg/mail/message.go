package mail

import (
	"io"
	"io/ioutil"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

type Message struct {
	Msg     *imap.Message
	Content []byte
}

func NewMessage(im *imap.Message) (*Message, error) {
	r := im.GetBody(&imap.BodySectionName{})
	if r == nil {
		return nil, errors.New("missing body section")
	}

	mm, err := message.Read(r)
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, errors.Wrapf(err, "failed to parse message")
	}

	out := &Message{Msg: im}
	found := false

	if mr := mm.MultipartReader(); mr != nil {
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, errors.Wrap(err, "failed to parse message part")
			}

			t, _, _ := p.Header.ContentType()
			if t != "text/html" {
				continue
			}
			out.Content, err = ioutil.ReadAll(p.Body)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read body")
			}
			klog.Infof("Read %d body bytes", len(out.Content))
			found = true
			break
		}
	} else {
		t, _, _ := mm.Header.ContentType()
		log.Println("This is a non-multipart message with type", t)
		if t == "text/html" {
			out.Content, err = ioutil.ReadAll(mm.Body)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read body")
			}
			klog.Infof("Read %d body bytes", len(out.Content))
			found = true

		}
	}
	if !found {
		return nil, errors.New("could not find TinyBeans content in message")
	}

	return out, nil

}

func ShouldProcessMessage(im *imap.Message) bool {
	return im.Envelope.Subject == "Hi Geezo, your Tinybeans daily updates"
}
