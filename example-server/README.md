# sdk-server使用说明

sdk-server 作为SDK使用示例，以HTTP服务的形式展示SDK所有接口功能。

## Server目录结构介绍

```sh
.
├── main.go              // 程序入口 提供HTTP服务
├── config.go            // HTTP服务和使用到的SDK相关配置
├── server.go            // HTTP服务实现
├── Makefile             // 编译
├── start.sh             // 运行脚本
├── README.md            // 文档
└── test                 // 测试使用的程序配置示例
    ├── auth.json             // SDK连接BaaS接入层的通信凭证。形如
    │                         // {"chainid":"10001023","id":"11",
    │                         // "key":"5d7d481d6a00504de0a32488bc7392c4"}
    ├── keystore              // 用户账户的秘钥文件
    │   └── UTC--2018-04-15T05-21-48.033606105Z--54fb1c7d0f011dd63b08f85ed7b518ab82028100
    ├── passwd.json           // 启动服务时，预解锁的账户和密码对。以账户地址和密码的键值对组成，形如：
    │                         // {"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc":"1234",  
    |                         //  "0xd7e9ff289d2e1405e1bd825168d36ec917aea971":"1234"}
    ├── sdk-server.conf       // 服务配置文件
    └── sdk-server.xml        // 日志配置文件
```

sdk-server.conf 配置文件说明如下：

```conf
# Server配置部分：
http.addr               0.0.0.0:8080        // http服务地址
http.read.timeout       10s                 // http服务读超时时间
http.write.timeout      10s                 // http服务写超时时间

# SDK配置部分：
xhost                   rpc-baas-blockchain.xunlei.com // BaaS接入层 Host
rpc.protocal            https                          // BaaS接入层 http协议
chain_id                30261                          // 链ID
keystore                ./keystore                     // 私钥存储文件
getfee                  false                          // 是否从BaaS获取fee
getgasprice             true                           // 是否从BaaS获取GasPrice
namespace               tcapi                          // 区块链名称空间 tcapi
```

## Server服务启动

启动服务前更新账号秘钥文件 keystore、passwd.json、auth.json 与服务配置文件 sdk-server.conf 。

在当前目录下执行 `make` 以编译Server，随后执行 `start.sh` 启动测试Server。


## Server API说明

Server提供访问SDK所有接口的功能，故其API也与SDK API一致。

### accounts

功能描述：
获取钱包管理的全部账户地址

参数：
none

返回结果：
账户地址数组

示例：
```json
//request
curl -H "Content-Type:application/json" --data '{"jsonrpc":"2.0","method": "accounts", "params": [], "id": 6}' localhost:8080
//result
{
 "id": 6,
 "jsonrpc": "2.0",
 "errcode": 0,
 "errmsg": "success",
 "result": ["0x33d4fcb75ce608920c7e5755304c282141dfc4dc", "0x7a4877494b59c0bd929747800ab86a8b89380ac5", "0x36419474a02a507e236dc473648197f07ab4722e", "0x7fc423bd7ed1f5d17a92bdb8c39ed620f48f7559", "0x8f470d7f2b2db7b83accd008ddabc5423c06044b", "0x622bc0938fae8b028fcf124f9ba8580719009fdc"]
}

```

### newAccount

功能描述：
创建新的账户

参数：
账户密码

返回结果：
新账户地址

示例：
```json
//request
curl -H "Content-Type:application/json" --data '{"jsonrpc":"2.0","method": "newAccount", "params": ["123456"], "id": 6}' localhost:8080
//result
{
 "id": 6,
 "jsonrpc": "2.0",
 "errcode": 0,
 "errmsg": "success",
 "result": "0x84d8698746dbe68c97965c48c7b56979c577df11"
}

```

### getBalance

功能描述：
查询账户余额

参数：
账户地址

返回结果：
账户地址余额

示例：
```json
//request
curl -H "Content-Type:application/json" --data '{"jsonrpc":"2.0","method": "getBalance", "params": ["0x33d4fcb75ce608920c7e5755304c282141dfc4dc"], "id": 6}' localhost:8080
//result
{
 "id": 6,
 "jsonrpc": "2.0",
 "errcode": 0,
 "errmsg": "success",
 "result": 99030093892100000000170
}
```

### getTransactionCount

功能描述：
查询账户nonce值

参数：
账户地址

返回结果：
账户地址当前nonce值

示例：
```json
//request
curl -H "Content-Type:application/json" --data '{"jsonrpc":"2.0","method": "getTransactionCount", "params": ["0x622bc0938fae8b028fcf124f9ba8580719009fdc"], "id": 6}' localhost:8080
//result
{
 "id": 6,
 "errcode": 0,
 "jsonrpc": "2.0",
 "errmsg": "success",
 "result": 26
}
```

### sendTransaction

功能描述：
普通转账交易

参数：
- Object： 交易对象
  - from: 转出账户地址
  - to: 转入账户地址
  - value: 转账金额 (推荐使用十进制)
  - nonce: (可选) from地址nonce值
- from账户密码(可选)

示例：
```json
params: [{
  "from": "0x622bc0938fae8b028fcf124f9ba8580719009fdc",
  "to": "0x33d4fcb75ce608920c7e5755304c282141dfc4dc",
  "value": "0x10", // 16 wei
}，“12345678”]
```

