package components

func NewPubsub() *Pubsub {
	return &Pubsub{
		ready: make(chan struct{}),
	}
}

type Pubsub struct {
	Pass     string
	Port     string
	DB       []string
	Hostname string
	Network  string

	ready chan struct{}
}

func (c *Pubsub) Start() {
	defer close(c.ready)
	cmd := `docker run --rm -d --network=mtf_net --name pubsub_mtf --hostname=pubsub_mtf -p 8085:8085 adilsoncarvalho/gcloud-pubsub-emulator`
	run(cmd)
}

func (c *Pubsub) Stop() {
	run("docker stop pubsub_mtf")
}

func (c *Pubsub) Ready() {
	<-c.ready
}