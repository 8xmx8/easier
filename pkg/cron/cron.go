package cron

import (
	"context"

	"github.com/robfig/cron/v3"
)

// EntryID64 is the type of the id of an entry in the cron schedule.
type EntryID64 int64

type Client struct {
	cron *cron.Cron
}

func NewClient(ctx context.Context) (*Client, error) {
	var opts []cron.Option
	opts = append(opts, cron.WithSeconds())
	c := cron.New(opts...)
	return &Client{
		cron: c,
	}, nil
}

func (client *Client) AddFunc(ctx context.Context, spec string, cmd func()) (EntryID64, error) {
	entryID, err := client.cron.AddFunc(spec, cmd)
	return EntryID64(entryID), err
}

func (client *Client) Start(ctx context.Context) {
	client.cron.Start()
}

func (client *Client) Stop() {
	client.cron.Stop()
}

func (client *Client) Remove(id cron.EntryID) {
	client.cron.Remove(id)
}
