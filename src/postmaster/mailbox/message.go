package mailbox

import (
	"time"
)

type Message struct {
	Id             string
	Body           string
	ReceiveCount   int64
	MailboxId      string
	CreatedAt      time.Time
	lastReceivedAt time.Time
}
