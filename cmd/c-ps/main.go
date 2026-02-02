package main

import (
	"cmas-cats-go/config"
	"cmas-cats-go/models"
	"cmas-cats-go/utils" // 导入一级utils包（已包含Logger功能）
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time" // 必须导入time包
)

// 动态计算网络延迟
func getNetworkDelay(csciID string) int {
	ip := strings.Split(csciID, ":")[0]
	if ip == "" {
		utils.Logger.Warn("CPS", "提取IP失败，使用兜底延迟5ms") // 修正：utils.Logger
		return 5
	}

	cmd := exec.Command("ping", "-c", "3", "-W", "1", ip)
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Logger.Warn("CPS", "ping %s 失败：%v，使用兜底延迟5ms", ip, err) // 修正：utils.Logger
		return 5
	}

	re := regexp.MustCompile(`time=(\d+\.?\d*) ms`)
	matches := re.FindAllStringSubmatch(string(output), -1)
	if len(matches) == 0 {
		utils.Logger.Warn("CPS", "未匹配到%s的延迟值，使用兜底延迟5ms", ip) // 修正：utils.Logger
		return 5
	}

	total := 0.0
	for _, m := range matches {
		delay, _ := strconv.ParseFloat(m[1], 64)
		total += delay
	}
	avgDelay := int(total / float64(len(matches)))
	utils.Logger.Info("CPS", "IP %s 的平均延迟：%dms", ip, avgDelay) // 修正：utils.Logger
	return avgDelay
}

// 检测Site是否可用
func isSiteAvailable(csciID string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + csciID + "/metrics")
	if err != nil || resp.StatusCode != 200 {
		utils.Logger.Warn("CPS", "Site %s 不可用", csciID) // 修正：utils.Logger
		return false
	}
	defer resp.Body.Close()
	utils.Logger.Debug("CPS", "Site %s 可用", csciID) // 修正：utils.Logger
	return true
}

// 从C-SMA获取聚合指标
func getCmaMetrics() (map[string][]models.ServiceInstanceInfo, error) {
	utils.Logger.Debug("CPS", "请求C-SMA聚合指标：%s/sync", config.Cfg.CSMA.URL) // 修正：utils.Logger
	resp, err := http.Get(config.Cfg.CSMA.URL + "/sync")
	if err != nil {
		utils.Logger.Error("CPS", "获取C-SMA指标失败：%v", err) // 修正：utils.Logger
		return nil, err
	}
	defer resp.Body.Close()

	var cmaResp struct {
		Success bool                                   `json:"success"`
		Data    map[string][]models.ServiceInstanceInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cmaResp); err != nil {
		utils.Logger.Error("CPS", "解析C-SMA指标失败：%v", err) // 修正：utils.Logger
		return nil, err
	}
	utils.Logger.Info("CPS", "获取C-SMA指标成功，共%d个服务", len(cmaResp.Data)) // 修正：utils.Logger
	return cmaResp.Data, nil
}

func main() {
	// 初始化日志（确保utils包已实现Logger初始化）
	utils.InitLogger() // 新增：若utils包有初始化函数，需调用

	utils.Logger.Info("CPS", "启动路径选择服务，端口：%d", config.Cfg.CPS.Port) // 修正：utils.Logger
	r := gin.Default()

	// 跨域中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 路径选择接口
	r.POST("/select", func(c *gin.Context) {
		var req models.ClientRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.Logger.Error("CPS", "参数解析失败：%v", err) // 修正：utils.Logger
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
			return
		}
		utils.Logger.Info("CPS", "接收选择请求，服务ID：%s，最大成本：%d，最大延迟：%d", req.ServiceID, req.MaxAcceptCost, req.MaxAcceptDelay) // 修正：utils.Logger

		// 1. 获取C-SMA指标
		metrics, err := getCmaMetrics()
		if err != nil {
			utils.Logger.Error("CPS", "获取指标失败：%v", err) // 修正：utils.Logger
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "获取指标失败：" + err.Error()})
			return
		}

		// 2. 筛选可用且符合条件的Site
		var candidates []models.ServiceInstanceInfo
		for _, instance := range metrics[req.ServiceID] {
			if !isSiteAvailable(instance.CSCIID) {
				continue
			}
			if instance.Cost <= req.MaxAcceptCost && instance.Gas > 0 {
				instance.Delay += getNetworkDelay(instance.CSCIID)
				candidates = append(candidates, instance)
				utils.Logger.Debug("CPS", "候选Site：%+v", instance) // 修正：utils.Logger
			}
		}

		// 3. 无可用实例
		if len(candidates) == 0 {
			utils.Logger.Warn("CPS", "无符合条件的服务实例，服务ID：%s", req.ServiceID) // 修正：utils.Logger
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "无符合条件的服务实例（所有Site均宕机或不满足条件）"})
			return
		}

		// 4. 排序
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].Cost == candidates[j].Cost {
				return candidates[i].Delay < candidates[j].Delay
			}
			return candidates[i].Cost < candidates[j].Cost
		})

		// 5. 返回最优Site
		best := candidates[0]
		utils.Logger.Info("CPS", "最优Site：%s，成本：%d，延迟：%d", best.CSCIID, best.Cost, best.Delay) // 修正：utils.Logger
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"service_id":   req.ServiceID,
				"csci_id":      best.CSCIID,
				"real_cost":    best.Cost,
				"real_delay":   best.Delay,
				"service_name": "最优服务：" + req.ServiceID,
			},
		})
	})

	// 启动服务
	if err := r.Run(fmt.Sprintf(":%d", config.Cfg.CPS.Port)); err != nil {
		utils.Logger.Error("CPS", "启动失败：%v", err) // 修正：utils.Logger
	}
}