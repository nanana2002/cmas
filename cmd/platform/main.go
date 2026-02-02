package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"cmas-cats-go/utils"
	"os/exec"
)

const (
	tempCodeDir     = "./temp/code"
	networkSubnet   = "172.18.0.0/16" // cmas-network子网
	basePort        = 5000            // 宿主机端口起始值
	networkName     = "cmas-network"
)

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// 跨域配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 初始化临时目录
	os.MkdirAll(tempCodeDir, 0777)

	// 1. 托管前端页面
	r.StaticFile("/provider", "./web/provider/index.html")
	r.StaticFile("/user", "./web/user/index.html")
	r.StaticFile("/show", "./web/show/show.html")

	// Serve static files for images
	// r.Static("/fig", "./fig")
	r.Static("/fig", "/home/daiyina/cmas-cats-go/fig")

	// 调试日志：打印静态文件路径
	fmt.Println("Serving static files from /fig at /home/daiyina/cmas-cats-go/fig")

	// 调试日志：打印所有路由
	for _, route := range r.Routes() {
		fmt.Printf("Registered route: %s %s\n", route.Method, route.Path)
	}

	// 2. 获取可用参数接口（核心：自动检测未使用的ID/IP/端口）
	r.GET("/api/get-available-params", func(c *gin.Context) {
		// 检测可用ServiceID（S1→S2→S3→...）
		availableServiceID, err := getAvailableServiceID()
		if err != nil {
			c.JSON(500, gin.H{"error": "检测可用ServiceID失败：" + err.Error()})
			return
		}

		// 检测可用容器IP（172.18.0.2→172.18.0.3→...）
		availableIP, err := getAvailableContainerIP()
		if err != nil {
			c.JSON(500, gin.H{"error": "检测可用容器IP失败：" + err.Error()})
			return
		}

		// 检测可用宿主机端口（5000→5001→...）
		availablePort, err := getAvailableHostPort()
		if err != nil {
			c.JSON(500, gin.H{"error": "检测可用宿主机端口失败：" + err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"serviceID":   availableServiceID,
			"containerIP": availableIP,
			"hostPort":    strconv.Itoa(availablePort),
		})
	})

	// 3. 代码上传接口
	r.POST("/api/upload/code", func(c *gin.Context) {
		file, err := c.FormFile("codeFile")
		if err != nil {
			c.JSON(400, gin.H{"error": "上传文件失败：" + err.Error()})
			return
		}

		fileName := filepath.Base(file.Filename)
		savePath := filepath.Join(tempCodeDir, fileName)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(500, gin.H{"error": "保存文件失败：" + err.Error()})
			return
		}

		// 调试日志：打印文件名和解压路径
		fmt.Printf("上传的文件名: %s\n", fileName)
		fmt.Printf("保存路径: %s\n", savePath)

		// 解压到 temp 文件夹
		unzipPath := "temp/unzipped/"
		os.MkdirAll(unzipPath, os.ModePerm)
		cmd := exec.Command("unzip", "-o", savePath, "-d", unzipPath)
		output, err := cmd.CombinedOutput()
		fmt.Printf("解压命令输出: %s\n", string(output))
		if err != nil {
			fmt.Printf("解压文件失败: %s, 输出: %s\n", err.Error(), string(output))
			c.JSON(500, gin.H{"error": "解压文件失败: " + err.Error()})
			return
		}

		// 检查 requirements.txt 是否存在
		requirementsPath := filepath.Join(unzipPath, "requirements.txt")
		if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
			fmt.Printf("未找到 requirements.txt 文件\n")
			c.JSON(400, gin.H{"error": "未找到 requirements.txt 文件"})
			return
		}

		// 调试日志：列出解压后的文件
		files, err := os.ReadDir(unzipPath)
		if err != nil {
			fmt.Printf("读取解压目录失败: %s\n", err.Error())
			c.JSON(500, gin.H{"error": "读取解压目录失败: " + err.Error()})
			return
		}
		fmt.Println("解压后的文件列表:")
		for _, file := range files {
			fmt.Println(file.Name())
		}

		// 安全检查
		securityResult := utils.CheckModelSecurity(unzipPath)
		if !securityResult.Pass {
			os.RemoveAll(unzipPath)
			c.JSON(403, gin.H{
				"error":  "模型安全评估不通过，禁止上传",
				"reason": securityResult.Reason,
				"threats": securityResult.Threats,
			})
			return
		}

		// 自动选择最佳路径
		bestServiceID, err := getAvailableServiceID()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法获取最佳服务路径: " + err.Error()})
			return
		}

		bestPath := filepath.Join("services", strings.ToLower(bestServiceID)+"_service")
		if err := os.MkdirAll(bestPath, 0777); err != nil {
			c.JSON(500, gin.H{"error": "无法创建最佳路径: " + err.Error()})
			return
		}

		finalPath := filepath.Join(bestPath, fileName)
		if err := os.Rename(savePath, finalPath); err != nil {
			c.JSON(500, gin.H{"error": "无法移动文件到最佳路径: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"msg":      "文件已上传到最佳路径",
			"bestPath": finalPath,
		})
	})

	// 4. 创建容器接口
	r.POST("/api/docker/create", func(c *gin.Context) {
		type CreateReq struct {
			ServiceID   string `json:"serviceID"`
			ContainerIP string `json:"containerIP"`
			HostPort    string `json:"hostPort"`
			CodePath    string `json:"codePath"`
			UploadDir   string `json:"uploadDir"`
		}
		var req CreateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数解析失败：" + err.Error()})
			return
		}

		// 参数校验
		if req.ServiceID == "" || req.ContainerIP == "" || req.HostPort == "" {
			c.JSON(400, gin.H{"error": "服务ID、容器IP、宿主机端口不能为空"})
			return
		}

		// 创建上传目录
		if err := os.MkdirAll(req.UploadDir, 0777); err != nil {
			c.JSON(500, gin.H{"error": "创建上传目录失败：" + err.Error()})
			return
		}
		exec.Command("chmod", "777", req.UploadDir).Run()

		// 检查容器是否已存在
		containerName := fmt.Sprintf("cmas-%s", req.ServiceID)
		if isContainerExist(containerName) {
			c.JSON(400, gin.H{"error": fmt.Sprintf("容器%s已存在", containerName)})
			return
		}

		// 执行docker run
		dockerRunCmd := exec.Command("docker", "run", "-d",
			"--name", containerName,
			"--network", networkName,
			"--ip", req.ContainerIP,
			"-p", fmt.Sprintf("%s:5000", req.HostPort),
			"-v", fmt.Sprintf("%s:/app/uploads", req.UploadDir),
			"cmas-service:v1",
		)
		output, err := dockerRunCmd.CombinedOutput()
		if err != nil {
			c.JSON(500, gin.H{
				"error": fmt.Sprintf("创建容器失败：%s，输出：%s", err.Error(), string(output)),
			})
			return
		}

		// 上传代码到容器
		if req.CodePath != "" && fileExists(req.CodePath) {
			dockerCpCmd := exec.Command("docker", "cp", req.CodePath, fmt.Sprintf("%s:/app/s_service.py", containerName))
			cpOutput, err := dockerCpCmd.CombinedOutput()
			if err != nil {
				exec.Command("docker", "rm", "-f", containerName).Run()
				c.JSON(500, gin.H{
					"error": fmt.Sprintf("上传代码失败，已回滚：%s，输出：%s", err.Error(), string(cpOutput)),
				})
				return
			}
		}

		c.JSON(200, gin.H{
			"msg": fmt.Sprintf("容器%s创建成功，ID：%s", containerName, strings.TrimSpace(string(output))),
		})
	})

	// 5. 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"msg":    "Platform服务运行正常",
		})
	})

	// 6. 代码检查接口
	r.POST("/api/check/code", func(c *gin.Context) {
		file, err := c.FormFile("codeFile")
		if err != nil {
			c.JSON(400, gin.H{"error": "上传文件失败: " + err.Error()})
			return
		}

		fileName := filepath.Base(file.Filename)
		savePath := filepath.Join(tempCodeDir, fileName)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(500, gin.H{"error": "保存文件失败: " + err.Error()})
			return
		}

		// 调试日志：打印文件名和解压路径
		fmt.Printf("上传的文件名: %s\n", fileName)
		fmt.Printf("保存路径: %s\n", savePath)

		// 解压到 temp 文件夹
		unzipPath := "temp/unzipped/"
		os.MkdirAll(unzipPath, os.ModePerm)
		cmd := exec.Command("unzip", "-o", savePath, "-d", unzipPath)
		output, err := cmd.CombinedOutput()
		fmt.Printf("解压命令输出: %s\n", string(output))
		if err != nil {
			fmt.Printf("解压文件失败: %s, 输出: %s\n", err.Error(), string(output))
			c.JSON(500, gin.H{"error": "解压文件失败: " + err.Error()})
			return
		}

		// 检查 requirements.txt 是否存在
		requirementsPath := filepath.Join(unzipPath, "requirements.txt")
		if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
			fmt.Printf("未找到 requirements.txt 文件\n")
			c.JSON(400, gin.H{"error": "未找到 requirements.txt 文件"})
			return
		}

		// 调试日志：列出解压后的文件
		files, err := os.ReadDir(unzipPath)
		if err != nil {
			fmt.Printf("读取解压目录失败: %s\n", err.Error())
			c.JSON(500, gin.H{"error": "读取解压目录失败: " + err.Error()})
			return
		}
		fmt.Println("解压后的文件列表:")
		for _, file := range files {
			fmt.Println(file.Name())
		}

		// 安全检查
		securityResult := utils.CheckModelSecurity(unzipPath)
		if !securityResult.Pass {
			os.RemoveAll(unzipPath)
			c.JSON(403, gin.H{
				"error":  "模型安全评估不通过，禁止上传",
				"reason": securityResult.Reason,
				"threats": securityResult.Threats,
			})
			return
		}

		// 自动选择最佳路径
		bestServiceID, err := getAvailableServiceID()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法获取最佳服务路径: " + err.Error()})
			return
		}

		bestPath := filepath.Join("services", strings.ToLower(bestServiceID)+"_service")
		if err := os.MkdirAll(bestPath, 0777); err != nil {
			c.JSON(500, gin.H{"error": "无法创建最佳路径: " + err.Error()})
			return
		}

		finalPath := filepath.Join(bestPath, fileName)
		if err := os.Rename(savePath, finalPath); err != nil {
			c.JSON(500, gin.H{"error": "无法移动文件到最佳路径: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"msg":      "文件已上传到最佳路径",
			"bestPath": finalPath,
		})
	})


	// 启动服务（8081端口）
	fmt.Println("Platform服务启动：0.0.0.0:8081")
	if err := r.Run(":8081"); err != nil {
		fmt.Printf("服务启动失败：%s\n", err.Error())
		os.Exit(1)
	}
}

