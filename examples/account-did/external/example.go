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
	// 构造链身份信息文档 Account Doc
	adminDoc := bitxid.AccountDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      bitxid.DID("did:bitxhub:appchain001:0x00000001"),
			Type:    int(bitxid.AccountDIDType),
			Created: uint64(time.Now().Second()),
			PublicKey: []bitxid.PubKey{
				{
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
	adminDID := adminDoc.ID
	docBytes, _ := adminDoc.Marshal()
	adminDocAddr := fakeStore(&adminDoc) // 假设将Doc进行了存储，返回了存储地址
	adminDocHash := fakeHash(docBytes)   // 假设将Doc进行了哈希，返回了哈希结果

	accountDoc := bitxid.AccountDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      bitxid.DID("did:bitxhub:appchain001:0x12345678"),
			Type:    int(bitxid.AccountDIDType),
			Created: 1616985209,
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
	accountDID := accountDoc.ID

	// 初始化存储：
	dirTable, _ := ioutil.TempDir("", "did.table")
	l := log.NewWithModule("account-did")
	sTable, _ := leveldb.New(dirTable)

	// 构建一个 AccountDIDRegistry 实例（ExternalDocDB模式，无需WithAccountDocStorage）：
	ar, _ := bitxid.NewAccountDIDRegistry(sTable, l,
		bitxid.WithGenesisAccountDocInfo(
			bitxid.DocInfo{ID: adminDID, Addr: adminDocAddr, Hash: adminDocHash[:]},
		),
	)

	// 初始化 AccountDIDRegistry
	_ = ar.SetupGenesis()

	// 注册 Account DID：
	docAddr := fakeStore(&accountDoc) // 假设将Doc进行了存储，返回了存储地址
	docBytes, _ = accountDoc.Marshal()
	docHash := fakeHash(docBytes) // 假设将Doc进行了哈希，返回了哈希结果
	ar.Register(accountDID, docAddr, docHash[:])

	// 更新 Account DID：
	accountDoc.Updated = uint64(1616986228)
	docAddr = fakeStore(&accountDoc) // 假设将Doc进行了存储，返回了存储地址
	docBytes, _ = bitxid.Marshal(accountDoc)
	docHash = fakeHash(docBytes) // 假设将Doc进行了哈希，返回了哈希结果
	ar.Update(accountDID, docAddr, docHash[:])

	// 冻结 Account DID：
	ar.Freeze(accountDID)

	// 解冻 Account DID：
	ar.UnFreeze(accountDID)

	// 解析 Account DID：
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

	// 删除 Account DID：
	ar.Delete(accountDID)
	fakeDeleteDoc(accountDID) // 假设此处去删除链下存储的文档

	// 清除存储
	ar.Table.Close()
	os.RemoveAll(dirTable)
}

func fakeHash(docBytes []byte) [32]byte {
	return sha256.Sum256(docBytes)
}

func fakeStore(d bitxid.Doc) string {
	return "/storage/address/to/doc/" + fmt.Sprint(time.Now().Second())
}

func fakeGetDoc(docAddr string) bitxid.AccountDoc {
	return bitxid.AccountDoc{
		BasicDoc: bitxid.BasicDoc{
			ID:      bitxid.DID("did:bitxhub:appchain001:0x12345678"),
			Type:    int(bitxid.AccountDIDType),
			Created: 1616985209,
			Updated: 1616986228,
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