说明：**如果启动服务时，已经解锁了from账户，可以不用再传密码参数**

返回结果：
交易hash

示例：
```json
//request
curl -H "Content-Type:application/json" --data '{"jsonrpc":"2.0","method": "sendTransaction", "params": [{"from": "0x622bc0938fae8b028fcf124f9ba8580719009fdc", "to": "0x33d4fcb75ce608920c7e5755304c282141dfc4dc", "value":"1200"},"12345678"], "id": 6}' localhost:8080
//result
{
 "id": 6,
 "jsonrpc": "2.0",
 "errcode": 0,
 "errmsg": "success",
 "result": "0x517490b857200702453f32ed0574487b44587958ff39b26554df4f4991cae18c"
}

```

### sendContractTransaction
功能描述：

合约执行交易，支持传1个参数，2个参数和3个参数:
- 1个参数则只穿交易对象
- 2个参数则传交易对象和扩展对象
- 3个参数则传交易对象、账户密码、扩展对象

参数：
- Object： 交易对象
  - from: 转出账户地址
  - to: 转入账户地址
  - gas：(可选，默认90000)手续费
  - gasPrice：(可选，默认1e11)
  - value: 转账金额
  - data：执行合约code
  - nonce: (可选) from地址nonce值
- from账户密码(可选)
- Object:  扩展对象（可选）
  - callback:  回调地址
  - prepay_id: 预交易id
  - service_id: 第三方服务号
  - sign：交易签名(签名方式：**MD5(SHA256(callback=XXX&prepay_id=XXX&service_id=XXX&to=XXX&value=XXX&key=XXX)), key为第三方服务号密匙**)
  - tx_type：交易类型描述
  - title：交易主题
  - desc：主题描述

示例：
```json
[{
 "from": "0x622bc0938fae8b028fcf124f9ba8580719009fdc",
 "to": "0x7f7f7dbf351d4272eb282f16091c96b4819007f5",
 "data": "0x49f3870b0000000000000000000000000000000000000000000000000000000000000001"
}, "12345678", {
 "callback": "http://www.baidu.com",
 "prepay_id": "201805171922030000010176643187014087",
 "service_id": "0",
 "sign": "80fa49b1c7e5ec06ab595850dc8e8f87",
 "tx_type": "contract",
 "title": "test",
 "desc": "this is a test by skl"
}]
```

说明：**如果启动服务时，已经解锁了from账户，可以不用再传密码参数**

返回结果：
交易hash

示例：
```json
//request
curl -H "Content-Type:application/json" --data '{"jsonrpc":"2.0","method": "sendContractTransaction", "params": [{"from": "0x622bc0938fae8b028fcf124f9ba8580719009fdc", "to": "0x7f7f7dbf351d4272eb282f16091c96b4819007f5", "data":"0x49f3870b0000000000000000000000000000000000000000000000000000000000000001"},"12345678",{"callback":"http://www.baidu.com","prepay_id":"201805171922030000010176643187014087","service_id":"0","sign":"80fa49b1c7e5ec06ab595850dc8e8f87","tx_type":"contract","title":"test","desc":"this is a test by skl"}], "id": 6}' localhost:8080
//result
{
 "id": 6,
 "jsonrpc": "2.0",
 "errcode": 0,
 "errmsg": "success",
 "result": "0x517490b857200702453f32ed0574487b44587958ff39b26554df4f4991cae18c"
}

```

## 错误说明
- method invalid  接口名有误
- params err    参数有误
- could not decrypt key with given passphrase   账户密码错误

#### 错误码定义

| 错误码 | 错误信息                             | 说明                                      |
| ------ | ------------------------------------ | ----------------------------------------- |
| 0      | success                              | 请求成功                                  |
| -1000  | invalid method                       | 接口请求方法错误                          |
| -1001  | params err                           | 参数错误， 一般由Params参数个数不一致导致 |
| -1002  | find account err                     | 账号查找错误                              |
| -1003  | rpc getBalance err                   | 查看余额 rpc调用失败                      |
| -1004  | ks.NewAccount err                    | 新建账号错误                              |
| -1005  | rpc getNonce err                     | 获取Nonce值 rpc调用失败                   |
| -1006  | SDK SignTx err                    | 交易签名错误                              |
| -1007  | bal EncodeToBytes err                | RLP编码错误                               |
| -1008  | rpc sendTransaction err              | 发送交易 rpc调用失败                      |
| -1009  | SDK SignTxWithPassphrase err      | 钱包签名交易错误                          |
| -1012  | ContractExtension err                | 合约交易参数解析错误                      |
| -1013  | rpc sendContractTransaction err      | 发送合约交易 rpc调用失败                  |
| -1018  | json unmarshal err                   | json解析错误                              |
| -1019  | rpc getTransactionByHash err         | 查询交易 rpc调用失败                      |
| -1020  | rpc getTransactionReceipt err        | 查询收据 rpc调用失败                      |
| -1021  | SendTxArgs err                       | 发送交易参数解析错误 或 获取Nonce值错误   |

注：其他错误码由BaaS透传返回
