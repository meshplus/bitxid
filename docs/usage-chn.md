# BitXID 中文使用文档

## 概述

BitXID是遵循W3C DID规范的DID框架，可用于快速开发属于自己的DID应用。本文档将介绍BitXID相关功能的使用。

BitXID的主要分为两大模块：账户数字身份模块（Account DID）和链数字身份模块（Chain DID）。前者主要用于构建单链上的账户数字身份，后者主要用于构建跨链网络中的链数字身份。账户数字身份是指基于区块链上的账户地址的数字身份，而链身份是指区块链自己的数字身份。如果您只想在单链上使用DID，则可以只使用Account DID部分功能，如果您想在某个跨链平台上集成数字身份功能，则可以考虑Chain DID + Account DID的组合。

BitXID 支持多种存储方式，给了开发者充足的选择权。BitXID中关于 **Account DID** 和 **Chain DID** 两部分功能的设计使得开发者可以实现两种存储类型的DID应用，一种是 **InternalDocDB**，开发者可以将 DID Doc 存储在链上；另一种是 **ExternalDocDB**，开发者可以将 DID Doc 存储在链下。

## 快速开始

**ExternalDocDB** 模式的 **Chain DID**，此处我们以启动一条4节点中继链后在其上初始化ChainDIDRegistry，并对一条单节点应用链进行其链DID相关操作为例。

