package postgreskit

// just copy from mongokit, not implemented yet!
import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/go-pg/pg/v9"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type PGConfig struct {
	URL string `long:"url" env:"URL" description:"the URL of PostgreSQL" required:"true"`
}

type PGClient struct {
	*pg.DB
	Closed atomic.Bool
}

func (c *PGClient) Close() error {
	c.Closed.Store(true)
	return c.DB.Close()
}

func NewPGClient(ctx context.Context, conf *PGConfig) *PGClient {
	if url := os.ExpandEnv(conf.URL); url != "" {
		conf.URL = url
	}

	logger := logkit.FromContext(ctx).With(zap.String("url", conf.URL))
	opts, err := pg.ParseURL(conf.URL)
	if err != nil {
		logger.Fatal("Failed to parse PostgreSQL url", zap.Error(err))
	}

	db := pg.Connect(opts).WithContext(ctx)

	// client, err := mongo.NewClient(o)
	// if err != nil {
	// 	logger.Fatal("failed to create MongoDB client", zap.Error(err))
	// }

	// if err := client.Connect(ctx); err != nil {
	// 	logger.Fatal("failed to connect to MongoDB", zap.Error(err))
	// }

	// if err := client.Ping(ctx, readpref.Primary()); err != nil {
	// 	logger.Fatal("failed to ping to MongoDB", zap.Error(err))
	// }

	// logger.Info("create MongoDB client successfully")

	// return &MongoClient{
	// 	Client:   client,
	// 	database: client.Database(conf.Database),
	// }
}
