package jobs

import (
	"log"
	"mvc/service"
)

type PropertyData struct {
	ID            uint   `json:"id"`
	YearInfo      string `json:"year_info"`
	CommunityName string `json:"community_name"`
	AddressInfo   string `json:"address_info"`
	PricePerSqm   string `json:"price_per_sqm"`
	PageNumber    string `json:"page_number"`
	Deal          string `json:"deal"`
	City          string `json:"city"`
	QuText        string `json:"qu_text"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func KafkaToEs() {

	messages, err := service.ConsumeMessages()
	if err != nil {
		log.Println("无法消费消息:", err)
		return
	}
	service.EsEnv(messages)
	// 解析出的message 然后存储到es 中

}
