package main

import (
	"flag"
	"fmt"
	"github.com/oneclickvirt/ecs/backtrace"
	"github.com/oneclickvirt/ecs/commediatest"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"github.com/oneclickvirt/ecs/ntrace"
	"github.com/oneclickvirt/ecs/speedtest"
	"github.com/oneclickvirt/ecs/unlocktest"
	"github.com/oneclickvirt/ecs/utils"
	"github.com/oneclickvirt/portchecker/email"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"
)

var (
	ecsVersion                       = "2024.06.30"
	showVersion                      bool
	language                         string
	cpuTestMethod, cpuTestThreadMode string
	memoryTestMethod                 string
	diskTestMethod, diskTestPath     string
	diskMultiCheck                   bool
	nt3CheckType, nt3Location        string
	spNum                            int
	width                            = 84
)

func main() {
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.StringVar(&language, "l", "zh", "Specify language (supported: en, zh)")
	flag.StringVar(&cpuTestMethod, "cpum", "sysbench", "Specify CPU test method (supported: sysbench, geekbench, winsat)")
	flag.StringVar(&cpuTestThreadMode, "cput", "multi", "Specify CPU test thread mode (supported: single multi)")
	flag.StringVar(&memoryTestMethod, "memorym", "dd", "Specify Memory test method (supported: sysbench, dd, winsat)")
	flag.StringVar(&diskTestMethod, "diskm", "fio", "Specify Disk test method (supported: fio, dd, winsat)")
	flag.StringVar(&diskTestPath, "diskp", "", "Specify Disk test path, example: -diskp /root")
	flag.BoolVar(&diskMultiCheck, "diskmc", false, "Enable multiple disk checks, example: -diskmc=false")
	flag.StringVar(&nt3Location, "nt3loc", "GZ", "指定三网回程路由检测的地址，支持 GZ, SH, BJ, CD 对应 广州，上海，北京，成都")
	flag.StringVar(&nt3CheckType, "nt3t", "ipv4", "指定三网回程路由检测的类型，支持 both, ipv4, ipv6")
	flag.IntVar(&spNum, "spnum", 2, "Specify speedtest each operator servers num")
	flag.Parse()
	if showVersion {
		fmt.Println(ecsVersion)
		return
	}
	startTime := time.Now()
	var (
		wg                                            sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo string
		output, tempOutput                            string
	)
	output = utils.PrintAndCapture(func() {
		switch language {
		case "zh":
			utils.PrintHead(language, width, ecsVersion)
			utils.PrintCenteredTitle("基础信息", width)
			basicInfo, securityInfo, nt3CheckType = utils.SecurityCheck(language, nt3CheckType)
			fmt.Printf(basicInfo)
			utils.PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", cpuTestMethod), width)
			cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
			utils.PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", cpuTestMethod), width)
			memorytest.MemoryTest(language, memoryTestMethod)
			utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", diskTestMethod), width)
			disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
			utils.PrintCenteredTitle("御三家流媒体解锁", width)
			wg.Add(2)
			go func() {
				defer wg.Done()
				emailInfo = email.EmailCheck()
			}()
			go func() {
				defer wg.Done()
				mediaInfo = unlocktest.MediaTest(language)
			}()
			commediatest.ComMediaTest(language)
			utils.PrintCenteredTitle("跨国流媒体解锁", width)
			wg.Wait() // 后台任务含流媒体测试和邮件测试
			fmt.Printf(mediaInfo)
			utils.PrintCenteredTitle("IP质量检测", width)
			fmt.Printf(securityInfo)
			utils.PrintCenteredTitle("邮件端口检测", width)
			fmt.Println(emailInfo)
			if runtime.GOOS != "windows" {
				utils.PrintCenteredTitle("三网回程", width)
				backtrace.BackTrace()
				// nexttrace 在win上不支持检测，报错 bind: An invalid argument was supplied.
				utils.PrintCenteredTitle("路由检测", width)
				ntrace.TraceRoute3(language, nt3Location, nt3CheckType)
			}
			utils.PrintCenteredTitle("就近节点测速", width)
			speedtest.ShowHead(language)
			speedtest.NearbySP()
			speedtest.CustomSP("net", "global", 4)
			speedtest.CustomSP("net", "cu", spNum)
			speedtest.CustomSP("net", "ct", spNum)
			speedtest.CustomSP("net", "cmcc", spNum)
			utils.PrintCenteredTitle("", width)
			endTime := time.Now()
			duration := endTime.Sub(startTime)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
			fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
			fmt.Printf("时间          : %s\n", currentTime)
			utils.PrintCenteredTitle("", width)
		case "en":
			utils.PrintHead(language, width, ecsVersion)
			utils.PrintCenteredTitle("Basic Information", width)
			basicInfo, securityInfo, nt3CheckType = utils.SecurityCheck(language, nt3CheckType)
			fmt.Printf(basicInfo)
			utils.PrintCenteredTitle(fmt.Sprintf("CPU Test - %s Method", cpuTestMethod), width)
			cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
			utils.PrintCenteredTitle(fmt.Sprintf("Memory Test - %s Method", memoryTestMethod), width)
			utils.PrintCenteredTitle(fmt.Sprintf("Disk Test - %s Method", diskTestMethod), width)
			utils.PrintCenteredTitle(fmt.Sprintf("Disk Test - %s Method", diskTestMethod), width)
			disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
			wg.Add(1)
			go func() {
				defer wg.Done()
				emailInfo = email.EmailCheck()
			}()
			utils.PrintCenteredTitle("The Three Families Streaming Media Unlock", width)
			commediatest.ComMediaTest(language)
			utils.PrintCenteredTitle("Cross-Border Streaming Media Unlock", width)
			unlocktest.MediaTest(language)
			utils.PrintCenteredTitle("IP Quality Check", width)
			fmt.Printf(securityInfo)
			utils.PrintCenteredTitle("Email Port Check", width)
			wg.Wait()
			fmt.Println(emailInfo)
			//utils.PrintCenteredTitle("Return Path Routing", width)
			utils.PrintCenteredTitle("Nearby Node Speed Test", width)
			speedtest.ShowHead(language)
			speedtest.NearbySP()
			speedtest.CustomSP("net", "global", -1)
			utils.PrintCenteredTitle("", width)
			endTime := time.Now()
			duration := endTime.Sub(startTime)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
			fmt.Printf("Cost    Time          : %d 分 %d 秒\n", minutes, seconds)
			fmt.Printf("Current Time          : %s\n", currentTime)
			utils.PrintCenteredTitle("", width)
		default:
			fmt.Println("Unsupported language")
		}
	}, tempOutput, output)
	// 创建文件
	filePath := "goecs.txt"
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Can not make file:", err)
		return
	}
	defer file.Close()
	// 将 output 写入文件
	// 匹配 ANSI 转义序列
	ansiRegex := regexp.MustCompile("\x1B\\[[0-9;]+[a-zA-Z]")
	// 移除 ANSI 转义序列
	cleanedOutput := ansiRegex.ReplaceAllString(output, "")
	_, err = file.WriteString(cleanedOutput)
	if err != nil {
		fmt.Println("Can not write file:", err)
		return
	}
	// 获取文件的绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Println("Get file absPath failed:", err)
		return
	}
	shorturl, err := utils.UploadText(absPath)
	if err != nil {
		fmt.Println("Upload failed, can not generate short URL.")
		fmt.Println(err.Error())
	}
	fmt.Println("Upload successful, short URL:", shorturl)
}
