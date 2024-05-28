package webengine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"time"

	"github.com/8xmx8/easier/pkg/engine/utils"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// 网站下载引擎
type Browser struct {
	brow *rod.Browser
}
type BrowserOption func(*Browser) error

func BrowserWithRemoteDevice(remote string) BrowserOption {
	return func(b *Browser) error {
		l := launcher.MustNewManaged(remote).MustLaunch()
		b.brow = b.brow.ControlURL(l)
		return nil
	}
}

func NewBrowser(ctx context.Context, ops ...BrowserOption) (*Browser, error) {
	var err error
	brow := &Browser{
		brow: rod.New().Context(ctx).WithPanic(func(i interface{}) {
			err = fmt.Errorf("创建网站引擎:err=%s", i)
		}),
	}
	for _, op := range ops {
		if inErr := op(brow); inErr != nil {
			return nil, inErr
		}
	}
	brow.brow = brow.brow.MustConnect()
	if err != nil {
		return nil, err
	}
	return brow, nil
}

// Close 关闭浏览器
// defer brow.Close()
func (brow *Browser) Close() {
	brow.brow.MustClose()
}

func (brow *Browser) LoadPage(ctx context.Context, urlPath string) (*BrowserPage, error) {
	var err error
	bp := brow.brow.MustPage(urlPath).Context(ctx).WithPanic(func(i interface{}) {
		err = fmt.Errorf("创建网站引擎:err=%s", i)
	}).MustWaitLoad()
	if err != nil {
		return nil, err
	}
	page := &BrowserPage{
		urlPath: urlPath,
		page:    bp,
	}
	return page, nil
}

type BrowserPage struct {
	page    *rod.Page
	urlPath string
}

// Close 关闭页面
func (p *BrowserPage) Close() {
	p.page.MustClose()
}

// GetTitle 获取标题
func (p *BrowserPage) GetTitle() string {
	return p.page.MustInfo().Title
}

// GetHTML 获取标题
func (p *BrowserPage) GetHTML() (string, error) {
	return p.page.HTML()
}

// GetIcon 获取Icon
func (p *BrowserPage) GetIcon() (string, []byte, error) {
	linkTags := p.page.MustElements("link")
	// 在 <link> 标签中查找图标链接
	var iconHref string
	for _, tag := range linkTags {
		rel, _ := tag.Attribute("rel")
		href, _ := tag.Attribute("href")
		// 检查是否是图标链接
		if *rel == "icon" || *rel == "shortcut icon" {
			iconHref = *href
			break
		}
	}
	if iconHref == "" {
		return "", nil, errors.New("解析不到icon")
	}
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
// isFull: 截全图, 滚动加载
// isFlag: 截图增加url地址栏
func (p *BrowserPage) ScreenshotPage(isFull, isFlag bool) (image.Image, error) {
	var screenshot []byte
	if isFull {
		screenshot = p.page.MustScreenshotFullPage()
	} else {
		screenshot = p.page.MustScrollScreenshotPage()
	}
	reader := bytes.NewReader(screenshot)
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
