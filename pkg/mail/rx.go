package mail

import (
	"github.com/pkg/errors"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"k8s.io/klog"

	"github.com/squeed/geezo-bot/pkg/config"
)

type IMAPConn struct {
	config *config.Config
	Conn   *client.Client
}

type GetMessagesResult struct {
	Messages map[uint32]*Message
	Unknown  imap.SeqSet
	Fail     imap.SeqSet
}

type Dest int

const (
	DestDone = iota
	DestUnkown
)

func IMAPConnect(c *config.Config) (*IMAPConn, error) {
	klog.Infof("opening IMAP connection to %s", c.Imap.Server)
	ic, err := client.DialTLS(c.Imap.Server, nil)
	if err != nil {
		klog.Errorf("failed to connect: %s", err)
		return nil, errors.Wrap(err, "imap connection failed")
	}

	klog.Infof("logging in as %s", c.Imap.Username)
	err = ic.Login(c.Imap.Username, c.Imap.Password)
	if err != nil {
		klog.Errorf("failed to log in: %s", err)
		return nil, errors.Wrap(err, "imap login failed")
	}
	klog.V(3).Info("IMAP authentication successful")

	return &IMAPConn{
		Conn:   ic,
		config: c,
	}, nil
}

// ListMessages retrieves all the unread messages from the server
func (conn *IMAPConn) GetMessages() (*GetMessagesResult, error) {
	mbox, err := conn.Conn.Select("INBOX", false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select INBOX")
	}
	out := GetMessagesResult{
		Messages: make(map[uint32]*Message, mbox.Messages),
	}

	if mbox.Messages == 0 {
		klog.Info("mailbox is empty!")
		return &out, nil
	}

	// Get the last 4 messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	msgchan := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		section := &imap.BodySectionName{}
		done <- conn.Conn.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope, imap.FetchUid,
			section.FetchItem(),
		},
			msgchan)
	}()

	for msg := range msgchan {
		if !ShouldProcessMessage(msg) {
			klog.Infof("skipping message %s", msg.Envelope.Subject)
			out.Unknown.AddNum(msg.Uid)
			continue
		}
		m, err := NewMessage(msg)
		if err != nil {
			klog.Error("failed to parse message, moving to fail", err)
			out.Fail.AddNum(msg.Uid)
			continue
		}
		out.Messages[msg.Uid] = m
	}

	if err := <-done; err != nil {
		klog.Error(err)
		return nil, errors.Wrap(err, "failed to retrieve messages")
	}

	klog.Infof("%d messages to process", len(out.Messages))

	return &out, nil
}

// MoveMessages moves processed messages
func (conn *IMAPConn) MoveMessages(conf *config.Config, res *GetMessagesResult) error {

	klog.Info("moving failure messages")
	if err := conn.move(&res.Fail, "INBOX.error"); err != nil {
		return errors.Wrap(err, "failed to move failure messages")
	}

	klog.Info("moving unknown messages")
	if err := conn.move(&res.Unknown, "INBOX.unknown"); err != nil {
		return errors.Wrap(err, "failed to move unknown messages")
	}

	ss := imap.SeqSet{}
	for uid := range res.Messages {
		ss.AddNum(uid)
	}
	klog.Infof("moving done messages")
	if err := conn.move(&ss, "INBOX.done"); err != nil {
		return errors.Wrap(err, "failed to move done messages")
	}

	return nil
}

func (conn *IMAPConn) move(set *imap.SeqSet, dest string) error {
	if set.Empty() {
		klog.Info("skipping empty set")
		return nil
	}

	_, err := conn.Conn.Select("INBOX", false)
	if err != nil {
		return errors.Wrap(err, "failed to select")
	}

	err = conn.Conn.UidCopy(set, dest)
	if err != nil {
		return errors.Wrap(err, "move: failed to copy messages")
	}

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := conn.Conn.UidStore(set, item, flags, nil); err != nil {
		return errors.Wrap(err, "move: failed to flag for deletion")
	}

	if err := conn.Conn.Expunge(nil); err != nil {
		return errors.Wrap(err, "move: failed to expunge old messages")
	}

	return nil

}

func (res *GetMessagesResult) FailMessage(uid uint32) {
	delete(res.Messages, uid)
	res.Fail.AddNum(uid)
}

func (conn *IMAPConn) Close() {
	conn.Conn.Logout()
}
