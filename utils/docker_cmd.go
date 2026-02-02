package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

// DockerCmd 执行docker命令，返回输出和错误
func DockerCmd(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("docker命令执行失败: %s, 错误: %s", stderr.String(), err.Error())
	}
	return stdout.String(), nil
}

// CreateServiceContainer 创建服务容器（核心：拼接docker run命令）
func CreateServiceContainer(serviceID, containerIP, hostPort, codePath, uploadDir string) (string, error) {
	// 1. 先创建宿主机上传目录（确保存在）
	_, err := ExecCmd("mkdir", "-p", uploadDir)
	if err != nil {
		return "", fmt.Errorf("创建上传目录失败: %v", err)
	}
	// 2. 赋予目录权限
	_, err = ExecCmd("chmod", "777", uploadDir)
	if err != nil {
		return "", fmt.Errorf("设置目录权限失败: %v", err)
	}
	// 3. 拼接docker run命令（固定cmas-前缀，避免冲突）
	containerName := fmt.Sprintf("cmas-%s", serviceID)
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--network", "cmas-network",
		"--ip", containerIP,
		"-p", fmt.Sprintf("%s:5000", hostPort),
		"-v", fmt.Sprintf("%s:/app/uploads", uploadDir),
		"-e", fmt.Sprintf("SERVICE_IP=%s", containerIP),
		"cmas-service:v1", // 复用之前构建的镜像
	}
	// 执行docker run命令
	containerID, err := DockerCmd(args...)
	if err != nil {
		return "", err
	}
	// 4. 传输服务代码到容器（docker cp）
	if codePath != "" {
		_, err = DockerCmd("cp", codePath, fmt.Sprintf("%s:/app/s_service.py", containerName))
		if err != nil {
			// 代码传输失败，回滚删除容器
			DockerCmd("rm", "-f", containerName)
			return "", fmt.Errorf("传输代码失败，已回滚删除容器: %v", err)
		}
	}
	return fmt.Sprintf("容器%s创建成功，ID: %s", containerName, containerID), nil
}

// ExecCmd 执行系统命令（辅助：创建目录/改权限）
func ExecCmd(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("命令执行失败: %s, 错误: %s", stderr.String(), err.Error())
	}
	return stdout.String(), nil
}

// ListCmasContainers 列出所有cmas-前缀的容器（用于前端展示）
func ListCmasContainers() (string, error) {
	// docker ps --filter "name=cmas-" --format "{{.Names}}\t{{.Status}}\t{{.Ports}}"
	args := []string{
		"ps",
		"--filter", "name=cmas-",
		"--format", "{{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Networks}}\t{{.IPAddress}}",
	}
	return DockerCmd(args...)
}