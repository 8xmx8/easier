package chromeengine

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg" // 图片解析器
	_ "image/png"  // 图片解析器
	"io"
	"net/http"
	"time"

	"github.com/8xmx8/easier/pkg/engine/utils"
	"github.com/8xmx8/easier/pkg/logger"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	shotQuality = 90 // 截图质量
	dialTimeout = time.Minute
)

type Butterfly struct {
	ctx         context.Context
	logger      logger.Logger
	isDebug     bool
	shotQuality int
}
type ButterflyOptionFunc func(*Butterfly)

// [0..100]
func SetShotQuality(q int) ButterflyOptionFunc {
	return func(b *Butterfly) {
		if q < 0 {
			b.shotQuality = 0
		} else if q > 100 {
			b.shotQuality = 100
		} else {
			b.shotQuality = q
		}
	}
}

func IsDebugWithButterfly(f bool) ButterflyOptionFunc {
	return func(b *Butterfly) {
		b.isDebug = f
	}
}
func WithLogger(logg logger.Logger) ButterflyOptionFunc {
	return func(b *Butterfly) {
		b.logger = logg
	}
}

func NewButterfly(ctx context.Context, ops ...ButterflyOptionFunc) (*Butterfly, context.CancelFunc, error) {
	bf := &Butterfly{
		logger:      logger.DefaultLogger(),
		isDebug:     false,
		shotQuality: shotQuality,
	}
	chromeOption := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.IgnoreCertErrors,
		chromedp.DisableGPU,
		chromedp.WindowSize(1920, 1080),
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
	)
	if bf.isDebug {
		chromeOption = append(chromeOption, chromedp.Flag("headless", false))
	}
	// create a new browser
	ctx, cancel := chromedp.NewExecAllocator(ctx, chromeOption...)
	bf.ctx = ctx
	for _, op := range ops {
		op(bf)
	}
	return bf, cancel, nil
}

type BrowserPage struct {
	urlPath     string
	htmlContent string
	iconLink    string
	title       string
	snapshot    []byte
}

func (b *Butterfly) LoadPage(ctx context.Context, urlPath string) (*BrowserPage, error) {
	runCtx, cancel := context.WithTimeout(b.ctx, dialTimeout)
	defer cancel()
	subCtx, chromeCancel := chromedp.NewContext(runCtx)
	defer chromeCancel()
	page := &BrowserPage{
		urlPath:     urlPath,
		htmlContent: "",
		iconLink:    "",
		title:       "",
		snapshot:    []byte{},
	}
	tasks := chromedp.Tasks{}
	// 获取icon
	// DOC: https://github.com/chromedp/chromedp/issues/1054
	js := `
		const iconElement = document.querySelector("link[rel~=icon]");
		const href = (iconElement && iconElement.href) || "/favicon.ico";
		const faviconURL = new URL(href, window.location).toString();
		faviconURL`
	tasks = append(tasks, chromedp.Evaluate(js, &page.iconLink))
	// 获取title
	tasks = append(tasks, chromedp.Title(&page.title))
	// 获取html
	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			b.logger.Error(logger.ErrorAgentStart, "get document node", logger.ErrorField(err))
			return err
		}
		page.htmlContent, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		if err != nil {
			b.logger.Error(logger.ErrorAgentStart, "get document outer HTML", logger.ErrorField(err))
			return err
		}
		return nil
	}))
	// 截图
	tasks = append(tasks, chromedp.FullScreenshot(&page.snapshot, b.shotQuality))
	// Run
	b.logger.Info("[butterfly]开始获取目标网站", logger.MakeField("target", urlPath))
	runOptions := []chromedp.Action{chromedp.Navigate(urlPath), tasks}
	if b.isDebug {
		runOptions = append(runOptions, chromedp.Sleep(5*time.Second))
	}
	if err := chromedp.Run(subCtx, runOptions...); err != nil {
		b.logger.Error(logger.ErrorAgentStart, "[butterfly]页面请求错误", logger.ErrorField(err),
			logger.MakeField("target", urlPath))
		return nil, err
	}
	return page, nil
}

// GetTitle 获取标题
func (p *BrowserPage) GetTitle() string {
	return p.title
}

// GetHTML 获取标题
func (p *BrowserPage) GetHTML() string {
	return p.htmlContent
}

// GetIcon 获取Icon
func (p *BrowserPage) GetIcon() (string, []byte, error) {
	iconHref := p.iconLink
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iconHref, http.NoBody)
	if err != nil {
		return iconHref, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return iconHref, nil, err
	}
	defer resp.Body.Close()
	iconB, err := io.ReadAll(resp.Body)
	return iconHref, iconB, err
}

// ScreenshotFullPage 截全图
// isFlag: 截图增加url地址栏
func (p *BrowserPage) ScreenshotPage(isFlag bool) (image.Image, error) {
	reader := bytes.NewReader(p.snapshot)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	if isFlag {
		newImg, err := utils.AddAddressBar(img, p.urlPath)
		if err == nil {
			img = newImg
		}
	}
	return img, nil
}
