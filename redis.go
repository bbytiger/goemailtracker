package emailtracker

// extend to connect your external service
type ExternalConnector interface {
	NotifyExternal(metadata *MailMetadata)
}

// example redis connector implementation
type RedisConnector struct {
}

func ConfigureRedisWithEnv() *RedisConnector {
	return &RedisConnector{}
}

func ConfigureRedisManual() *RedisConnector {
	return &RedisConnector{}
}

func (rc RedisConnector) NotifyExternal(metadata *MailMetadata) {

}
