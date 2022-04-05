package configs

import (
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

func init() {
	lvl, ok := os.LookupEnv("FLOW_WALLET_LOG_LEVEL")
	if !ok {
		// LOG_LEVEL not set, default to info
		lvl = "info"
	}

	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.DebugLevel
	}

	log.SetLevel(ll)

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

type Config struct {
	// -- Feature flags --

	DisableRawTransactions   bool `env:"FLOW_WALLET_DISABLE_RAWTX"`
	DisableFungibleTokens    bool `env:"FLOW_WALLET_DISABLE_FT"`
	DisableNonFungibleTokens bool `env:"FLOW_WALLET_DISABLE_NFT"`
	DisableChainEvents       bool `env:"FLOW_WALLET_DISABLE_CHAIN_EVENTS"`

	// -- Admin account --

	AdminAddress    string `env:"FLOW_WALLET_ADMIN_ADDRESS,notEmpty"`
	AdminKeyIndex   int    `env:"FLOW_WALLET_ADMIN_KEY_INDEX" envDefault:"0"`
	AdminKeyType    string `env:"FLOW_WALLET_ADMIN_KEY_TYPE" envDefault:"local"`
	AdminPrivateKey string `env:"FLOW_WALLET_ADMIN_PRIVATE_KEY,notEmpty"`
	// This sets the number of proposal keys to be used on the admin account.
	// You can increase transaction throughput by using multiple proposal keys for
	// parallel transaction execution.
	AdminProposalKeyCount uint16 `env:"FLOW_WALLET_ADMIN_PROPOSAL_KEY_COUNT" envDefault:"1"`

	// -- Keys --

	// When "DefaultKeyType" is set to "local", private keys are generated by the API
	// and stored as encrypted text in the database.
	// KMS key types:
	// - aws_kms
	// - google_kms
	DefaultKeyType  string `env:"FLOW_WALLET_DEFAULT_KEY_TYPE" envDefault:"local"`
	DefaultKeyIndex int    `env:"FLOW_WALLET_DEFAULT_KEY_INDEX" envDefault:"0"`
	// If the default of "-1" is used for "DefaultKeyWeight"
	// the service will use flow.AccountKeyWeightThreshold from the Flow SDK.
	DefaultKeyWeight int    `env:"FLOW_WALLET_DEFAULT_KEY_WEIGHT" envDefault:"-1"`
	DefaultSignAlgo  string `env:"FLOW_WALLET_DEFAULT_SIGN_ALGO" envDefault:"ECDSA_P256"`
	DefaultHashAlgo  string `env:"FLOW_WALLET_DEFAULT_HASH_ALGO" envDefault:"SHA3_256"`
	// This symmetrical key is used to encrypt private keys
	// that are stored in the database. Values per type:
	// - local: 32 bytes long encryption key
	// - aws_kms: key ARN, e.g. arn:aws:kms:us-west-1:123456789000:key/00000000-1111-2222-3333-444444444444
	// - google_kms: key resource name (without version info), e.g. projects/my-project/locations/europe-north1/keyRings/my-keyring/cryptoKeys/my-encryption-key
	EncryptionKey string `env:"FLOW_WALLET_ENCRYPTION_KEY,notEmpty"`
	// Encryption key type, one of: local, aws_kms, google_kms
	EncryptionKeyType string `env:"FLOW_WALLET_ENCRYPTION_KEY_TYPE,notEmpty" envDefault:"local"`
	// DefaultAccountKeyCount specifies how many times the account key will be duplicated upon account creation, does not affect existing accounts
	DefaultAccountKeyCount uint `env:"FLOW_WALLET_DEFAULT_ACCOUNT_KEY_COUNT" envDefault:"1"`

	// -- Database --

	DatabaseDSN     string `env:"FLOW_WALLET_DATABASE_DSN" envDefault:"wallet.db"`
	DatabaseType    string `env:"FLOW_WALLET_DATABASE_TYPE" envDefault:"sqlite"`
	DatabaseVersion string `env:"FLOW_WALLET_DATABASE_VERSION" envDefault:""`

	// -- Host and chain access --

	Host                 string        `env:"FLOW_WALLET_HOST"`
	Port                 int           `env:"FLOW_WALLET_PORT" envDefault:"3000"`
	ServerRequestTimeout time.Duration `env:"FLOW_WALLET_SERVER_REQUEST_TIMEOUT" envDefault:"60s"`
	AccessAPIHost        string        `env:"FLOW_WALLET_ACCESS_API_HOST,notEmpty"`
	ChainID              flow.ChainID  `env:"FLOW_WALLET_CHAIN_ID" envDefault:"flow-emulator"`

	// -- Templates --

	EnabledTokens           []string `env:"FLOW_WALLET_ENABLED_TOKENS" envSeparator:","`
	ScriptPathCreateAccount string   `env:"FLOW_WALLET_SCRIPT_PATH_CREATE_ACCOUNT" envDefault:""`

	// -- Workerpool --

	// Defines the maximum number of active jobs that can be queued before
	// new jobs are rejected.
	WorkerQueueCapacity uint `env:"FLOW_WALLET_WORKER_QUEUE_CAPACITY" envDefault:"1000"`
	// Number of concurrent workers handling incoming jobs.
	// You can increase the number of workers if you're sending
	// too many transactions and find that the queue is often backlogged.
	WorkerCount uint `env:"FLOW_WALLET_WORKER_COUNT" envDefault:"1"`
	// Webhook endpoint to receive job status updates
	JobStatusWebhookUrl string `env:"FLOW_WALLET_JOB_STATUS_WEBHOOK" envDefault:""`
	// Duration for which to wait for a response, if 0 wait indefinitely. Default: 30s.
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// For more info: https://pkg.go.dev/time#ParseDuration
	JobStatusWebhookTimeout time.Duration `env:"FLOW_WALLET_JOB_STATUS_WEBHOOK_TIMEOUT" envDefault:"30s"`

	// -- Google KMS --

	GoogleKMSProjectID  string `env:"FLOW_WALLET_GOOGLE_KMS_PROJECT_ID"`
	GoogleKMSLocationID string `env:"FLOW_WALLET_GOOGLE_KMS_LOCATION_ID"`
	GoogleKMSKeyRingID  string `env:"FLOW_WALLET_GOOGLE_KMS_KEYRING_ID"`

	// -- Misc --

	// Duration for which to wait for a transaction seal, if 0 wait indefinitely. Default: 0.
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// For more info: https://pkg.go.dev/time#ParseDuration
	TransactionTimeout time.Duration `env:"FLOW_WALLET_TRANSACTION_TIMEOUT" envDefault:"0"`

	// Idempotency middleware configuration
	DisableIdempotencyMiddleware bool `env:"FLOW_WALLET_DISABLE_IDEMPOTENCY_MIDDLEWARE" envDefault:"false"`
	// Idempotency middleware database type;
	// - "local", in-memory w/ no multi-instance support
	// - "shared", sql (gorm) database shared with the app (DatabaseType)
	// - "redis"
	IdempotencyMiddlewareDatabaseType string `env:"FLOW_WALLET_IDEMPOTENCY_MIDDLEWARE_DATABASE_TYPE" envDefault:"local"`
	// Redis URL for idempotency key storage, e.g. "redis://walletapi:wallet-api-redis@localhost:6379/"
	IdempotencyMiddlewareRedisURL string `env:"FLOW_WALLET_IDEMPOTENCY_MIDDLEWARE_REDIS_URL" envDefault:""`

	// Set the starting height for event polling. This won't have any effect if the value in
	// database (chain_event_status[0].latest_height) is greater.
	// If 0 (default) use latest block height if starting fresh (no previous value in database).
	ChainListenerStartingHeight uint64 `env:"FLOW_WALLET_EVENTS_STARTING_HEIGHT" envDefault:"0"`
	// Maximum number of blocks to check at once.
	ChainListenerMaxBlocks uint64 `env:"FLOW_WALLET_EVENTS_MAX_BLOCKS" envDefault:"100"`
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// For more info: https://pkg.go.dev/time#ParseDuration
	ChainListenerInterval time.Duration `env:"FLOW_WALLET_EVENTS_INTERVAL" envDefault:"10s"`

	// Max transactions per second, rate at which the service can submit transactions to Flow
	TransactionMaxSendRate int `env:"FLOW_WALLET_MAX_TPS" envDefault:"10"`

	// maxJobErrorCount is the maximum number of times a Job can be tried to
	// execute before considering it completely failed.
	MaxJobErrorCount int `env:"MAX_JOB_ERROR_COUNT" envDefault:"10"`

	// Poll DB for new schedulable jobs every 30s.
	DBJobPollInterval time.Duration `env:"DB_JOB_POLL_INTERVAL" envDefault:"30s"`

	// Grace time period before re-scheduling jobs that are in state INIT or
	// ACCEPTED. These are jobs where the executor processing has been
	// unexpectedly disrupted (such as bug, dead node, disconnected
	// networking etc.).
	AcceptedGracePeriod time.Duration `env:"ACCEPTED_GRACE_PERIOD" envDefault:"180s"`

	// Grace time period before re-scheduling jobs that are up for immediate
	// restart (such as NO_AVAILABLE_WORKERS or ERROR).
	ReSchedulableGracePeriod time.Duration `env:"RESCHEDULABLE_GRACE_PERIOD" envDefault:"60s"`

	// Sleep duration in case of service isHalted
	PauseDuration time.Duration `env:"PAUSE_DURATION" envDefault:"60s"`
}

type Options struct {
	EnvFilePath string
	Version     string
}

// ParseConfig parses environment variables and flags to a valid Config.
func ParseConfig(opt *Options) (*Config, error) {
	if opt != nil && opt.EnvFilePath != "" {
		// Load variables from a file to the environment of the process
		if err := godotenv.Load(opt.EnvFilePath); err != nil {
			log.
				WithFields(log.Fields{"error": err}).
				Warn("Could not load environment variables from file. If running inside a docker container this can be ignored.")
		}
	}

	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
