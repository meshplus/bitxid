# BitXID 中文使用文档

## 概述

BitXID是遵循W3C DID规范的DID框架，可用于快速开发属于自己的DID应用。本文档将展示如何使用BitXID来开发属于自己的DID应用。

BitXID的主要分为两大模块：账户数字身份模块（Account DID）和链数字身份模块（Chain DID）。前者主要用于构建单链上的账户数字身份，后者主要用于构建跨链网络中的链数字身份。账户数字身份是指基于区块链上的账户地址的数字身份，而链身份是指区块链自己的数字身份。如果您只想在单链上使用DID，则可以只使用Account DID部分功能，如果您想在某个跨链平台上集成数字身份功能，则可以考虑Chain DID + Account DID的组合。

BitXID 支持多种存储方式，给了开发者充足的选择权。BitXID中关于 **Account DID** 和 **Chain DID** 两部分功能的设计是的开发者者可以实现两种存储类型的DID应用，一种是 **InternalDocDB**，开发者可以将 DID Doc 存储在链上；另一种是 **ExternalDocDB**，开发者可以将 DID Doc 存储在链下。

## 快速开始

### Chain DID

**ExternalDocDB** 模式的 **Chain DID**：

```go
package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/bitxhub/bitxid"
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage/leveldb"
)

func main() {
	dir_table, _ := ioutil.TempDir("testdata", "did.table")
	drtPath := dir_table
	l := log.NewWithModule("did")
	s_table, _ := leveldb.New(drtPath)

	// 构建一个 DIDRegistry 实例：
	mr, _ := bitxid.NewMethodRegistry(s_table, l)
	// 初始化 DIDRegistry：
	_ = mr.SetupGenesis()

	mcaller := bitxid.DID("did:bitxhub:relayroot:0x12345678")
	chainDID := bitxid.DID("did:bitxhub:appchain001:.")
	// 申请 Chain DID：
	mr.Apply(mcaller, chainDID)

	// 审批 Chain DID：
	mr.AuditApply(chainDID, true)

	doc := bitxid.MethodDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      chainDID,
			Type:    fmt.Sprint(bitxid.MethodDocType),
			Created: uint64(time.Now().Second()),
			PublicKey: []bitxid.PubKey{
				{
					ID:           "KEY#1",
					Type:         "Ed25519",
					PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
				}},
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
		},
	}
	docABytes, _ := bitxid.Struct2Bytes(doc)
	docHash := sha256.Sum256(docABytes)
	docAddr := "/addr1/to/doc"
	// 注册 Chain DID：
	mr.Register(bitxid.DocOption{ID: chainDID, Addr: docAddr, Hash: docHash[:]})

	// 更新 Chain DID：
	doc.Updated = uint64(time.Now().Second())
	docABytes, _ = bitxid.Struct2Bytes(doc)
	docHash = sha256.Sum256(docABytes)
	docAddr = "./addr2/to/doc"
	mr.Update(bitxid.DocOption{ID: chainDID, Addr: docAddr, Hash: docHash[:]})

	// 冻结DID：
	mr.Freeze(chainDID)

	// 解冻DID：
	mr.UnFreeze(chainDID)

	// 解析 Chain DID：
	item, _, _, _ := mr.Resolve(chainDID)
	fmt.Println(item)

	// 删除 Chain DID：
	mr.Delete(chainDID)
}
```

**InternalDocDB** 模式的 **Chain DID**：

