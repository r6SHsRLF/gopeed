package main

import (
	"fmt"
	"github.com/gen2brain/beeep" // 导入 beeep 库来发送通知
	"github.com/GopeedLab/gopeed/internal/fetcher"
	"github.com/GopeedLab/gopeed/pkg/base"
	"github.com/GopeedLab/gopeed/pkg/download"
	"github.com/GopeedLab/gopeed/pkg/protocol/http"
	"github.com/GopeedLab/gopeed/pkg/util"
	"strings"
	"sync"
)

const progressWidth = 20

// 主函数
func main() {
	// 解析参数
	args := parse()

	var wg sync.WaitGroup
	wg.Add(1)

	// 设置下载任务
	_, err := download.Boot().
		URL(args.url).
		Listener(func(event *download.Event) {
			// 下载过程中显示进度
			if event.Key == download.EventKeyProgress {
				printProgress(event.Task, "downloading...")
			}

			// 下载完成时触发的事件
			if event.Key == download.EventKeyFinally {
				var title string
				if event.Err != nil {
					title = "fail" // 下载失败
				} else {
					title = "complete" // 下载成功
				}

				// 打印下载进度
				printProgress(event.Task, title)
				fmt.Println()
				if event.Err != nil {
					// 下载失败，打印错误信息
					fmt.Printf("reason: %s", event.Err.Error())
				} else {
					// 下载成功，打印保存路径
					fmt.Printf("saving path: %s", *args.dir)
				}

				// 发送桌面通知
				err := beeep.Notify("Download Completed", fmt.Sprintf("Your file download is %s", title), "assets/information.png")
				if err != nil {
					fmt.Println("Error sending notification:", err)
				}

				// 下载完成后通知等待组结束
				wg.Done()
			}
		}).
		// 设置下载选项
		Create(&base.Options{
			Path:  *args.dir,
			Extra: http.OptsExtra{Connections: *args.connections},
		})
	if err != nil {
		panic(err)
	}

	// 初始的进度条
	printProgress(emptyTask, "downloading...")
	// 等待下载完成
	wg.Wait()
}

// 进度条相关
var (
	lastLineLen = 0
	sb          = new(strings.Builder)
	emptyTask   = &download.Task{
		Progress: &download.Progress{},
		Meta: &fetcher.FetcherMeta{
			Res: &base.Resource{},
		},
	}
)

// 打印下载进度的函数
func printProgress(task *download.Task, title string) {
	var rate float64
	if task.Meta.Res == nil {
		task = emptyTask
	}
	if task.Meta.Res.Size <= 0 {
		rate = 0
	} else {
		rate = float64(task.Progress.Downloaded) / float64(task.Meta.Res.Size)
	}
	completeWidth := int(progressWidth * rate)
	speed := util.ByteFmt(task.Progress.Speed)
	totalSize := util.ByteFmt(task.Meta.Res.Size)
	sb.WriteString(fmt.Sprintf("\r%s [", title))
	for i := 0; i < progressWidth; i++ {
		if i < completeWidth {
			sb.WriteString("■")
		} else {
			sb.WriteString("□")
		}
	}
	sb.WriteString(fmt.Sprintf("] %.1f%%    %s/s    %s", rate*100, speed, totalSize))
	if lastLineLen != 0 {
		paddingLen := lastLineLen - sb.Len()
		if paddingLen > 0 {
			sb.WriteString(strings.Repeat(" ", paddingLen))
		}
	}
	lastLineLen = sb.Len()
	fmt.Print(sb.String())
	sb.Reset()
}
