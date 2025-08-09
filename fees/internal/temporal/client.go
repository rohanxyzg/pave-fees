package temporal

import (
	"log/slog"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const TaskQueue = "fees-task-queue"

type Client struct {
	client.Client
	worker worker.Worker
}

type ClientOptions struct {
	Target    string
	Namespace string
}

func NewClient(opts ClientOptions) (*Client, error) {
	c, err := client.NewLazyClient(client.Options{
		HostPort:  opts.Target,
		Namespace: opts.Namespace,
	})
	if err != nil {
		return nil, err
	}

	w := worker.New(c, TaskQueue, worker.Options{})
	return &Client{
		Client: c,
		worker: w,
	}, nil
}

func (c *Client) RegisterWorkflow(w interface{}) {
	c.worker.RegisterWorkflow(w)
}

func (c *Client) RegisterActivity(a interface{}) {
	c.worker.RegisterActivity(a)
}

func (c *Client) StartWorker() error {
	slog.Info("Starting Temporal worker")
	err := c.worker.Start()
	if err != nil {
		slog.Error("Failed to start worker", "error", err)
	}

	return err
}