```go
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/bitxhub/bitxid"
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage/leveldb"
)

func main() {
	dir_table, _ := ioutil.TempDir("testdata", "did.table")
	dir_docdb, _ := ioutil.TempDir("testdata", "did.docdb")
	drtPath := dir_table
	ddbPath := dir_docdb
	l := log.NewWithModule("did")
	s_table, _ := leveldb.New(drtPath)
	s_docdb, _ := leveldb.New(ddbPath)

	// 构建一个 DIDRegistry 实例：
	mr, _ := bitxid.NewMethodRegistry(s_table, l, bitxid.WithMethodDocStorage(s_docdb))
	// 初始化 DIDRegistry：
	_ = mr.SetupGenesis()

	mcaller := bitxid.DID("did:bitxhub:relayroot:0x12345678")
	chainDID := bitxid.DID("did:bitxhub:appchain001:.")
	// 申请 Chain DID：
	mr.Apply(mcaller, chainDID)

	// 审批 Chain DID：
	mr.AuditApply(chainDID, true)

	doc := bitxid.MethodDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      chainDID,
			Type:    fmt.Sprint(bitxid.MethodDocType),
			Created: uint64(time.Now().Second()),
			PublicKey: []bitxid.PubKey{
				{
					ID:           "KEY#1",
					Type:         "Ed25519",
					PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
				}},
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
		},
	}
	// 注册 Chain DID：
	mr.Register(bitxid.DocOption{Content: &doc})

	// 更新 Chain DID：
	doc.Updated = uint64(time.Now().Second())
	mr.Update(bitxid.DocOption{Content: &doc})

	// 冻结DID：
	mr.Freeze(chainDID)

	// 解冻DID：
	mr.UnFreeze(chainDID)

	// 解析 Chain DID：
	item, docGet, _, _ := mr.Resolve(chainDID)
	fmt.Println(item)
	fmt.Println(docGet)

	// 删除 Chain DID：
	mr.Delete(chainDID)
}
```

### Account DID

**ExternalDocDB** 模式的 **Account DID**：

```go
package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/bitxhub/bitxid"
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage/leveldb"
)

func main() {
	dir_table, _ := ioutil.TempDir("testdata", "did.table")
	drtPath := dir_table
	l := log.NewWithModule("did")
	s_table, _ := leveldb.New(drtPath)

	// 构建一个 DIDRegistry 实例：
	r, _ := bitxid.NewDIDRegistry(s_table, l)
	// 初始化 DIDRegistry：
	_ = r.SetupGenesis()

	AccountDID := bitxid.DID("did:bitxhub:appchain001:0x12345678")
	doc := bitxid.DIDDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      AccountDID,
			Type:    fmt.Sprint(bitxid.DIDDocType),
			Created: uint64(time.Now().Second()),
			PublicKey: []bitxid.PubKey{
				{
					ID:           "KEY#1",
					Type:         "Ed25519",
					PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
				}},
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
		},
	}
	docABytes, _ := bitxid.Struct2Bytes(doc)
	docHash := sha256.Sum256(docABytes)
	docAddr := "./addr1/" + string(AccountDID)
	// 注册 Account DID：
	r.Register(bitxid.DocOption{ID: AccountDID, Addr: docAddr, Hash: docHash[:]})

	// 更新 Account DID：
	doc.Updated = uint64(time.Now().Second())
	docABytes, _ = bitxid.Struct2Bytes(doc)
	docHash = sha256.Sum256(docABytes)
	docAddr = "./addr2/" + string(AccountDID)
	r.Update(bitxid.DocOption{ID: AccountDID, Addr: docAddr, Hash: docHash[:]})

	// 冻结 Account DID：
	r.Freeze(AccountDID)

	// 解冻 Account DID：
	r.UnFreeze(AccountDID)

	// 解析 Account DID：
	item, _, _, _ := r.Resolve(AccountDID)
	fmt.Println(item)

	// 删除 Account DID：
	r.Delete(AccountDID)
}
```

**InternalDocDB** 模式的 **Account DID**：

