package config

import (
	"encoding/json"
	"fmt"
	"os"        // 读取文件必备
	"os/exec"   // 执行docker命令
	"strings"   // 字符串处理
	"time"      // 可选：增加重试延迟
)

// DockerSiteConfig 容器配置结构体
// 对应 config/docker_sites.json 中的每个site配置
type DockerSiteConfig struct {
	ServiceID     string `json:"service_id"`      // 服务ID（S1/S2/S3）
	ContainerName string `json:"container_name"`  // 容器名（cmas-site-1/2/3）
	Port          int    `json:"port"`           // 容器内服务端口（固定5000）
	IP            string // 自动获取的容器内网IP
	MetricsURL    string // 拼接后的metrics接口地址（http://IP:5000/metrics）
	CSCIID        string // 拼接后的CSCI-ID（IP:5000）
}

// GlobalConfig 全局配置结构体
// 包含所有模块的基础配置
type GlobalConfig struct {
	Platform struct {
		IP   string // Platform模块IP
		Port int    // Platform模块端口（8080）
		URL  string // Platform模块完整URL（http://127.0.0.1:8080）
	}
	CSMA struct {
		IP   string // CSMA模块IP
		Port int    // CSMA模块端口（8083）
		URL  string // CSMA模块完整URL（http://127.0.0.1:8083）
	}
	CPS struct {
		IP   string // CPS模块IP
		Port int    // CPS模块端口（8084）
		URL  string // CPS模块完整URL（http://127.0.0.1:8084）
	}
	DockerSites []DockerSiteConfig `json:"sites"` // 所有Docker Site的配置
}

// Cfg 全局配置实例
// 整个项目通过该变量访问配置
var Cfg GlobalConfig

// getContainerIP 安全获取容器IP（兼容所有自定义网络）
func getContainerIP(containerName string) string {
    var ip string
    // 重试3次
    for i := 0; i < 3; i++ {
        // 执行docker inspect，获取所有网络的IP（关键：兼容自定义网络）
        cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Printf("[CONFIG] 第%d次获取容器%s IP失败：%v\n", i+1, containerName, err)
            time.Sleep(1 * time.Second)
            continue
        }
        // 清理输出（去空格/换行）
        ip = strings.TrimSpace(string(output))
        if ip != "" {
            break
        }
    }
    if ip == "" {
        fmt.Printf("[CONFIG] 容器%s IP获取失败\n", containerName)
    } else {
        fmt.Printf("[CONFIG] 容器%s IP获取成功：%s\n", containerName, ip)
    }
    return ip
}

// init 初始化函数（包加载时自动执行）
// 完成所有配置的加载和拼接
func init() {
	// ===================== 1. 基础模块配置 =====================
	// Platform模块配置
	Cfg.Platform.IP = "127.0.0.1"
	Cfg.Platform.Port = 8080
	Cfg.Platform.URL = fmt.Sprintf("http://%s:%d", Cfg.Platform.IP, Cfg.Platform.Port)

	// CSMA模块配置
	Cfg.CSMA.IP = "127.0.0.1"
	Cfg.CSMA.Port = 8083
	Cfg.CSMA.URL = fmt.Sprintf("http://%s:%d", Cfg.CSMA.IP, Cfg.CSMA.Port)

	// CPS模块配置
	Cfg.CPS.IP = "127.0.0.1"
	Cfg.CPS.Port = 8084
	Cfg.CPS.URL = fmt.Sprintf("http://%s:%d", Cfg.CPS.IP, Cfg.CPS.Port)

	// ===================== 2. Docker Site配置 =====================
	// 优先读取配置文件（config/docker_sites.json）
	filePath := "config/docker_sites.json"
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("[CONFIG] 读取配置文件%s失败：%v，使用内置默认配置\n", filePath, err)
		
		// 内置默认配置（无需外部文件，兜底使用）
		defaultContainers := []struct {
			ServiceID     string
			ContainerName string
			Port          int
		}{
			{"S1", "cmas-site-1", 5000},
			{"S2", "cmas-site-2", 5000},
			{"S3", "cmas-site-3", 5000},
		}

		// 构建DockerSites配置（获取IP+拼接地址）
		var dockerSites []DockerSiteConfig
		for _, container := range defaultContainers {
			// 获取容器IP
			ip := getContainerIP(container.ContainerName)
			
			// 初始化Site配置
			site := DockerSiteConfig{
				ServiceID:     container.ServiceID,
				ContainerName: container.ContainerName,
				Port:          container.Port,
				IP:            ip,
			}

			// 仅当IP非空时，拼接MetricsURL和CSCIID
			if ip != "" {
				site.MetricsURL = fmt.Sprintf("http://%s:%d/metrics", ip, container.Port)
				site.CSCIID = fmt.Sprintf("%s:%d", ip, container.Port)
				fmt.Printf("[CONFIG] 容器%s配置完成：MetricsURL=%s, CSCIID=%s\n", 
					container.ContainerName, site.MetricsURL, site.CSCIID)
			} else {
				fmt.Printf("[CONFIG] 容器%sIP为空，跳过地址拼接\n", container.ContainerName)
			}

			dockerSites = append(dockerSites, site)
		}

		// 赋值给全局配置
		Cfg.DockerSites = dockerSites
		return
	}

	// ===================== 3. 解析外部配置文件 =====================
	var dockerConfig struct {
		Sites []DockerSiteConfig `json:"sites"`
	}
	// 解析JSON文件
	if err := json.Unmarshal(file, &dockerConfig); err != nil {
		fmt.Printf("[CONFIG] 解析配置文件%s失败：%v，使用内置默认配置\n", filePath, err)
		// 解析失败时，回退到内置默认配置（同上）
		defaultContainers := []struct {
			ServiceID     string
			ContainerName string
			Port          int
		}{
			{"S1", "cmas-site-1", 5000},
			{"S2", "cmas-site-2", 5000},
			{"S3", "cmas-site-3", 5000},
		}
		var dockerSites []DockerSiteConfig
		for _, container := range defaultContainers {
			ip := getContainerIP(container.ContainerName)
			site := DockerSiteConfig{
				ServiceID:     container.ServiceID,
				ContainerName: container.ContainerName,
				Port:          container.Port,
				IP:            ip,
			}
			if ip != "" {
				site.MetricsURL = fmt.Sprintf("http://%s:%d/metrics", ip, container.Port)
				site.CSCIID = fmt.Sprintf("%s:%d", ip, container.Port)
			}
			dockerSites = append(dockerSites, site)
		}
		Cfg.DockerSites = dockerSites
		return
	}

	// ===================== 4. 补全配置文件中的IP和地址 =====================
	for i := range dockerConfig.Sites {
		site := &dockerConfig.Sites[i]
		// 获取容器IP（覆盖配置文件中的空IP）
		ip := getContainerIP(site.ContainerName)
		if ip == "" {
			fmt.Printf("[CONFIG] 容器%sIP获取失败，跳过配置\n", site.ContainerName)
			continue
		}
		// 更新IP、MetricsURL、CSCIID
		site.IP = ip
		site.MetricsURL = fmt.Sprintf("http://%s:%d/metrics", ip, site.Port)
		site.CSCIID = fmt.Sprintf("%s:%d", ip, site.Port)
		fmt.Printf("[CONFIG] 配置文件中容器%s配置完成：MetricsURL=%s, CSCIID=%s\n", 
			site.ContainerName, site.MetricsURL, site.CSCIID)
	}

	// 赋值给全局配置
	Cfg.DockerSites = dockerConfig.Sites
}