```go
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/bitxhub/bitxid"
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage/leveldb"
	"github.com/meshplus/bitxhub-model/pb"
)

func main() {
	// 构造链身份信息文档 Chain Doc
	adminDID := bitxid.DID("did:bitxhub:relayroot:0x00000001")
	relaychainDoc := bitxid.ChainDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      bitxid.DID("did:bitxhub:relayroot:."),
			Type:    int(bitxid.ChainDIDType),
			Created: uint64(time.Now().Second()),
			PublicKey: []bitxid.PubKey{
				{
					ID:   "KEY#1",
					Type: "secp256k1",
					PublicKeyPem: `MIICWjCCAf+gAwIBAgIDDhFLMAoGCCqGSM49BAMCMIGaMQswCQYDVQQGEwJDTjER
					MA8GA1UECBMIWmhlSmlhbmcxETAPBgNVBAcTCEhhbmdaaG91MR8wDQYDVQQJEwZz
					dHJlZXQwDgYDVQQJEwdhZGRyZXNzMQ8wDQYDVQQREwYzMjQwMDAxDzANBgNVBAoT
					BkFnZW5jeTEQMA4GA1UECxMHQml0WEh1YjEQMA4GA1UEAxMHQml0WEh1YjAgFw0y
					MDA4MTEwNTA3MTNaGA8yMDcwMDczMDA1MDcxM1owgZkxCzAJBgNVBAYTAkNOMREw
					DwYDVQQIEwhaaGVKaWFuZzERMA8GA1UEBxMISGFuZ1pob3UxHzANBgNVBAkTBnN0
					cmVldDAOBgNVBAkTB2FkZHJlc3MxDzANBgNVBBETBjMyNDAwMDEOMAwGA1UEChMF
					Tm9kZTExEDAOBgNVBAsTB0JpdFhIdWIxEDAOBgNVBAMTB0JpdFhIdWIwWTATBgcq
					hkjOPQIBBggqhkjOPQMBBwNCAATgjTYEnavxerFuEKJ8C39QUY12xh/TC2E5V7ni
					nmQcOgDDRv5HW4sskTSm/WX2D0BMzwb7XE5ATyoDeM9qcurDozEwLzAOBgNVHQ8B
					Af8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAMBgNVHRMBAf8EAjAAMAoGCCqGSM49
					BAMCA0kAMEYCIQD5Oz1xJvFgzYm/lTzoaO/i0ayPVRgSdBwvK6hEICo5lAIhAMtG
					aswjd2wVA4zB5GPEmJ/tvPUnxrlOAU67AQMYR4zf`,
				}, {
					ID:   "KEY#2",
					Type: "secp256k1",
					PublicKeyPem: `MIICWjCCAf+gAwIBAgIDDhFLMAoGCCqGSM49BAMCMIGaMQswCQYDVQQGEwJDTjER
						MA8GA1UECBMIWmhlSmlhbmcxETAPBgNVBAcTCEhhbmdaaG91MR8wDQYDVQQJEwZz
						dHJlZXQwDgYDVQQJEwdhZGRyZXNzMQ8wDQYDVQQREwYzMjQwMDAxDzANBgNVBAoT
						BkFnZW5jeTEQMA4GA1UECxMHQml0WEh1YjEQMA4GA1UEAxMHQml0WEh1YjAgFw0y
						MDA4MTEwNTA3MTNaGA8yMDcwMDczMDA1MDcxM1owgZkxCzAJBgNVBAYTAkNOMREw
						DwYDVQQIEwhaaGVKaWFuZzERMA8GA1UEBxMISGFuZ1pob3UxHzANBgNVBAkTBnN0
						cmVldDAOBgNVBAkTB2FkZHJlc3MxDzANBgNVBBETBjMyNDAwMDEOMAwGA1UEChMF
						Tm9kZTExEDAOBgNVBAsTB0JpdFhIdWIxEDAOBgNVBAMTB0JpdFhIdWIwWTATBgcq
						hkjOPQIBBggqhkjOPQMBBwNCAATgjTYEnavxerFuEKJ8C39QUY12xh/TC2E5V7ni
						nmQcOgDDRv5HW4sskTSm/WX2D0BMzwb7XE5ATyoDeM9qcurDozEwLzAOBgNVHQ8B
						Af8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAMBgNVHRMBAf8EAjAAMAoGCCqGSM49
						BAMCA0kAMEYCIQD5Oz1xJvFgzYm/lTzoaO/i0ayPVRgSdBwvK6hEICo5lAIhAMtG
						aswjd2wVA4zB5GPEmJ/tvPUnxrlOAU67AQMYR4zf`,
				}, {
					ID:   "KEY#3",
					Type: "secp256k1",
					PublicKeyPem: `MIICWTCCAf+gAwIBAgIDCUHDMAoGCCqGSM49BAMCMIGaMQswCQYDVQQGEwJDTjER
					MA8GA1UECBMIWmhlSmlhbmcxETAPBgNVBAcTCEhhbmdaaG91MR8wDQYDVQQJEwZz
					dHJlZXQwDgYDVQQJEwdhZGRyZXNzMQ8wDQYDVQQREwYzMjQwMDAxDzANBgNVBAoT
					BkFnZW5jeTEQMA4GA1UECxMHQml0WEh1YjEQMA4GA1UEAxMHQml0WEh1YjAgFw0y
					MDA4MTEwNTA3MTRaGA8yMDcwMDczMDA1MDcxNFowgZkxCzAJBgNVBAYTAkNOMREw
					DwYDVQQIEwhaaGVKaWFuZzERMA8GA1UEBxMISGFuZ1pob3UxHzANBgNVBAkTBnN0
					cmVldDAOBgNVBAkTB2FkZHJlc3MxDzANBgNVBBETBjMyNDAwMDEOMAwGA1UEChMF
					Tm9kZTMxEDAOBgNVBAsTB0JpdFhIdWIxEDAOBgNVBAMTB0JpdFhIdWIwWTATBgcq
					hkjOPQIBBggqhkjOPQMBBwNCAAQ9IPBBKkqSwWkwDdK+ARw2qlBmBD9bF8HJ0z3P
					XeKaTmnnEBJu1e0vjHl+uQGBz5x1ulBRVeq4xhmkZtPZByO+ozEwLzAOBgNVHQ8B
					Af8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAMBgNVHRMBAf8EAjAAMAoGCCqGSM49
					BAMCA0gAMEUCIQCMgYSwQ9go1jjAcC4SxpJl4moA8Ba/GEb0qwFPaNmSCwIgDEOo
					UpUSNYEQJvahR4BxxVLOBf/CNlKhAGBVNKTccxk`,
				}, {
					ID:   "KEY#4",
					Type: "secp256k1",
					PublicKeyPem: `MIICWTCCAf+gAwIBAgIDCGR3MAoGCCqGSM49BAMCMIGaMQswCQYDVQQGEwJDTjER
					MA8GA1UECBMIWmhlSmlhbmcxETAPBgNVBAcTCEhhbmdaaG91MR8wDQYDVQQJEwZz
					dHJlZXQwDgYDVQQJEwdhZGRyZXNzMQ8wDQYDVQQREwYzMjQwMDAxDzANBgNVBAoT
					BkFnZW5jeTEQMA4GA1UECxMHQml0WEh1YjEQMA4GA1UEAxMHQml0WEh1YjAgFw0y
					MDA4MTEwNTA3MTRaGA8yMDcwMDczMDA1MDcxNFowgZkxCzAJBgNVBAYTAkNOMREw
					DwYDVQQIEwhaaGVKaWFuZzERMA8GA1UEBxMISGFuZ1pob3UxHzANBgNVBAkTBnN0
					cmVldDAOBgNVBAkTB2FkZHJlc3MxDzANBgNVBBETBjMyNDAwMDEOMAwGA1UEChMF
					Tm9kZTQxEDAOBgNVBAsTB0JpdFhIdWIxEDAOBgNVBAMTB0JpdFhIdWIwWTATBgcq
					hkjOPQIBBggqhkjOPQMBBwNCAARN1y/FhZpSg1kpXF38szDNRXdPkqoc8oRKdGzv
					3HdhtBdUO7jXe2xNaWVtNMGXVo+NuBi5t9qEoo+euxfnjlc9ozEwLzAOBgNVHQ8B
					Af8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAMBgNVHRMBAf8EAjAAMAoGCCqGSM49
					BAMCA0gAMEUCIQCbsG7E158uzqYCzrrnrr2Xsnz7f5cFA2o4SXAF7R/IyAIgSxYS
					MGj0g0OBcxJqwTyyvF2FFOhlWjF9nq2eYK/rlzI`,
				},
			},
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1", "KEY#2", "KEY#3", "KEY#4"}, Strategy: "1-of-4"}},
		},
	}
	relaychainDID := relaychainDoc.ID
	docBytes, _ := relaychainDoc.Marshal()
	relaychainDocAddr := fakeStore(&relaychainDoc) // 假设将Doc进行了存储，返回了存储地址
	relaychainDocHash := fakeHash(docBytes)        // 假设将Doc进行了哈希，返回了哈希结果

	mcaller := bitxid.DID("did:bitxhub:relayroot:0x12345678")
	chainDID := bitxid.DID("did:bitxhub:appchain001:.")
	appchainDoc := bitxid.ChainDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      chainDID,
			Type:    int(bitxid.ChainDIDType),
			Created: 1616985208,
			PublicKey: []bitxid.PubKey{
				{
					ID:   "KEY#1",
					Type: "secp256k1",
					PublicKeyPem: `MIICWjCCAf+gAwIBAgIDDhFLMAoGCCqGSM49BAMCMIGaMQswCQYDVQQGEwJDTjER
					MA8GA1UECBMIWmhlSmlhbmcxETAPBgNVBAcTCEhhbmdaaG91MR8wDQYDVQQJEwZz
					dHJlZXQwDgYDVQQJEwdhZGRyZXNzMQ8wDQYDVQQREwYzMjQwMDAxDzANBgNVBAoT
					BkFnZW5jeTEQMA4GA1UECxMHQml0WEh1YjEQMA4GA1UEAxMHQml0WEh1YjAgFw0y
					MDA4MTEwNTA3MTNaGA8yMDcwMDczMDA1MDcxM1owgZkxCzAJBgNVBAYTAkNOMREw
					DwYDVQQIEwhaaGVKaWFuZzERMA8GA1UEBxMISGFuZ1pob3UxHzANBgNVBAkTBnN0
					cmVldDAOBgNVBAkTB2FkZHJlc3MxDzANBgNVBBETBjMyNDAwMDEOMAwGA1UEChMF
					Tm9kZTExEDAOBgNVBAsTB0JpdFhIdWIxEDAOBgNVBAMTB0JpdFhIdWIwWTATBgcq
					hkjOPQIBBggqhkjOPQMBBwNCAATgjTYEnavxerFuEKJ8C39QUY12xh/TC2E5V7ni
					nmQcOgDDRv5HW4sskTSm/WX2D0BMzwb7XE5ATyoDeM9qcurDozEwLzAOBgNVHQ8B
					Af8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAMBgNVHRMBAf8EAjAAMAoGCCqGSM49
					BAMCA0kAMEYCIQD5Oz1xJvFgzYm/lTzoaO/i0ayPVRgSdBwvK6hEICo5lAIhAMtG
					aswjd2wVA4zB5GPEmJ/tvPUnxrlOAU67AQMYR4zf`,
				},
			},
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}, Strategy: "1-of-1"}},
		},
	}

	// 初始化存储：
	dirTable, _ := ioutil.TempDir("", "did.table")
	l := log.NewWithModule("chain-did")
	sTable, _ := leveldb.New(dirTable)

	// 构建一个 ChainDIDRegistry 实例（ExternalDocDB模式，无需WithChainDocStorage）：
	mr, _ := bitxid.NewChainDIDRegistry(sTable, l,
		bitxid.WithAdmin(adminDID),
		bitxid.WithGenesisChainDocInfo(
			bitxid.DocInfo{ID: relaychainDID, Addr: relaychainDocAddr, Hash: relaychainDocHash[:]},
		),
	)

	// 初始化 ChainDIDRegistry
	_ = mr.SetupGenesis()

	// 申请 Chain DID：
	mr.Apply(mcaller, chainDID)

	// 审批 Chain DID：
	mr.AuditApply(chainDID, true)

	// 注册 Chain DID：
	docAddr := fakeStore(&appchainDoc) // 假设将Doc进行了存储，返回了存储地址
	docBytes, _ = appchainDoc.Marshal()
	docHash := fakeHash(docBytes) // 假设将Doc进行了哈希，返回了哈希结果
	mr.Register(chainDID, docAddr, docHash[:])

	// 更新 Chain DID：
	appchainDoc.Updated = uint64(1616986227)
	docBytes, _ = bitxid.Marshal(appchainDoc)
	docAddr = fakeStore(&appchainDoc) // 假设将Doc进行了存储，返回了存储地址
	docHash = fakeHash(docBytes)      // 假设将Doc进行了哈希，返回了哈希结果
	mr.Update(chainDID, docAddr, docHash[:])

	// 冻结 Chain DID：
	mr.Freeze(chainDID)

	// 解冻 Chain DID：
	mr.UnFreeze(chainDID)

	// 解析 Chain DID：
	item, _, _, _ := mr.Resolve(chainDID)
	fmt.Println(item)
	docGet := fakeGetDoc(item.DocAddr) // 假设去链下获取存储的文档原文
	docGetBytes, _ := docGet.Marshal()
	docHashGet := fakeHash(docGetBytes)
	if !bytes.Equal(docHashGet[:], item.DocHash) {
		return // 哈希不匹配，说明链下存储的文档被篡改不可信，解析失败
	}
	interchain_tx := fakeReceiveInterchainTX()
	if !fakeVerify(docGet.Authentication, interchain_tx) {
		return // 交易合法性验证失败，说明交易发起者不符合权限要求
	}

	// 删除 Chain DID：
	mr.Delete(chainDID)
	fakeDeleteDoc(chainDID) // 假设此处去删除链下存储的文档

	// 清除存储
	mr.Table.Close()
	os.RemoveAll(dirTable)
}

func fakeHash(docBytes []byte) [32]byte {
	return sha256.Sum256(docBytes)
}

func fakeStore(d bitxid.Doc) string {
	return "/storage/address/to/doc/" + fmt.Sprint(time.Now().Second())
}

func fakeGetDoc(docAddr string) bitxid.ChainDoc {
	return bitxid.ChainDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      "did:bitxhub:appchain001:.",
			Type:    int(bitxid.ChainDIDType),
			Created: 1616985208,
			Updated: 1616986227,
			PublicKey: []bitxid.PubKey{
				{
					ID:   "KEY#1",
					Type: "secp256k1",
					PublicKeyPem: `MIICWjCCAf+gAwIBAgIDDhFLMAoGCCqGSM49BAMCMIGaMQswCQYDVQQGEwJDTjER
					MA8GA1UECBMIWmhlSmlhbmcxETAPBgNVBAcTCEhhbmdaaG91MR8wDQYDVQQJEwZz
					dHJlZXQwDgYDVQQJEwdhZGRyZXNzMQ8wDQYDVQQREwYzMjQwMDAxDzANBgNVBAoT
					BkFnZW5jeTEQMA4GA1UECxMHQml0WEh1YjEQMA4GA1UEAxMHQml0WEh1YjAgFw0y
					MDA4MTEwNTA3MTNaGA8yMDcwMDczMDA1MDcxM1owgZkxCzAJBgNVBAYTAkNOMREw
					DwYDVQQIEwhaaGVKaWFuZzERMA8GA1UEBxMISGFuZ1pob3UxHzANBgNVBAkTBnN0
					cmVldDAOBgNVBAkTB2FkZHJlc3MxDzANBgNVBBETBjMyNDAwMDEOMAwGA1UEChMF
					Tm9kZTExEDAOBgNVBAsTB0JpdFhIdWIxEDAOBgNVBAMTB0JpdFhIdWIwWTATBgcq
					hkjOPQIBBggqhkjOPQMBBwNCAATgjTYEnavxerFuEKJ8C39QUY12xh/TC2E5V7ni
					nmQcOgDDRv5HW4sskTSm/WX2D0BMzwb7XE5ATyoDeM9qcurDozEwLzAOBgNVHQ8B
					Af8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAMBgNVHRMBAf8EAjAAMAoGCCqGSM49
					BAMCA0kAMEYCIQD5Oz1xJvFgzYm/lTzoaO/i0ayPVRgSdBwvK6hEICo5lAIhAMtG
					aswjd2wVA4zB5GPEmJ/tvPUnxrlOAU67AQMYR4zf`,
				},
			},
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}, Strategy: "1-of-1"}},
		},
	}
}

func fakeReceiveInterchainTX() *pb.Transaction {
	return &pb.Transaction{}
}

func fakeVerify(a []bitxid.Auth, tx *pb.Transaction) bool {
	// 此处理论上只需要成功匹配一种验证方式就可以返回true
	return true
}

func fakeDeleteDoc(did bitxid.DID) error {
	// 去链下存储删除对应文档
	return nil
}
```

