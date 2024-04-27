package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"one-api/common"
	"one-api/controller"
	"one-api/middleware"
	"one-api/model"
	"one-api/router"
	"one-api/service"
	"os"
	"strconv"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	_ "net/http/pprof"
)

//go:embed web/dist
var buildFS embed.FS

//go:embed web/dist/index.html
var indexPage []byte

func main() {
	// 首先执行一次 systemCode()
	if err := systemCode(); err != nil {
		fmt.Println("失败: 程序退出", err)
		os.Exit(1)
	}
	// 启动定时器，每5秒执行一次 systemCode()
	go func() {
		// 使用 time.Tick() 创建一个每隔 时间 发送一次时间的通道
		// ticker := time.Tick(5 * time.Second)
		ticker := time.Tick(24 * time.Hour)

		// 无限循环，不断接收来自 ticker 通道的时间事件
		for {
			<-ticker // 每秒钟触发一次
			if err := systemCode(); err != nil {
				fmt.Println("失败: 程序退出", err)
				os.Exit(1)

			}
		}
	}()

	// 主 goroutine 继续执行其他任务或者等待
	// select {}
	common.SetupLogger()
	common.SysLog("New API " + common.Version + " started")
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	if common.DebugEnabled {
		common.SysLog("running in debug mode")
	}
	// Initialize SQL Database
	err := model.InitDB()
	if err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.FatalLog("failed to close database: " + err.Error())
		}
	}()

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		common.FatalLog("failed to initialize Redis: " + err.Error())
	}

	// Initialize options
	model.InitOptionMap()
	if common.RedisEnabled {
		// for compatibility with old versions
		common.MemoryCacheEnabled = true
	}
	if common.MemoryCacheEnabled {
		common.SysLog("memory cache enabled")
		common.SysError(fmt.Sprintf("sync frequency: %d seconds", common.SyncFrequency))
		model.InitChannelCache()
	}
	if common.RedisEnabled {
		go model.SyncTokenCache(common.SyncFrequency)
	}
	if common.MemoryCacheEnabled {
		go model.SyncOptions(common.SyncFrequency)
		go model.SyncChannelCache(common.SyncFrequency)
	}

	// 数据看板
	go model.UpdateQuotaData()

	if os.Getenv("CHANNEL_UPDATE_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_UPDATE_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_UPDATE_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyUpdateChannels(frequency)
	}
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_TEST_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyTestChannels(frequency)
	}
	common.SafeGoroutine(func() {
		controller.UpdateMidjourneyTaskBulk()
	})
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		common.BatchUpdateEnabled = true
		common.SysLog("batch update enabled with interval " + strconv.Itoa(common.BatchUpdateInterval) + "s")
		model.InitBatchUpdater()
	}

	if os.Getenv("ENABLE_PPROF") == "true" {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:8005", nil))
		}()
		go common.Monitor()
		common.SysLog("pprof enabled")
	}

	service.InitTokenEncoders()

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		common.SysError(fmt.Sprintf("panic detected: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Panic detected, error: %v. Please submit a issue here: https://github.com/Calcium-Ion/new-api", err),
				"type":    "new_api_panic",
			},
		})
	}))
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)
	// Initialize session store
	store := cookie.NewStore([]byte(common.SessionSecret))
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS, indexPage)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	err = server.Run(":" + port)
	if err != nil {
		common.FatalLog("failed to start HTTP server: " + err.Error())
	}
}

func systemCode() error {
	// 机器码
	id, err := machineid.ID()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	// 授权码
	var AUTHORIZATION = os.Getenv("AUTHORIZATION")

	// 目标 URL
	url := "http://38.207.165.63:8600/authorize"

	// 准备请求体参数
	body := []byte(`{"model": "new-api"}`)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return err
	}

	// 添加自定义请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", AUTHORIZATION)
	// req.Header.Set("Authorization", "sk-LhVEhsiAJASgEs0wBc4e05F9E7654253BcFa2e6d9a194198")
	req.Header.Set("systemCode", id)

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发送请求失败:", err)
		return err
	}
	defer resp.Body.Close()

	// 处理响应
	fmt.Println("响应状态码:", resp.Status)
	if resp.Status != "200 OK" {
		// 读取响应体
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("读取响应体失败:", err)
			return err
		}
		fmt.Println("响应体:", string(body))
		return fmt.Errorf("非200响应状态码: %s", resp.Status) // 如果响应状态码不是200，返回自定义错误
	}
	// 读取响应体
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("读取响应体失败:", err)
	// 	return
	// }
	// fmt.Println("响应体:", string(body))
	return nil
}
