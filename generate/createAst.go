package generate

import (
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"mvc/service"
	"time"
)

func CreateAst(c *gin.Context) {
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	defer cancel()

	var pdfBuf []byte

	// 设置等待加载完成的超时时间
	waitTimeout := 70 * time.Second
	ctx, cancel = context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.miluokou.com"),
		//chromedp.WaitVisible("#head_wrapper .s_btn"), // Wait for the element to be visible
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().Do(ctx)
			return err
		}),
	)
	if err != nil {
		service.LogInfo(err)
	}

	if len(pdfBuf) > 0 {
		err = ioutil.WriteFile("output3.pdf", pdfBuf, 0644)
		if err != nil {
			service.LogInfo(err)
		}
	} else {
		service.LogInfo("Empty PDF buffer")
	}
}
