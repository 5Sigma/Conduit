package mailbox

import (
	"time"
)

type Message struct {
	Id             string
	Body           string
	ReceiveCount   int64
	Mailbox        string
	CreatedAt      time.Time
	LastReceivedAt time.Time
	Deployment     string
	Deleted        bool
}
