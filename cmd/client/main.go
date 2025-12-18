package main

import (
    "cmas-cats-go/config"
    "cmas-cats-go/models"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// 1. 向Platform注册服务（初始化）
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

// 2. 向C-PS请求最优Site
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

// 3. 访问最优Site的/run接口
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

func main() {
    // 步骤1：注册测试服务（S1：AR/VR服务）
    fmt.Println("=== 1. 注册服务到Platform ===")
    testService := models.Service{
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
    }
    serviceID, err := registerService(testService)
    if err != nil {
        fmt.Println("注册失败：", err)
        return
    }
    fmt.Println("服务注册成功，ID：", serviceID)

    // 步骤2：向C-PS发起请求（选择最优Site）
    fmt.Println("\n=== 2. 向C-PS请求最优Site ===")
    cpsReq := models.ClientRequest{
        ServiceID:     "S1",
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

    // 步骤3：访问最优Site获取结果
    fmt.Println("\n=== 3. 访问最优Site ===")
    result, err := callSite(csciID)
    if err != nil {
        fmt.Println("访问失败：", err)
        return
    }
    fmt.Println("服务返回结果：", result)
}

