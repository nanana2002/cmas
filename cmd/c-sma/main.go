package main

import (
    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "encoding/json"
    "fmt" 
    "github.com/gin-gonic/gin"
    "net/http"
    "sync"
    "time"
)

var metricsMap sync.Map // 聚合指标：ServiceID → []models.ServiceInstanceInfo

// 定时拉取所有Site的metrics
func fetchMetrics() {
    ticker := time.NewTicker(10 * time.Second) // 每10秒拉取一次
    defer ticker.Stop()

    for range ticker.C {
        var wg sync.WaitGroup
        // 遍历所有Docker Site
    for _, site := range config.Cfg.DockerSites {
        wg.Add(1)
        go func(siteCfg struct {
            ServiceID string
            MetricsURL string
            CSCIID    string
        }) {
            defer wg.Done()
            // 手动打印拉取地址，便于调试（新增）
            fmt.Printf("拉取指标：%s\n", siteCfg.MetricsURL)
            
            // 拉取metrics接口（新增超时设置，避免卡壳）
            client := &http.Client{Timeout: 5 * time.Second}
            resp, err := client.Get(siteCfg.MetricsURL)
            if err != nil {
                fmt.Printf("拉取失败：%s，错误：%v\n", siteCfg.MetricsURL, err)
                return
            }
            defer resp.Body.Close()

            // 解析指标数据
            var metric models.ServiceInstanceInfo
            if err := json.NewDecoder(resp.Body).Decode(&metric); err != nil {
                fmt.Printf("解析失败：%s，错误：%v\n", siteCfg.MetricsURL, err)
                return
            }

            // 手动补充CSCIID（避免Site返回的csci_id为空）
            metric.CSCIID = siteCfg.CSCIID

            // 聚合到metricsMap
            if list, ok := metricsMap.Load(metric.ServiceID); ok {
                list = append(list.([]models.ServiceInstanceInfo), metric)
                metricsMap.Store(metric.ServiceID, list)
            } else {
                metricsMap.Store(metric.ServiceID, []models.ServiceInstanceInfo{metric})
            }
        }(site)
    }
        wg.Wait()
    }
}

func main() {
    go fetchMetrics() // 启动定时拉取

    r := gin.Default()
    // 暴露sync接口，供C-PS获取聚合指标
    r.GET("/sync", func(c *gin.Context) {
        result := make(map[string][]models.ServiceInstanceInfo)
        metricsMap.Range(func(key, value interface{}) bool {
            result[key.(string)] = value.([]models.ServiceInstanceInfo)
            return true
        })
        c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
    })

    // 启动C-SMA服务（8083端口）
    r.Run(":8083")
}
