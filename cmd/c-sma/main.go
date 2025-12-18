package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "cmas-cats-go/utils/logger"

    "github.com/gin-gonic/gin"
)

// 全局存储指标数据
var metricsMap sync.Map

func main() {
    // 1. 创建Gin引擎
    r := gin.Default()

    // 2. 跨域中间件（必须在路由前加载）
    r.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
        c.Header("Access-Control-Max-Age", "86400")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // 3. 启动定时拉取指标的goroutine
    go fetchMetricsLoop()

    // 4. 定义/sync接口（供前端获取指标）
    r.GET("/sync", func(c *gin.Context) {
        // 构建返回数据
        result := make(map[string][]models.ServiceInstanceInfo)
        // 遍历metricsMap，转换为返回格式
        metricsMap.Range(func(key, value interface{}) bool {
            serviceID := key.(string)
            instances := value.([]models.ServiceInstanceInfo)
            result[serviceID] = instances
            return true
        })

        // 返回JSON
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    result,
            "msg":     "success",
        })
    })

    // 5. 启动CSMA服务（监听0.0.0.0:8083）
    logger.Info("CSMA", "启动指标收集服务，端口：%d", config.Cfg.CSMA.Port)
    if err := r.Run(fmt.Sprintf(":%d", config.Cfg.CSMA.Port)); err != nil {
        logger.Error("CSMA", "启动失败：%v", err)
    }
}

// fetchMetricsLoop 定时拉取所有Site的指标（10秒/次）
func fetchMetricsLoop() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    // 首次启动立即拉取一次
    fetchAllMetrics()

    for range ticker.C {
        fetchAllMetrics()
    }
}

// fetchAllMetrics 拉取所有Docker Site的指标
func fetchAllMetrics() {
    logger.Info("CSMA", "开始拉取所有Site的指标")
    var wg sync.WaitGroup

    // 遍历配置中的所有Docker Site
    for _, site := range config.Cfg.DockerSites {
        if site.MetricsURL == "" {
            logger.Warn("CSMA", "Site %s的MetricsURL为空，跳过拉取", site.ServiceID)
            continue
        }

        wg.Add(1)
        // 并发拉取每个Site的指标
        go func(site config.DockerSiteConfig) {
            defer wg.Done()
            logger.Debug("CSMA", "拉取指标：%s", site.MetricsURL)

            // 发送HTTP请求拉取指标
            client := &http.Client{Timeout: 5 * time.Second}
            resp, err := client.Get(site.MetricsURL)
            if err != nil {
                logger.Error("CSMA", "拉取失败：%s，错误：%v", site.MetricsURL, err)
                return
            }
            defer resp.Body.Close()

            // 解析返回的JSON数据（使用原有结构体 ServiceInstanceInfo）
            var instance models.ServiceInstanceInfo
            if err := json.NewDecoder(resp.Body).Decode(&instance); err != nil {
                logger.Error("CSMA", "解析失败：%s，错误：%v", site.MetricsURL, err)
                return
            }

            // 补充CSCIID和ServiceID（确保字段完整）
            instance.CSCIID = site.CSCIID
            instance.ServiceID = site.ServiceID

            // 存储到全局map中
            if _, ok := metricsMap.Load(instance.ServiceID); ok {
                // 替换原有数据（去重）
                metricsMap.Store(instance.ServiceID, []models.ServiceInstanceInfo{instance})
            } else {
                // 新增数据
                metricsMap.Store(instance.ServiceID, []models.ServiceInstanceInfo{instance})
            }

            logger.Info("CSMA", "拉取成功：%s，指标：%+v", site.MetricsURL, instance)
        }(site)
    }

    // 等待所有拉取任务完成
    wg.Wait()
    logger.Info("CSMA", "所有Site指标拉取完成")
}