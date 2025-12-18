package main

import (
    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "cmas-cats-go/utils/logger"
    "encoding/json"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "sync"
    "time"
)

var metricsMap sync.Map

// 定时拉取所有Site的metrics
func fetchMetrics() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        logger.Info("CSMA", "开始拉取所有Site的指标")
        var wg sync.WaitGroup
        // 遍历所有Docker Site
        for _, site := range config.Cfg.DockerSites {
            wg.Add(1)
            go func(siteCfg config.DockerSiteConfig) {
                defer wg.Done()
                logger.Debug("CSMA", "拉取指标：%s", siteCfg.MetricsURL)
                
                // 拉取metrics接口
                client := &http.Client{Timeout: 5 * time.Second}
                resp, err := client.Get(siteCfg.MetricsURL)
                if err != nil {
                    logger.Error("CSMA", "拉取失败：%s，错误：%v", siteCfg.MetricsURL, err)
                    return
                }
                defer resp.Body.Close()

                // 解析指标数据
                var metric models.ServiceInstanceInfo
                if err := json.NewDecoder(resp.Body).Decode(&metric); err != nil {
                    logger.Error("CSMA", "解析失败：%s，错误：%v", siteCfg.MetricsURL, err)
                    return
                }

                // 补充CSCIID
                metric.CSCIID = siteCfg.CSCIID

                // 覆盖存储（去重）
                metricsMap.Store(metric.ServiceID, []models.ServiceInstanceInfo{metric})
                logger.Info("CSMA", "拉取成功：%s，指标：%+v", siteCfg.MetricsURL, metric)
            }(site)
        }
        wg.Wait()
        logger.Info("CSMA", "所有Site指标拉取完成")
    }
}

func main() {
    logger.Info("CSMA", "启动指标收集服务，端口：%d", config.Cfg.CSMA.Port)
    go fetchMetrics() // 启动定时拉取

    r := gin.Default()
    // 暴露sync接口，供C-PS获取聚合指标
    r.GET("/sync", func(c *gin.Context) {
        logger.Debug("CSMA", "C-PS请求聚合指标")
        result := make(map[string][]models.ServiceInstanceInfo)
        metricsMap.Range(func(key, value interface{}) bool {
            result[key.(string)] = value.([]models.ServiceInstanceInfo)
            return true
        })
        logger.Info("CSMA", "返回聚合指标，共%d个服务", len(result))
        c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
    })

    // 启动C-SMA服务
    if err := r.Run(fmt.Sprintf(":%d", config.Cfg.CSMA.Port)); err != nil {
        logger.Error("CSMA", "启动失败：%v", err)
    }
}