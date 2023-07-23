package controllers

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"mvc/models" // 请替换为正确的模型包引用路径
)

type PaymentNotification struct {
	AppID         string `xml:"appid"`
	BankType      string `xml:"bank_type"`
	CashFee       string `xml:"cash_fee"`
	FeeType       string `xml:"fee_type"`
	IsSubscribe   string `xml:"is_subscribe"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	OpenID        string `xml:"openid"`
	OutTradeNo    string `xml:"out_trade_no"`
	ResultCode    string `xml:"result_code"`
	ReturnCode    string `xml:"return_code"`
	Sign          string `xml:"sign"`
	TimeEnd       string `xml:"time_end"`
	TotalFee      string `xml:"total_fee"`
	TradeType     string `xml:"trade_type"`
	TransactionID string `xml:"transaction_id"`
}

func WechatCallback(c *gin.Context) {
	// 读取请求体数据
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Fatal(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// 打印回调信息到控制台
	fmt.Println(string(body))

	// 将XML数据解析为PaymentNotification结构体
	payment := PaymentNotification{}
	err = xml.Unmarshal(body, &payment)
	if err != nil {
		log.Fatal(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// 获取OpenID和订单信息
	openID := payment.OpenID
	orderID := payment.OutTradeNo
	totalFee := payment.TotalFee

	// 创建Order实例并设置回调信息
	order := models.Order{
		OrderNumber:   orderID,
		CallbackTime:  payment.TimeEnd,
		TransactionID: payment.TransactionID,
		PaymentAmount: StringToFloat64(totalFee),
		PaymentStatus: "complete",
		OpenID:        openID, // 设置订单的 OpenID
	}

	// 检查数据库中是否已存在相同的交易订单
	existingOrder, err := models.GetOrderByOrderNumber(orderID)
	if err != nil {
		log.Fatal(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// 如果不存在相同的交易订单，则创建新的订单记录
	if existingOrder == (models.Order{}) {
		createdOrder, err := models.CreateOrder(order)
		if err != nil {
			log.Fatal(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		fmt.Printf("订单已创建：%v\n", createdOrder)
	} else {
		fmt.Println("该交易订单已存在，无需重复存储")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Callback received",
	})
}

func StringToFloat64(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
