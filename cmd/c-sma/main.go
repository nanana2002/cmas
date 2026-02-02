package main

import (
	"cmas-cats-go/models"
	"cmas-cats-go/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ServiceMetric 服务指标结构体
type ServiceMetric struct {
	ServiceID string  `json:"serviceID"`
	IP        string  `json:"ip"`
	Port      string  `json:"port"`
	Gas       float64 `json:"gas"`
	Cost      float64 `json:"cost"`
	Delay     float64 `json:"delay"`
}

var metricMap = make(map[string][]models.ServiceInstanceInfo) // 存储所有服务指标

func main() {
	// 定时扫描cmas容器（每5秒一次）
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			scanCmasContainers()
		}
	}()

	// 提供指标查询接口（给Provider前端调用）
	http.HandleFunc("/api/metrics/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 转换metricMap为JSON返回
		response := make(map[string][]models.ServiceInstanceInfo)
		for serviceID, instances := range metricMap {
			response[serviceID] = instances
		}

		json.NewEncoder(w).Encode(response)
	})

	// 提供同步指标接口（给CPS调用）
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 返回指标数据给CPS
		response := map[string]interface{}{
			"success": true,
			"data":    metricMap,
			"msg":     "指标同步成功",
		}

		json.NewEncoder(w).Encode(response)
	})

	// 启动CSMA服务（8083端口）
	fmt.Println("CSMA服务启动：0.0.0.0:8083")
	http.ListenAndServe(":8083", nil)
}

// scanCmasContainers 扫描cmas容器，拉取指标
func scanCmasContainers() {
	// 1. 调用工具函数获取cmas容器列表
	containerOutput, err := utils.ListCmasContainers()
	if err != nil {
		fmt.Printf("扫描容器失败: %v\n", err)
		return
	}
	lines := strings.Split(containerOutput, "\n")

	// 临时存储新扫描到的指标
	newMetrics := make(map[string][]models.ServiceInstanceInfo)

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}
		containerName := fields[0]
		containerIP := fields[4]
		serviceID := strings.TrimPrefix(containerName, "cmas-")
		hostPort := strings.Split(fields[2], ":")[0] // 提取宿主机端口

		// 2. 拉取该容器的/metrics接口（用宿主机IP+端口）
		metricURL := fmt.Sprintf("http://%s:%s/metrics", "192.168.235.48", hostPort) // 替换为你的服务器IP
		resp, err := http.Get(metricURL)
		if err != nil {
			fmt.Printf("拉取%s指标失败: %v\n", serviceID, err)
			continue
		}
		defer resp.Body.Close()

		// 3. 解析指标
		var containerMetrics map[string]*models.ServiceInstanceInfo
		if err := json.NewDecoder(resp.Body).Decode(&containerMetrics); err != nil {
			fmt.Printf("解析%s指标失败: %v\n", serviceID, err)
			continue
		}

		// 4. 更新到newMetrics
		for sid, metric := range containerMetrics {
			// 更新CSCIID为容器内部IP:端口格式
			metric.CSCIID = fmt.Sprintf("%s:%s", containerIP, "5000") // 使用容器内部端口
			newMetrics[sid] = append(newMetrics[sid], *metric)
			fmt.Printf("更新%s指标成功: %+v\n", sid, metric)
		}
	}

	// 5. 将新指标更新到全局metricMap
	metricMap = newMetrics
}