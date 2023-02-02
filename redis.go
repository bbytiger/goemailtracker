package emailtracker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// extend to connect your external service
type ExternalConnector interface {
	QueryTrackerStatus(emailId string) (Status, error)
	NotifyExternal(metadata *MailMetadata) error
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

type MailMetadataCompressed struct {
	Timestamp    time.Time `json:"timestamp"`
	UserAgent    string    `json:"user_agent"`
	UserIP       string    `json:"user_ip"`
	Action       string    `json:"action"`
	StatusUpdate string    `json:"status"`
	HTML         string    `json:"html_result"`
	SenderId     string    `json:"sender_id"`
	SenderEmail  string    `json:"sender_email"`
	RecvEmail    string    `json:"recv_email"`
	EmailId      string    `json:"email_id"`
}

func (metadata *MailMetadata) Marshal() ([]byte, error) {
	return json.Marshal(&MailMetadataCompressed{
		Timestamp:    metadata.Timestamp,
		UserAgent:    metadata.UserAgent,
		UserIP:       metadata.UserIP,
		Action:       string(metadata.Action),
		StatusUpdate: string(metadata.StatusUpdate),
		HTML:         metadata.HTML,
		SenderId:     metadata.SenderInfo.SenderId,
		SenderEmail:  metadata.SenderInfo.SenderEmail,
		RecvEmail:    metadata.SenderInfo.RecvEmail,
		EmailId:      metadata.SenderInfo.EmailId,
	})
}

func (rc RedisConnector) QueryTrackerStatus(emailId string) (Status, error) {
	val, err := rc.Client.HGetAll(rc.Ctx, emailId).Result()
	if err != nil {
		return Untracked, err
	}
	return Status(val["current_status"]), nil
}

func (rc RedisConnector) NotifyExternal(metadata *MailMetadata) error {
	if metadata.SenderInfo == nil {
		return errors.New("no identifying PII found")
	}
	val, err := rc.Client.HGetAll(rc.Ctx, metadata.SenderInfo.EmailId).Result()
	actionKey := fmt.Sprintf("%s_%s", string(metadata.Action), metadata.Timestamp)
	compressedData, marshalErr := metadata.Marshal()
	if marshalErr != nil {
		return marshalErr
	}

	var values map[string]interface{}

	switch action := metadata.Action; action {
	case AppendPixel:
		if err == redis.Nil {
			values = map[string]interface{}{
				"open_count":     "0",
				"append_count":   "1",
				"current_status": string(Attached),
				actionKey:        string(compressedData),
			}
		} else if err != nil {
			return err
		} else {
			prevCount, parseErr := strconv.ParseInt(val["append_count"], 10, 32)
			if parseErr != nil {
				return parseErr
			}
			values = map[string]interface{}{
				"append_count": strconv.Itoa(int(prevCount) + 1),
				actionKey:      string(compressedData),
			}
		}
	case ServePixel:
		if err == redis.Nil {
			return errors.New("serve pixel called before email appended")
		} else if err != nil {
			return err
		} else {
			prevCount, parseErr := strconv.ParseInt(val["open_count"], 10, 32)
			if parseErr != nil {
				return parseErr
			}
			currentStatus := val["current_status"]
			if currentStatus != string(Responded) {
				currentStatus = string(Opened)
			}
			values = map[string]interface{}{
				"open_count":     strconv.Itoa(int(prevCount) + 1),
				"current_status": currentStatus,
				actionKey:        string(compressedData),
			}
		}
	}

	return rc.Client.HSet(rc.Ctx, metadata.SenderInfo.EmailId, values).Err()
}
