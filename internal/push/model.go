package push

import "time"

type PushSubscription struct {
	ID           int64      `db:"id"`
	UserID       string     `db:"user_id"`  
	Endpoint     string     `db:"endpoint"`
	P256dh       string     `db:"p256dh"`
	Auth         string     `db:"auth"`
	UserAgent    string     `db:"user_agent"`
    CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}