package main

import (
    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "cmas-cats-go/utils/logger"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "sync"
)

var (
    serviceMap sync.Map // 存储已注册的服务
    sampleMap  sync.Map // 存储服务的验证样本
)

func main() {
    // 初始化日志
    logger.Info("PLATFORM", "启动公共服务平台，端口：%d", config.Cfg.Platform.Port)
    
    // 创建Gin引擎
    r := gin.Default()

    // 在main函数中，r := gin.Default() 后添加：
    // 托管服务提供者页面
    r.StaticFile("/provider", "./web/provider/index.html")
    // 托管用户页面
    r.StaticFile("/user", "./web/user/index.html")
    // 允许跨域（解决前端请求后端的跨域问题）
    r.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })
    
    // 1. 服务注册接口（供Client调用）
    r.POST("/api/v1/services", func(c *gin.Context) {
        var s models.Service
        if err := c.ShouldBindJSON(&s); err != nil {
            logger.Error("PLATFORM", "服务注册参数解析失败：%v", err)
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "msg":     err.Error(),
            })
            return
        }

        // 存储服务信息
        serviceMap.Store(s.ID, s)
        sampleMap.Store(s.ID, s.ValidationResult)
        logger.Info("PLATFORM", "服务注册成功，ID：%s", s.ID)

        c.JSON(http.StatusOK, gin.H{
            "success":   true,
            "service_id": s.ID,
            "msg":       "服务注册成功",
        })
    })

    // 2. 服务查询接口（供其他模块调用）
    r.GET("/api/v1/services/:id", func(c *gin.Context) {
        id := c.Param("id")
        s, ok := serviceMap.Load(id)
        if !ok {
            logger.Error("PLATFORM", "服务查询失败，ID：%s（不存在）", id)
            c.JSON(http.StatusNotFound, gin.H{
                "success": false,
                "msg":     "服务不存在",
            })
            return
        }

        logger.Info("PLATFORM", "服务查询成功，ID：%s", id)
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    s,
        })
    })

    // 3. 部署验证接口（供CSMA/CPS调用）
    r.POST("/api/v1/validate", func(c *gin.Context) {
        var req struct {
            ServiceID string `json:"service_id"`
            Result    string `json:"result"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            logger.Error("PLATFORM", "部署验证参数解析失败：%v", err)
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "msg":     err.Error(),
            })
            return
        }

        expected, ok := sampleMap.Load(req.ServiceID)
        if !ok {
            logger.Error("PLATFORM", "部署验证失败，ID：%s（不存在）", req.ServiceID)
            c.JSON(http.StatusNotFound, gin.H{
                "success": false,
                "msg":     "服务不存在",
            })
            return
        }

        success := req.Result == expected.(string)
        logger.Info("PLATFORM", "部署验证完成，ID：%s，结果：%t", req.ServiceID, success)

        c.JSON(http.StatusOK, gin.H{
            "success": success,
        })
    })

    // 启动Platform服务（8080端口）
    if err := r.Run(fmt.Sprintf(":%d", config.Cfg.Platform.Port)); err != nil {
        logger.Error("PLATFORM", "启动失败：%v", err)
    }
}