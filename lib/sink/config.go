package sink

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
)

const (
	QueueTypeBeastAll    = "beast-all"
	QueueTypeBeastReduce = "beast-reduce"
	QueueTypeAvrAll      = "avr-all"
	QueueTypeAvrReduce   = "avr-reduce"
	QueueTypeSbs1All     = "sbs1-all"
	QueueTypeSbs1Reduce  = "sbs1-reduce"
	QueueLocationUpdates = "location-updates"
)

var AllQueues = [...]string{
	QueueTypeBeastAll,
	QueueTypeBeastReduce,
	QueueTypeAvrAll,
	QueueTypeAvrReduce,
	QueueTypeSbs1All,
	QueueTypeSbs1Reduce,
	QueueLocationUpdates,
}

type (
	Config struct {
		host, port string
		secure     bool

		vhost      string
		user, pass string
		queue      map[string]string

		waiter sync.WaitGroup

		sourceTag         string
		connectionName    string
		messageTtlSeconds int

		createTestQueues bool

		stats struct {
			dedupeFrame, frame, planeLoc prometheus.Counter
		}

		// for remembering if we have recently sent this message
	}

	Option func(*Config)
)

func (c *Config) setupConfig(opts []Option) {
	c.queue = map[string]string{}
	for _, opt := range opts {
		opt(c)
	}
}

func WithConnectionName(name string) Option {
	return func(conf *Config) {
		conf.connectionName = name
	}
}
func WithHost(host, port string) Option {
	return func(conf *Config) {
		conf.host = host
		conf.port = port
	}
}
func WithUserPass(user, pass string) Option {
	return func(conf *Config) {
		conf.user = user
		conf.pass = pass
	}
}

func WithSourceTag(tag string) Option {
	return func(config *Config) {
		config.sourceTag = tag
	}
}

func WithMessageTtl(ttl int) Option {
	return func(config *Config) {
		if ttl >= 0 {
			config.messageTtlSeconds = ttl
		}
	}
}

func WithLogFile(file string) Option {
	return func(config *Config) {
		f, err := os.Create(file)
		if nil != err {
			println("Cannot open file: ", file)
			return
		}
		log.Logger = zerolog.New(f).With().Timestamp().Logger()
	}
}

func WithPrometheusCounters(frame, dedupeFrame, planeLoc prometheus.Counter) Option {
	return func(conf *Config) {
		conf.stats.frame = frame
		conf.stats.dedupeFrame = dedupeFrame
		conf.stats.planeLoc = planeLoc
	}
}

func (c *Config) Finish() {
	c.waiter.Wait()
}

func WithQueues(queues []string) Option {
	return func(conf *Config) {
		if 0 == len(queues) {
			WithAllQueues()(conf)
			log.Debug().Msg("With all output types")
			return
		}

		for _, requestedQueue := range queues {
			found := false
			for _, validQueue := range AllQueues {
				if requestedQueue == validQueue {
					log.Debug().Str("publish-type", requestedQueue).Msg("With publish type")
					conf.queue[validQueue] = validQueue
					found = true
					break
				}
			}
			if !found {
				log.Error().Msgf("Error: Unknown Queue Type: %s", requestedQueue)
			}
		}
	}
}
func WithAllQueues() Option {
	return func(conf *Config) {
		conf.queue[QueueTypeAvrAll] = QueueTypeAvrAll
		conf.queue[QueueTypeAvrReduce] = QueueTypeAvrReduce
		conf.queue[QueueTypeBeastAll] = QueueTypeBeastAll
		conf.queue[QueueTypeBeastReduce] = QueueTypeBeastReduce
		conf.queue[QueueTypeSbs1All] = QueueTypeSbs1All
		conf.queue[QueueTypeSbs1Reduce] = QueueTypeSbs1Reduce
		conf.queue[QueueLocationUpdates] = QueueLocationUpdates
	}
}
