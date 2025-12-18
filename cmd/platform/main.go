package main

import (
    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "sync"
)

var (
    serviceMap sync.Map // 存储注册的服务（ServiceID -> models.Service）
    sampleMap  sync.Map // 存储验证样本（ServiceID -> 预期结果）
)

func main() {
    r := gin.Default()
    // 1. 服务注册接口（草案Figure 1）
    r.POST("/api/v1/services", func(c *gin.Context) {
        var s models.Service
        if err := c.ShouldBindJSON(&s); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
            return
        }
        serviceMap.Store(s.ID, s)
        sampleMap.Store(s.ID, s.ValidationResult)
        c.JSON(http.StatusOK, gin.H{"success": true, "service_id": s.ID})
    })

    // 2. 服务查询接口
    r.GET("/api/v1/services/:id", func(c *gin.Context) {
        id := c.Param("id")
        s, ok := serviceMap.Load(id)
        if !ok {
            c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "服务不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"success": true, "data": s})
    })

    // 3. 部署验证接口（草案Figure 2）
    r.POST("/api/v1/validate", func(c *gin.Context) {
        var req struct {
            ServiceID string `json:"service_id"`
            Result    string `json:"result"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
            return
        }
        expected, ok := sampleMap.Load(req.ServiceID)
        if !ok {
            c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "服务不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"success": req.Result == expected.(string)})
    })

    // 启动服务（使用配置中的端口）
    r.Run(fmt.Sprintf(":%d", config.Cfg.Platform.Port))
}