// ------------------------ 辅助函数：自动检测可用参数 ------------------------
// 检测可用ServiceID（S1、S2、S3...）
func getAvailableServiceID() (string, error) {
	cmd := exec.Command("docker", "ps", "--filter", "name=cmas-", "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	usedIDs := make(map[int]bool)
	lines := strings.Split(string(output), "\n")
	reg := regexp.MustCompile(`cmas-S(\d+)`)

	for _, line := range lines {
		if line == "" {
			continue
		}
		matches := reg.FindStringSubmatch(line)
		if len(matches) == 2 {
			id, _ := strconv.Atoi(matches[1])
			usedIDs[id] = true
		}
	}

	// 从1开始找未使用的ID
	for i := 1; ; i++ {
		if !usedIDs[i] {
			return fmt.Sprintf("S%d", i), nil
		}
	}
}

// 检测可用容器IP（172.18.0.2开始，跳过网关172.18.0.1）
func getAvailableContainerIP() (string, error) {
	// 获取cmas-network中已使用的IP
	cmd := exec.Command("docker", "network", "inspect", networkName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	var networkInfo []map[string]interface{}
	if err := json.Unmarshal(output, &networkInfo); err != nil {
		return "", err
	}

	usedIPs := make(map[string]bool)
	if len(networkInfo) > 0 {
		containers := networkInfo[0]["Containers"].(map[string]interface{})
		for _, container := range containers {
			ip := container.(map[string]interface{})["IPv4Address"].(string)
			ip = strings.Split(ip, "/")[0] // 去掉子网掩码
			usedIPs[ip] = true
		}
	}
	usedIPs["172.18.0.1"] = true // 跳过网关IP

	// 遍历子网找可用IP
	_, subnet, err := net.ParseCIDR(networkSubnet)
	if err != nil {
		return "", err
	}

	ip := net.ParseIP("172.18.0.2")
	for {
		if subnet.Contains(ip) && !usedIPs[ip.String()] {
			return ip.String(), nil
		}
		// IP自增
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] != 0 {
				break
			}
		}
	}
}

// 检测可用宿主机端口（从5000开始）
func getAvailableHostPort() (int, error) {
	// 获取已占用的端口
	cmd := exec.Command("docker", "ps", "--format", "{{.Ports}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	usedPorts := make(map[int]bool)
	lines := strings.Split(string(output), "\n")
	reg := regexp.MustCompile(`(\d+)->5000/tcp`)

	for _, line := range lines {
		if line == "" {
			continue
		}
		matches := reg.FindStringSubmatch(line)
		if len(matches) == 2 {
			port, _ := strconv.Atoi(matches[1])
			usedPorts[port] = true
		}
	}

	// 从basePort开始找未使用的端口
	for port := basePort; ; port++ {
		if !usedPorts[port] {
			// 检查端口是否被系统占用
			conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				return port, nil
			}
			conn.Close()
		}
	}
}

// 辅助函数：检查容器是否存在
func isContainerExist(name string) bool {
	cmd := exec.Command("docker", "inspect", name)
	_, err := cmd.CombinedOutput()
	return err == nil
}

// 辅助函数：检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
