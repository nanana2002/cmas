package models

// Service 对应草案Table 1（公共服务平台的服务表）
type Service struct {
    ID                 string   `json:"id"`                  // 服务ID（S1/S2/S3）
    Name               string   `json:"name"`                // 服务名称（AR/VR等）
    InputFormat        string   `json:"input_format"`        // 输入格式
    ComputingRequirement string `json:"computing_requirement"`// 计算要求
    StorageRequirement string   `json:"storage_requirement"` // 存储要求
    ComputingTime      string   `json:"computing_time"`     // 计算延迟
    CodeLocation       string   `json:"code_location"`       // 代码地址
    SoftwareDependency []string `json:"software_dependency"` // 软件依赖
    ValidationSample   string   `json:"-"`                  // 验证样本（私有）
    ValidationResult   string   `json:"-"`                  // 预期结果（私有）
}

// ServiceInstanceInfo 对应草案Table 3（Site的服务模型表）
type ServiceInstanceInfo struct {
    ServiceID string `json:"service_id"` // 服务ID
    Gas       int    `json:"gas"`        // 可用实例数
    Cost      int    `json:"cost"`       // 成本
    CSCIID    string `json:"csci_id"`    // 访问地址
    Delay     int    `json:"delay"`      // 延迟（ms）
}

// ClientRequest 客户端请求结构（草案Section 8）
type ClientRequest struct {
    ServiceID     string `json:"service_id"`     // 目标服务ID
    MaxAcceptCost int    `json:"max_accept_cost"`// 最大可接受成本
    MaxAcceptDelay int   `json:"max_accept_delay"`// 最大可接受延迟（ms）
}
