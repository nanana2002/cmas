package config

import "fmt"

// 全局配置实例
var Cfg GlobalConfig

// GlobalConfig 所有模块的配置
type GlobalConfig struct {
    Platform struct {
        IP   string // Platform服务地址（本地启动，默认127.0.0.1）
        Port int    // 端口8080
        URL  string
    }
    CSMA struct {
        IP   string // C-SMA服务地址
        Port int    // 端口8083
        URL  string
    }
    CPS struct {
        IP   string // C-PS服务地址
        Port int    // 端口8084
        URL  string
    }
    // 3个Docker Site的配置（从启动脚本中获取的IP/端口）
    DockerSites []struct {
        ServiceID string // 服务ID：S1/S2/S3
        MetricsURL string // /metrics接口地址（如http://172.17.0.10:5000/metrics）
        CSCIID    string // 服务访问地址（CSCI-ID，如172.17.0.10:5000）
    }
}

// 初始化配置（程序启动时自动加载）
func init() {
    // 本地模块配置
    Cfg.Platform.IP = "127.0.0.1"
    Cfg.Platform.Port = 8080
    Cfg.Platform.URL = fmt.Sprintf("http://%s:%d", Cfg.Platform.IP, Cfg.Platform.Port)

    Cfg.CSMA.IP = "127.0.0.1"
    Cfg.CSMA.Port = 8083
    Cfg.CSMA.URL = fmt.Sprintf("http://%s:%d", Cfg.CSMA.IP, Cfg.CSMA.Port)

    Cfg.CPS.IP = "127.0.0.1"
    Cfg.CPS.Port = 8084
    Cfg.CPS.URL = fmt.Sprintf("http://%s:%d", Cfg.CPS.IP, Cfg.CPS.Port)

    // Docker Site配置（启动后需替换为实际容器IP）
    Cfg.DockerSites = []struct {
        ServiceID string
        MetricsURL string
        CSCIID    string
    }{
        {ServiceID: "S1", MetricsURL: "http://172.17.0.8:5000/metrics", CSCIID: "172.17.0.8:5000"},
        {ServiceID: "S2", MetricsURL: "http://172.17.0.9:5000/metrics", CSCIID: "172.17.0.9:5000"},
        {ServiceID: "S3", MetricsURL: "http://172.17.0.10:5000/metrics", CSCIID: "172.17.0.10:5000"},
    }
}