```go
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/bitxhub/bitxid"
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-kit/storage/leveldb"
)

func main() {
	dir_table, _ := ioutil.TempDir("testdata", "did.table")
	dir_docdb, _ := ioutil.TempDir("testdata", "did.docdb")
	drtPath := dir_table
	ddbPath := dir_docdb
	l := log.NewWithModule("did")
	s_table, _ := leveldb.New(drtPath)
	s_docdb, _ := leveldb.New(ddbPath)

	// 构建一个 DIDRegistry 实例：
	r, _ := bitxid.NewDIDRegistry(s_table, l, bitxid.WithDIDDocStorage(s_docdb))
	// 初始化 DIDRegistry：
	_ = r.SetupGenesis()

	AccountDID := bitxid.DID("did:bitxhub:appchain001:0x12345678")
	doc := bitxid.DIDDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      AccountDID,
			Type:    fmt.Sprint(bitxid.DIDDocType),
			Created: uint64(time.Now().Second()),
			PublicKey: []bitxid.PubKey{
				{
					ID:           "KEY#1",
					Type:         "Ed25519",
					PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
				}},
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
		},
	}
	// 注册 Account DID：
	r.Register(bitxid.DocOption{Content: &doc})

	// 更新 Account DID：
	doc.Updated = uint64(time.Now().Second())
	r.Update(bitxid.DocOption{Content: &doc})

	// 冻结 Account DID：
	r.Freeze(AccountDID)

	// 解冻 Account DID：
	r.UnFreeze(AccountDID)

	// 解析 Account DID：
	item, docGet, _, _ := r.Resolve(AccountDID)
	fmt.Println(item)
	fmt.Println(docGet)

	// 删除 Account DID：
	r.Delete(AccountDID)
}
```

## 基础功能

此部分是 **Account DID** 和 **Chain DID** 都有的功能。

### 初始化

```go
// 省略其他代码
superAdmin := bitxid.DID("did:bitxhub:relaychain001:0x00000001")
mr, _ := bitxid.NewMethodRegistry(s_table, l, bitxid.WithMethodAdmin(superAdmin))
_ = mr.SetupGenesis()
```

### 获取所在链身份

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

### 实例化

```go
// InternalDocDB 模式：
r, _ := bitxid.NewMethodRegistry(s_table, l, bitxid.WithMethodDocStorage(s_docdb))
// ExternalDocDB 模式：
r, _ := bitxid.NewMethodRegistry(s_table, l)
```

### 申请

申请一个Chain DID的所有权：

```go
// Omit part of the code
chainDID := bitxid.DID("did:bitxhub:appchain001:.")
mcaller := bitxid.DID("did:bitxhub:relayroot:0x12345678")
err := mr.Apply(mcaller, method)
```

### 审批

对某个Chain DID的申请进行审批（审批结果为“通过”）：

```go
// Omit part of the code
err := mr.AuditApply(chainDID, true)
```

对某个Chain DID的申请进行审批（审批结果为“驳回”）：

```go
// Omit part of the code
err := mr.AuditApply(method, false)
```

### 注册

在审核通过后，可以进行注册，如果是 **ExternalDocDB** 模式：

```go
doc := bitxid.MethodDoc{
  BasicDoc: bitxid.BasicDoc{
    ID:      chainDID,
    Type:    fmt.Sprint(bitxid.MethodDocType),
    Created: uint64(time.Now().Second()),
    PublicKey: []bitxid.PubKey{
      {
        ID:           "KEY#1",
        Type:         "Ed25519",
        PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
      }},
    Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
  },
}
docABytes, _ := bitxid.Struct2Bytes(doc)
docHash := sha256.Sum256(docABytes)
docAddr := "/addr1/to/doc"

mr.Register(bitxid.DocOption{ID: chainDID, Addr: docAddr, Hash: docHash[:]})
```

如果是 **InternalDocDB** 模式：

```go
doc := bitxid.MethodDoc{
  BasicDoc: bitxid.BasicDoc{
    ID:      chainDID,
    Type:    fmt.Sprint(bitxid.MethodDocType),
    Created: uint64(time.Now().Second()),
    PublicKey: []bitxid.PubKey{
      {
        ID:           "KEY#1",
        Type:         "Ed25519",
        PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
      }},
    Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
  },
}

mr.Register(bitxid.DocOption{Content: &doc})
```

 **InternalDocDB** 模式下看上去更加简单，因为链上的逻辑能帮你完成所有事情——各种格式变换以及存储，但是链上的存储是非常昂贵的。

### 更新

更新一个Chain DID所绑定的信息，如果是 **ExternalDocDB** 模式：

```go
doc.Updated = uint64(time.Now().Second())
docABytes, _ = bitxid.Struct2Bytes(doc)
docHash = sha256.Sum256(docABytes)
docAddr = "./addr2/to/doc"
mr.Update(bitxid.DocOption{ID: chainDID, Addr: docAddr, Hash: docHash[:]})
```

