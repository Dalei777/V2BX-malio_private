package panel

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/InazumaV/V2bX/conf"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

// SSPanelNodeData 结构体补充 CustomConfig 字段
type SSPanelNodeData struct {
	NodeGroup      int    `json:"node_group"`
	NodeClass      int    `json:"node_class"`
	NodeSpeedLimit int    `json:"node_speedlimit"`
	TrafficRate    float64`json:"traffic_rate"`
	MuOnly         int    `json:"mu_only"`
	Sort           int    `json:"sort"`
	Server         string `json:"server"`
	Type           string `json:"type"`
	Online         int    `json:"online"`
	// 【新增】接收面板下发的 custom_config
	CustomConfig   string `json:"custom_config"` 
}

// ... (NewSSPanel 和其余结构体保持不变)

// GetNodeInfo 修改部分：在 createVlessNode 前注入 custom_config
func (c *SSPanelClient) GetNodeInfo() (node *NodeInfo, err error) {
    // ... (前文请求代码保持不变，直到 switch sspanelResp.Data.Sort)

    switch sspanelResp.Data.Sort {
    case 15, 16: // VLESS
        logrus.Infof("创建VLESS节点配置")
        // 【关键注入】：如果 custom_config 存在，合并到参数字典中
        if sspanelResp.Data.CustomConfig != "" {
            var extraParams map[string]string
            if err := json.Unmarshal([]byte(sspanelResp.Data.CustomConfig), &extraParams); err == nil {
                for k, v := range extraParams {
                    serverParams[k] = v
                }
                logrus.Infof("已从 custom_config 注入 %d 个新参数", len(extraParams))
            }
        }
        node = c.createVlessNode(node, serverHost, serverParams, sspanelResp.Data.Sort)
    
    // ... (其他 case 保持不变)
    }

    // ... (后续 interval 和 CommonNode 逻辑保持不变)
    return node, nil
}

// createVlessNode 保持不变，它现在会自动使用合并后的 serverParams (包含 custom_config)
// ...
