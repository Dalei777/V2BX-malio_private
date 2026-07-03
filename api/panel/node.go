package panel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// Security type
const (
	None    = 0
	Tls     = 1
	Reality = 2
)

type NodeInfo struct {
	Id           int
	Type         string
	Security     int
	PushInterval time.Duration
	PullInterval time.Duration
	RawDNS       RawDNS
	Rules        Rules

	// origin
	VAllss      *VAllssNode
	Shadowsocks *ShadowsocksNode
	Trojan      *TrojanNode
	Tuic        *TuicNode
	AnyTls      *AnyTlsNode
	Hysteria    *HysteriaNode
	Hysteria2   *Hysteria2Node
	Common      *CommonNode
}

type CommonNode struct {
	Host       string      `json:"host"`
	ServerPort int         `json:"server_port"`
	ServerName string      `json:"server_name"`
	Routes     []Route     `json:"routes"`
	BaseConfig *BaseConfig `json:"base_config"`
}

type Route struct {
	Id          int         `json:"id"`
	Match       interface{} `json:"match"`
	Action      string      `json:"action"`
	ActionValue string      `json:"action_value"`
}

type BaseConfig struct {
	PushInterval any `json:"push_interval"`
	PullInterval any `json:"pull_interval"`
}

// VAllssNode is vmess and vless node info
type VAllssNode struct {
	CommonNode
	Tls                 int             `json:"tls"`
	TlsSettings         TlsSettings     `json:"tls_settings"`
	TlsSettingsBack     *TlsSettings    `json:"tlsSettings"`
	Network             string          `json:"network"`
	NetworkSettings     json.RawMessage `json:"network_settings"`
	NetworkSettingsBack json.RawMessage `json:"networkSettings"`
	ServerName          string          `json:"server_name"`
	// 【新增】：接收面板下发的 custom_config
	CustomConfig json.RawMessage `json:"custom_config"`

	// vless only
	Flow          string        `json:"flow"`
	RealityConfig RealityConfig `json:"-"`
}

type TlsSettings struct {
	ServerName string `json:"server_name"`
	Dest       string `json:"dest"`
	ServerPort string `json:"server_port"`
	ShortId    string `json:"short_id"`
	PrivateKey string `json:"private_key"`
	Xver       uint64 `json:"xver,string"`
}

type RealityConfig struct {
	Xver         uint64 `json:"Xver"`
	MinClientVer string `json:"MinClientVer"`
	MaxClientVer string `json:"MaxClientVer"`
	MaxTimeDiff  string `json:"MaxTimeDiff"`
}

// ... (其他结构体保持不变：ShadowsocksNode, TrojanNode, TuicNode, AnyTlsNode, HysteriaNode, Hysteria2Node, RawDNS, Rules)

func (c *Client) GetNodeInfo() (node *NodeInfo, err error) {
	if c.sspanelClient != nil {
		return c.sspanelClient.GetNodeInfo()
	}

	const path = "/mod_mu/nodes/" // 请根据你的实际路由确认路径
	// 注意：如果你之前的路由是 /api/v1/server/UniProxy/config，请保持原样
	// 如果你改用了 /mod_mu/nodes/{id}/info，请确保这里的 path 匹配
	r, err := c.client.R().SetHeader("If-None-Match", c.nodeEtag).ForceContentType("application/json").Get(path + strconv.Itoa(c.NodeId) + "/info")

	if r.StatusCode() == 304 {
		return nil, nil
	}
	// ... (哈希校验逻辑保持不变)

	node = &NodeInfo{
		Id:     c.NodeId,
		Type:   c.NodeType,
		RawDNS: RawDNS{DNSMap: make(map[string]map[string]interface{}), DNSJson: []byte("")},
	}

	var cm *CommonNode
	switch c.NodeType {
	case "vmess", "vless":
		rsp := &VAllssNode{}
		err = json.Unmarshal(r.Body(), rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		
		// 桥接逻辑：解析 CustomConfig 到 TlsSettings
		if len(rsp.CustomConfig) > 0 {
			var tempTls TlsSettings
			if err := json.Unmarshal(rsp.CustomConfig, &tempTls); err == nil {
				rsp.TlsSettings = tempTls
			}
		}

		if len(rsp.NetworkSettingsBack) > 0 {
			rsp.NetworkSettings = rsp.NetworkSettingsBack
		}
		if rsp.TlsSettingsBack != nil {
			rsp.TlsSettings = *rsp.TlsSettingsBack
		}
		cm = &rsp.CommonNode
		node.VAllss = rsp
		node.Security = node.VAllss.Tls
		
	// ... (其他 case 保持不变)
	}

	// ... (剩余的 Rules 解析、interval 计算及返回值逻辑保持不变)
	return node, nil
}