如果是 **InternalDocDB** 模式：

```go
doc.Updated = uint64(time.Now().Second())
mr.Update(bitxid.DocOption{Content: &doc})
```

### 解析

获得相关Chain DID的在链上的信息，如果是 **ExternalDocDB** 模式：

```go
item, _, _, _ := mr.Resolve(chainDID)
fmt.Println(item)
```

此时还需要通过item里的地址和哈希去获取并验证链下存储的DID文档。

如果是 **InternalDocDB** 模式：

```go
item, docGet, _, _ := mr.Resolve(chainDID)
fmt.Println(item)
fmt.Println(docGet)
```

可以从链上直接获取到DID文档。

### 删除

```go
mr.Delete(chainDID)
```

这里需要注意，如果是 **ExternalDocDB** 模式，则链下存储的DID文档需要调用者手动去删除。

## Account DID

### 实例化

```go
// ExternalDocDB 模式：
r, _ := bitxid.NewDIDRegistry(s_table, l)
// InternalDocDB 模式：
r, _ := bitxid.NewDIDRegistry(s_table, l, bitxid.WithDIDDocStorage(s_docdb))
```

### 注册

Account DID不需要申请可以直接进行注册，bitxid强烈建议每个地址应当只能注册以自己为`address`，以自己所在链为`chain-name`的DID（`did:bitxhub:chain-name:address`）。如果开发者不得不打破这个规定，可以自己编写Account DID的申请等功能。

**ExternalDocDB** 模式下的注册：

```go
AccountDID := bitxid.DID("did:bitxhub:appchain001:0x12345678")
doc := bitxid.DIDDoc{
  BasicDoc: bitxid.BasicDoc{
    ID:      AccountDID,
    Type:    fmt.Sprint(bitxid.DIDDocType),
    Created: uint64(time.Now().Second()),
    PublicKey: []bitxid.PubKey{
      {
        ID:           "KEY#1",
        Type:         "Ed25519",
        PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
      }},
    Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
  },
}
docABytes, _ := bitxid.Struct2Bytes(doc)
docHash := sha256.Sum256(docABytes)
docAddr := "./addr1/" + string(AccountDID)

r.Register(bitxid.DocOption{ID: AccountDID, Addr: docAddr, Hash: docHash[:]})
```

**InternalDocDB** 模式下的注册：

```go
AccountDID := bitxid.DID("did:bitxhub:appchain001:0x12345678")
doc := bitxid.DIDDoc{
  BasicDoc: bitxid.BasicDoc{
    ID:      AccountDID,
    Type:    fmt.Sprint(bitxid.DIDDocType),
    Created: uint64(time.Now().Second()),
    PublicKey: []bitxid.PubKey{
      {
        ID:           "KEY#1",
        Type:         "Ed25519",
        PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
      }},
    Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1"}}},
  },
}

r.Register(bitxid.DocOption{Content: &doc})
```

### 更新

更新一个Account DID所绑定的信息，如果是 **ExternalDocDB** 模式：

```go
doc.Updated = uint64(time.Now().Second())
docABytes, _ = bitxid.Struct2Bytes(doc)
docHash = sha256.Sum256(docABytes)
docAddr = "./addr2/" + string(AccountDID)
r.Update(bitxid.DocOption{ID: AccountDID, Addr: docAddr, Hash: docHash[:]})
```

如果是 **InternalDocDB** 模式：

```go
doc.Updated = uint64(time.Now().Second())
r.Update(bitxid.DocOption{Content: &doc})
```

### 冻结

```go
r.Freeze(AccountDID)
```

### 解冻

```go
r.UnFreeze(AccountDID)
```

### 解析

如果是 **ExternalDocDB** 模式：

```go
item, _, _, _ := r.Resolve(AccountDID)
fmt.Println(item)
```

如果是 **InternalDocDB** 模式：

```go
item, docGet, _, _ := r.Resolve(AccountDID)
fmt.Println(item)
fmt.Println(docGet)
```

### 删除

```go
r.Delete(AccountDID)
```

这里需要注意，如果是 **ExternalDocDB** 模式，则链下存储的DID文档需要调用者手动去删除。

