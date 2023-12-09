package generate

import (
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mvc/service"
	"time"
)

func CreateAst(c *gin.Context) {
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("enable-automation", false),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var pdfBuf []byte

	// 设置等待加载完成的超时时间
	waitTimeout := 70 * time.Second
	ctx, cancel = context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate("http://127.0.0.1:9090/base_report"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			eval := chromedp.Evaluate(`
				document.documentElement.style.contentVisibility = 'auto';
				document.body.style.transform = 'scale(0.8)'; // 缩小页面 (根据需要更改缩放比例)
				document.body.style.margin = '10px'; // 调整页边距（根据需要更改数值）
			`, nil)
			err := eval.Do(ctx)
			if err != nil {
				return err
			}

			// 等待资源加载完成
			err = chromedp.WaitVisible("body", chromedp.ByQuery).Do(ctx)
			if err != nil {
				return err
			}

			// 增加延迟以等待地图背景加载完成
			time.Sleep(2 * time.Second)

			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPaperWidth(10).Do(ctx)
			return err
		}),
	)

	if err != nil {
		service.LogInfo(err)
	}

	if len(pdfBuf) > 0 {
		err = ioutil.WriteFile("output8.pdf", pdfBuf, 0644)
		if err != nil {
			service.LogInfo(err)
		}
	} else {
		service.LogInfo("Empty PDF buffer")
	}
}
