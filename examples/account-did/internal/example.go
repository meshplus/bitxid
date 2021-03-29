package main

import (
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
			Authentication: []bitxid.Auth{{PublicKey: []string{"KEY#1", "KEY#2", "KEY#3", "KEY#4"}, Strategy: "1-of-4"}},
		},
	}
	adminDID := adminDoc.ID

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
	dirTable, _ := ioutil.TempDir("../../../testdata", "did.table")
	dirDocdb, _ := ioutil.TempDir("../../../testdata", "did.docdb")
	l := log.NewWithModule("did")
	sTable, _ := leveldb.New(dirTable)
	sDocdb, _ := leveldb.New(dirDocdb)

	// 构建一个 AccountDIDRegistry 实例（InternalDocDB模式）：
	ar, _ := bitxid.NewAccountDIDRegistry(sTable, l,
		bitxid.WithAccountDocStorage(sDocdb),
		bitxid.WithDIDAdmin(adminDID),
		bitxid.WithGenesisAccountDoc(
			bitxid.DocOption{Content: &adminDoc},
		),
	)

	// 初始化 AccountDIDRegistry
	_ = ar.SetupGenesis()

	// 注册 Account DID
	ar.RegisterWithDoc(&accountDoc)

	// 更新 Account DID
	accountDoc.Updated = uint64(1616986228)
	ar.UpdateWithDoc(&accountDoc)

	// 冻结 Account DID
	ar.Freeze(accountDID)

	// 解冻 Account DID
	ar.UnFreeze(accountDID)

	// 解析 Account DID
	item, docGet, _, _ := ar.Resolve(accountDID)
	fmt.Println(item)
	fmt.Println(docGet)
	interchain_tx := fakeReceiveInterchainTX()
	if !fakeVerify(docGet.Authentication, interchain_tx) {
		return // 交易合法性验证失败，说明交易发起者不符合权限要求
	}

	// 删除 Account DID
	ar.Delete(accountDID)

	// 清除存储
	ar.Table.Close()
	ar.Docdb.Close()
	os.RemoveAll(dirTable)
	os.RemoveAll(dirDocdb)
}

func fakeReceiveInterchainTX() *pb.Transaction {
	return &pb.Transaction{}
}

func fakeVerify(a []bitxid.Auth, tx *pb.Transaction) bool {
	// 此处理论上只需要成功匹配一种验证方式就可以返回true
	return true
}
