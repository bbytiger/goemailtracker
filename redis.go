package emailtracker

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// extend to connect your external service
type ExternalConnector interface {
	QueryTrackerStatus(emailId string)
	NotifyExternal(metadata *MailMetadata)
}

// example redis connector implementation
type RedisConnector struct {
	Ctx    context.Context
	Client *redis.Client
}

func NewConnector(opt *redis.Options) *RedisConnector {
	return &RedisConnector{
		Ctx:    context.Background(),
		Client: redis.NewClient(opt),
	}
}

// /
func (rc RedisConnector) QueryTrackerStatus(emailId string) {

}

func (rc RedisConnector) NotifyExternal(metadata *MailMetadata) {
	val, err := rc.Client.Get(rc.Ctx, metadata.SenderInfo.EmailId).Result()
	
	switch action := metadata.Action; action {
	case AppendPixel:
		if err != redis.Nil {

		} else {

		}
	case ServePixel:
		
	}

	rc.Client.Set()
}
