package cron

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCron(t *testing.T) {
	//TODO 实际生产中不建议使用，换一个好的上下文
	ctx := context.TODO()
	client, err := NewClient(ctx)
	assert.NoError(t, err)
	signal := make(chan struct{})
	cmd := func() {
		fmt.Println("定时开始任务执行")
		a := make([]int, 10)
		a[0] = 1
		for i := 1; i < 10; i++ {
			a[i] = a[i] + a[i-1]
		}
		signal <- struct{}{}
	}
	t.Log("阻塞线程，等待定时任务执行")
	// 每2分钟执行一次，第一次执行接受信号，退出线程
	_, err = client.AddFunc(ctx, "0 0/2 * * * ?", cmd)
	assert.NoError(t, err)
	client.Start(ctx)
	<-signal
	t.Log("定时任务执行完毕")
}
