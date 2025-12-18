package main

import (
    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "net/http"
    "sort"
)

// 模拟C-NMA获取网络延迟（实际可根据IP ping计算）
func getNetworkDelay(csciID string) int {
    return 5 // 固定延迟，仅模拟
}

// 从C-SMA获取聚合指标
func getCmaMetrics() (map[string][]models.ServiceInstanceInfo, error) {
    resp, err := http.Get(config.Cfg.CSMA.URL + "/sync")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var cmaResp struct {
        Success bool                                   `json:"success"`
        Data    map[string][]models.ServiceInstanceInfo `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&cmaResp); err != nil {
        return nil, err
    }
    return cmaResp.Data, nil
}

func main() {
    r := gin.Default()

    // 路径选择接口（供Client调用）
    r.POST("/select", func(c *gin.Context) {
        var req models.ClientRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
            return
        }

        // 1. 获取C-SMA的聚合指标
        metrics, err := getCmaMetrics()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "获取指标失败：" + err.Error()})
            return
        }

        // 2. 筛选符合条件的Site（成本≤最大可接受、有可用实例）
        var candidates []models.ServiceInstanceInfo
        for _, instance := range metrics[req.ServiceID] {
            if instance.Cost <= req.MaxAcceptCost && instance.Gas > 0 {
                // 补充网络延迟（计算总延迟）
                instance.Delay += getNetworkDelay(instance.CSCIID)
                candidates = append(candidates, instance)
            }
        }

        // 3. 无可用实例
        if len(candidates) == 0 {
            c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "无符合条件的服务实例"})
            return
        }

        // 4. 排序：成本优先，延迟为辅（草案决策逻辑）
        sort.Slice(candidates, func(i, j int) bool {
            if candidates[i].Cost == candidates[j].Cost {
                return candidates[i].Delay < candidates[j].Delay
            }
            return candidates[i].Cost < candidates[j].Cost
        })

        // 5. 返回最优Site的CSCI-ID
        best := candidates[0]
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data": gin.H{
                "service_id":  req.ServiceID,
                "csci_id":     best.CSCIID,
                "real_cost":   best.Cost,
                "real_delay":  best.Delay,
                "service_name": "最优服务：" + req.ServiceID,
            },
        })
    })

    // 启动C-PS服务（8084端口）
    r.Run(":8084")
}
