package jobs

import (
	"fmt"
	"log"
	"mvc/service"
)

func TestGetMsgFromKafka() {
	messages, err := service.ConsumeMessages()
	if err != nil {
		log.Println("无法消费消息:", err)
		return
	}

	for _, message := range messages {
		fmt.Println(message.Value)
	}
}
