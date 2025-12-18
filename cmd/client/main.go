package main

import (
    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// 1. 向Platform注册服务（通用函数）
func registerService(service models.Service) (string, error) {
    jsonData, _ := json.Marshal(service)
    resp, err := http.Post(config.Cfg.Platform.URL+"/api/v1/services", "application/json", strings.NewReader(string(jsonData)))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Success   bool   `json:"success"`
        ServiceID string `json:"service_id"`
        Msg       string `json:"msg"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    if !result.Success {
        return "", fmt.Errorf("注册失败：%s", result.Msg)
    }
    return result.ServiceID, nil
}

// 2. 向C-PS请求最优Site（通用函数）
func requestCPS(req models.ClientRequest) (map[string]interface{}, error) {
    jsonData, _ := json.Marshal(req)
    resp, err := http.Post(config.Cfg.CPS.URL+"/select", "application/json", strings.NewReader(string(jsonData)))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Success bool                   `json:"success"`
        Data    map[string]interface{} `json:"data"`
        Msg     string                 `json:"msg"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    if !result.Success {
        return nil, fmt.Errorf("路径选择失败：%s", result.Msg)
    }
    return result.Data, nil
}

// 3. 访问最优Site的/run接口（通用函数）
func callSite(csciID string) (string, error) {
    jsonData, _ := json.Marshal(map[string]string{"input": "测试数据：模拟用户请求"})
    resp, err := http.Post("http://"+csciID+"/run", "application/json", strings.NewReader(string(jsonData)))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Success bool   `json:"success"`
        Result  string `json:"result"`
        Msg     string `json:"msg"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    if !result.Success {
        return "", fmt.Errorf("访问Site失败：%s", result.Msg)
    }
    return result.Result, nil
}

// 4. 测试单个服务的全流程
func testService(service models.Service) {
    fmt.Printf("\n=== 测试服务：%s ===\n", service.ID)
    
    // 步骤1：注册服务
    fmt.Println("1. 注册服务到Platform")
    serviceID, err := registerService(service)
    if err != nil {
        fmt.Println("注册失败：", err)
        return
    }
    fmt.Println("服务注册成功，ID：", serviceID)

    // 步骤2：向C-PS请求最优Site
    fmt.Println("2. 向C-PS请求最优Site")
    cpsReq := models.ClientRequest{
        ServiceID:     service.ID,
        MaxAcceptCost: 10,
        MaxAcceptDelay: 30,
    }
    cpsData, err := requestCPS(cpsReq)
    if err != nil {
        fmt.Println("C-PS请求失败：", err)
        return
    }
    csciID := cpsData["csci_id"].(string)
    fmt.Println("最优Site的CSCI-ID：", csciID)
    fmt.Println("成本：", cpsData["real_cost"])
    fmt.Println("延迟：", cpsData["real_delay"])

    // 步骤3：访问最优Site
    fmt.Println("3. 访问最优Site")
    result, err := callSite(csciID)
    if err != nil {
        fmt.Println("访问失败：", err)
        return
    }
    fmt.Println("服务返回结果：", result)
}

func main() {
    // 定义S1/S2/S3三个测试服务
    services := []models.Service{
        {
            ID:                   "S1",
            Name:                 "AR/VR轻量服务",
            InputFormat:          "Motion Capture",
            ComputingRequirement: "CPU≥2.0GHz",
            StorageRequirement:   "16GB DRAM",
            ComputingTime:        "≤8ms",
            CodeLocation:         "https://github.com/xxx/ar-service",
            SoftwareDependency:   []string{"Unity"},
            ValidationSample:     "test.mp4",
            ValidationResult:     "result.json",
        },
        {
            ID:                   "S2",
            Name:                 "智能交通轻量服务",
            InputFormat:          "GPS Data",
            ComputingRequirement: "CPU≥2.5GHz",
            StorageRequirement:   "8GB DRAM",
            ComputingTime:        "≤12ms",
            CodeLocation:         "https://github.com/xxx/traffic-service",
            SoftwareDependency:   []string{"Python"},
            ValidationSample:     "gps.csv",
            ValidationResult:     "traffic.json",
        },
        {
            ID:                   "S3",
            Name:                 "大模型轻量服务",
            InputFormat:          "Text",
            ComputingRequirement: "CPU≥3.0GHz",
            StorageRequirement:   "32GB DRAM",
            ComputingTime:        "≤15ms",
            CodeLocation:         "https://github.com/xxx/llm-service",
            SoftwareDependency:   []string{"PyTorch"},
            ValidationSample:     "prompt.txt",
            ValidationResult:     "response.json",
        },
    }

    // 批量测试三个服务
    for _, s := range services {
        testService(s)
    }
}