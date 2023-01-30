package emailtracker

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
