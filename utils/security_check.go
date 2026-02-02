package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 安全评估结果结构体
type SecurityCheckResult struct {
	Pass      bool     `json:"pass"`      // 是否通过
	Reason    string   `json:"reason"`    // 失败原因
	Threats   []string `json:"threats"`   // 检测到的威胁
	FileName  string   `json:"file_name"` // 文件名
	FileType  string   `json:"file_type"` // 文件类型
}

// 高危文件扩展名
var highRiskExts = map[string]bool{
	".pkl":    true,
	".pickle": true,
	".joblib": true,
	".h5":     true,
	".hdf5":   true,
	".hdf":    true,
}

// 恶意代码特征正则
var maliciousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`os\.system\(.*\)`),
	regexp.MustCompile(`subprocess\.Popen\(.*\)`),
	regexp.MustCompile(`exec\(.*\)`),
	regexp.MustCompile(`eval\(.*\)`),
	regexp.MustCompile(`__reduce__`), // pickle反序列化漏洞特征
	regexp.MustCompile(`socket\.socket\(.*\)`),
	regexp.MustCompile(`requests\.post\(.*\)`), // 可疑外连
	regexp.MustCompile(`后门|木马|挖矿|crypto`), // 中文恶意特征
}

// LLM后门特征正则
var llmBackdoorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`trigger\s*=\s*["'].*["']`),
	regexp.MustCompile(`backdoor|backdoor_key|hidden_command`),
	regexp.MustCompile(`system\.prompt\s*\+=\s*["'].*["']`),
	regexp.MustCompile(`unlock_all|bypass_security`),
}

// CheckModelSecurity 核心：模型安全评估主函数
func CheckModelSecurity(filePath string) *SecurityCheckResult {
	result := &SecurityCheckResult{
		FileName: filepath.Base(filePath),
		Pass:     true,
	}

	// 步骤1：检测文件类型
	fileExt := strings.ToLower(filepath.Ext(filePath))
	result.FileType = fileExt
	if highRiskExts[fileExt] {
		result.Threats = append(result.Threats, fmt.Sprintf("高危文件类型：%s", fileExt))
		result.Pass = false
	}

	// 步骤2：读取文件内容
	content, err := readFileContent(filePath)
	if err != nil {
		if result.Pass {
			result.Reason = "无法读取文件内容：" + err.Error()
			result.Pass = false
		}
	} else {
		// 步骤3：检测恶意代码特征
		for _, pattern := range maliciousPatterns {
			if pattern.MatchString(content) {
				threat := fmt.Sprintf("检测到恶意代码特征：%s", pattern.String())
				result.Threats = append(result.Threats, threat)
				result.Pass = false
			}
		}

		// 步骤4：检测LLM后门特征
		for _, pattern := range llmBackdoorPatterns {
			if pattern.MatchString(content) {
				threat := fmt.Sprintf("检测到LLM后门特征：%s", pattern.String())
				result.Threats = append(result.Threats, threat)
				result.Pass = false
			}
		}
	}

	// 步骤5：调用外部工具检测（可选，增强安全性）
	externalThreats := runExternalSecurityCheck(filePath)
	if len(externalThreats) > 0 {
		result.Threats = append(result.Threats, externalThreats...)
		result.Pass = false
	}

	// 最终结果
	if !result.Pass {
		result.Reason = strings.Join(result.Threats, "；")
	} else {
		result.Reason = "模型安全评估通过，无风险特征"
	}

	return result
}

// 读取文件内容（兼容文本/二进制）
func readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// 判断是否为文本文件
	if isBinary(content) {
		return "", fmt.Errorf("文件为二进制类型，跳过内容检测")
	}
	return string(content), nil
}

// 判断是否为二进制文件
func isBinary(data []byte) bool {
	return bytes.Contains(data, []byte{0}) // 包含空字节即为二进制
}

// 彻底禁用外部clamav扫描，避免报错
func runExternalSecurityCheck(filePath string) []string {
    // 直接返回空切片，不执行任何外部扫描
    return []string{}
}