其他示例包括**Chain DID**的**InternalDocDB**模式以及**Account DID**的**ExternalDocDB**模式和**InternalDocDB**模式示例见[examples](./examples)。

## 实例化

以Chain DID Registry的实例化为例来进行说明，Account DID Registry的实例化方式类似。

传入实现`Storage`接口的数据结构、日志即可完成实例化（关于`Storage`接口的详细信息见[此处](https://github.com/meshplus/bitxhub-kit/blob/master/storage/storage.go)）：

```go
// 省略其他代码，见examples/chain-did/internal/example.go
	mr, _ := bitxid.NewChainDIDRegistry(sTable, l,
		bitxid.WithChainDocStorage(sDocdb),
		bitxid.WithAdmin(adminDID),
		bitxid.WithGenesisChainDocContent(&relaychainDoc),
	)
```

实例化函数拥有如下几个选项函数：

+ `WithChainDocStorage`：使用该选项后Registry的将为**InternalDocDB**模式，如果不使用该选项则为**ExternalDocDB**模式，入参也是一个`Storage`接口的数据结构。
+ `WithAdmin`：指定管理员账号
+ `WithGenesisChainDocContent`：用于**InternalDocDB**模式，指定网络中第一条链的身份信息文档原文内容
+ `WithGenesisChainDocInfo`：用于**ExternalDocDB**模式，指定网络中第一条链的身份信息文档信息（包括存储地址等）

## 基础功能

此部分是 **Chain DID Registry** 和 **Account DID Registry** 都有的功能，以 Chain DID Registry 为例进行说明。

### 初始化

进行初始化不需要传入任何参数：

```go
// 省略其他代码，见examples/chain-did/internal/example.go
_ = mr.SetupGenesis()
```

### 获取自己的身份

对链管理来说是`GenesisChainDID`，对账户管理来说是`GenesisAccountDID`。

```go
chainDID := mr.GetSelfID()
```

### 获取管理员列表

```go
admins := mr.GetAdmins()
```

### 添加管理员

```go
admin := bitxid.DID("did:bitxhub:relaychain001:0x12345678")
err := mr.AddAdmin(admin)
```

### 移除管理员

```go
err := mr.RemoveAdmin(admin)
```

### 查询是否是管理员

```go
err := mr.HasAdmin(admin)
```

## Chain DID

以下是Chain DID的特有功能。

### 申请

申请一个Chain DID的所有权：

```go
// 省略其他代码，见examples/chain-did/internal/example.go
chainDID := bitxid.DID("did:bitxhub:appchain001:.")
mcaller := bitxid.DID("did:bitxhub:relayroot:0x12345678")
err := mr.Apply(mcaller, method)
```

### 审批

对某个Chain DID的申请进行审批（审批结果为“通过”）：

```go
err := mr.AuditApply(chainDID, true)
```

对某个Chain DID的申请进行审批（审批结果为“驳回”）：

```go
err := mr.AuditApply(method, false)
```

### 注册

在审核通过后，可以进行注册，如果是 **ExternalDocDB** 模式：

```go
docAddr := fakeStore(&appchainDoc) // 假设将Doc进行了存储，返回了存储地址
docBytes, _ = appchainDoc.Marshal()
docHash := fakeHash(docBytes) // 假设将Doc进行了哈希，返回了哈希结果
mr.Register(chainDID, docAddr, docHash[:])
```

此处是 **ExternalDocDB** 模式，因此需要自己手动将相关信息的文档进行存储，并进行哈希，然后将`chainDID`, `docAddr`, `docHash`作为参数传入`Register`方法。

如果是 **InternalDocDB** 模式：

```go
mr.RegisterWithDoc(&appchainDoc)
```

 **InternalDocDB** 模式看上去更加简单，因为链上的逻辑能帮你完成所有事情——各种格式变换以及存储，但是链上的计算和存储是非常昂贵的。

### 更新

更新一个Chain DID所绑定的信息，如果是 **ExternalDocDB** 模式：

```go
appchainDoc.Updated = uint64(1616986227)
docBytes, _ = bitxid.Marshal(appchainDoc)
docAddr = fakeStore(&appchainDoc) // 假设将Doc进行了存储，返回了存储地址
docHash = fakeHash(docBytes)      // 假设将Doc进行了哈希，返回了哈希结果
mr.Update(chainDID, docAddr, docHash[:])
```

此处是 **ExternalDocDB** 模式，因此需要自己手动将相关信息的文档进行存储，并进行哈希，然后将`chainDID`, `docAddr`, `docHash`作为参数传入`Update`方法。

如果是 **InternalDocDB** 模式：

```go
appchainDoc.Updated = uint64(1616986227)
mr.UpdateWithDoc(&appchainDoc)
```

### 解析

获得相关Chain DID的信息，如果是 **ExternalDocDB** 模式：

```go
item, _, _, _ := mr.Resolve(chainDID)
fmt.Println(item)
docGet := fakeGetDoc(item.DocAddr) // 假设去链下获取存储的文档原文
docGetBytes, _ := docGet.Marshal()
docHashGet := fakeHash(docGetBytes)
if !bytes.Equal(docHashGet[:], item.DocHash) {
  return // 哈希不匹配，说明链下存储的文档被篡改不可信，解析失败
}
interchain_tx := fakeReceiveInterchainTX()
if !fakeVerify(docGet.Authentication, interchain_tx) {
  return // 交易合法性验证失败，说明交易发起者不符合权限要求
}
```

此处是 **ExternalDocDB** 模式，因此需要自己手动用存储地址去链下存储获取信息文档，然后将文档内容进行哈希，并和原哈希进行比对以验证文档。然后在文档的使用上，bitxid虽然不强迫但建议使用者在验证身份时只需要符合`Authentication`数组中的一条验证规则即可证明用户确实已经授权交易。

如果是 **InternalDocDB** 模式：

```go
item, docGet, _, _ := mr.Resolve(chainDID)
fmt.Println(item)
fmt.Println(docGet)
interchain_tx := fakeReceiveInterchainTX()
if !fakeVerify(docGet.Authentication, interchain_tx) {
  return // 交易合法性验证失败，说明交易发起者不符合权限要求
}
```

可以从链上直接获取到DID文档，文档的使用方式一样。

### 删除

删除相关Chain DID的信息，如果是 **ExternalDocDB** 模式：

```go
mr.Delete(chainDID)
fakeDeleteDoc(chainDID)
```

链下存储的DID文档需要调用者手动去删除。

如果是 **InternalDocDB** 模式：

```go
mr.Delete(chainDID)
```

## Account DID

以下是Account DID的特有功能。

### 获取链身份

获取Account DID Registry所在链的身份：

```
chainDID := r.GetChainDID()
```

### 注册

Account DID不需要申请可以直接进行注册，bitxid强烈建议每个地址应当只能注册以自己为`address`，以自己所在链为`chain-name`的DID（`did:bitxhub:chain-name:address`）。如果开发者不得不打破这个规定，可以自己编写Account DID的申请等功能。

**ExternalDocDB** 模式下的注册：

```go
docAddr := fakeStore(&accountDoc) // 假设将Doc进行了存储，返回了存储地址
docBytes, _ = accountDoc.Marshal()
docHash := fakeHash(docBytes) // 假设将Doc进行了哈希，返回了哈希结果
ar.Register(accountDID, docAddr, docHash[:])
```

此处是 **ExternalDocDB** 模式，因此需要自己手动将相关信息的文档进行存储，并进行哈希，然后`将accountDID`, `docAddr`, `docHash`作为参数传入`Register`方法。

**InternalDocDB** 模式下的注册：

```go
ar.RegisterWithDoc(&accountDoc)
```

 **InternalDocDB** 模式看上去更加简单，因为链上的逻辑能帮你完成所有事情——各种格式变换以及存储，但是链上的计算和存储是非常昂贵的。

### 更新

更新一个Account DID所绑定的信息，如果是 **ExternalDocDB** 模式：

```go
accountDoc.Updated = uint64(1616986228)
docAddr = fakeStore(&accountDoc) // 假设将Doc进行了存储，返回了存储地址
docBytes, _ = bitxid.Marshal(accountDoc)
docHash = fakeHash(docBytes) // 假设将Doc进行了哈希，返回了哈希结果
ar.Update(accountDID, docAddr, docHash[:])
```

此处是 **ExternalDocDB** 模式，因此需要自己手动将相关信息的文档进行存储，并进行哈希，然后将`accountDID`, `docAddr`, `docHash`作为参数传入`Update`方法。

如果是 **InternalDocDB** 模式：

```go
accountDoc.Updated = uint64(1616986228)
ar.UpdateWithDoc(&accountDoc)
```

### 冻结

```go
ar.Freeze(accountDID)
```

### 解冻

```go
ar.UnFreeze(accountDID)
```

### 解析

如果是 **ExternalDocDB** 模式：

```go
item, _, _, _ := ar.Resolve(accountDID)
fmt.Println(item)
docGet := fakeGetDoc(item.DocAddr) // 假设去链下获取存储的文档原文
docGetBytes, _ := docGet.Marshal()
docHashGet := fakeHash(docGetBytes)
if !bytes.Equal(docHashGet[:], item.DocHash) {
  return // 哈希不匹配，说明链下存储的文档被篡改不可信，解析失败
}
interchain_tx := fakeReceiveInterchainTX()
if !fakeVerify(docGet.Authentication, interchain_tx) {
  return // 交易合法性验证失败，说明交易发起者不符合权限要求
}
```

此处是 **ExternalDocDB** 模式，因此需要自己手动用存储地址去链下存储获取信息文档，然后将文档内容进行哈希，并和原哈希进行比对以验证文档。然后在文档的使用上，与ChainDoc类似。

如果是 **InternalDocDB** 模式：

```go
item, docGet, _, _ := ar.Resolve(accountDID)
fmt.Println(item)
fmt.Println(docGet)
interchain_tx := fakeReceiveInterchainTX()
if !fakeVerify(docGet.Authentication, interchain_tx) {
  return // 交易合法性验证失败，说明交易发起者不符合权限要求
}
```

可以从链上直接获取到DID文档，文档的使用方式一样。

### 删除

删除相关Chain DID的信息，如果是 **ExternalDocDB** 模式：

```go
ar.Delete(accountDID)
fakeDeleteDoc(accountDID) // 假设此处去删除链下存储的文档
```

链下存储的DID文档需要调用者手动去删除。

如果是 **InternalDocDB** 模式：

```go
ar.Delete(accountDID)
```


