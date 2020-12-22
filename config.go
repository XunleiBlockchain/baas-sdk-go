package sdk

type AuthInfo struct {
	ChainID string `json:"chainid"` // 链ID
	ID      string `json:"id"`      // BaaS为开发者分配的ID
	Key     string `json:"key"`     // BaaS为开发者分配的Key
}

type Config struct {
	Keystore       string            // Keystore目录 保存用户账户秘钥
	UnlockAccounts map[string]string // 预解锁账户 从passwd.json中解析得到
	Retry          int               // 请求失败的至多重复次数
	RPCProtocal    string            // BaaS接入层 协议
	XHost          string            // BaaS接入层 Host
	Namespace      string            // 区块链名称空间 tcapi
	ChainID        int64             // 链ID
	GetGasPrice    bool              // 是否从BaaS获取GasPrice
	AuthInfo       AuthInfo          // 与BaaS通信凭证 从auth.json中解析得到
}
