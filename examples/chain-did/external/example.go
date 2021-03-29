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
	dirTable, _ := ioutil.TempDir("../../../testdata", "did.table")
	l := log.NewWithModule("chain-did")
	sTable, _ := leveldb.New(dirTable)

	// 构建一个 ChainDIDRegistry 实例（ExternalDocDB模式，无需WithChainDocStorage）：
	mr, _ := bitxid.NewChainDIDRegistry(sTable, l,
		bitxid.WithAdmin(adminDID),
		bitxid.WithGenesisChainDoc(
			bitxid.DocOption{ID: relaychainDID, Addr: relaychainDocAddr, Hash: relaychainDocHash[:]},
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
