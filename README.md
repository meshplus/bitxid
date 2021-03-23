# BitXID

BitXID is an DID framework comply with W3C DID(Decentralized Identifiers) specifications. It has the following features:

1. **Multiple storage management**: BitXID offers both on-chain storage and off-chain storage for DID storage. The best practice will be store small amounts of data(e.g. id, status, etc.) on-chain while store large amounts of data(e.g. public keys, authentication methods, etc.) off-chain(e.g. IPFS), and combines them by store hash of the data on-chain.
2. **Multiple method management**: not only can BitXID be used to build digital identity for a blockchain but also it can be to build digital identity ecosystem for a blockchain network(i.e. cross-chain platform).

## Installation

Install `bitxid` package:

```shell
go get -u github.com/bitxhub/bitxid
```

import it in your code:

```go
import "github.com/bitxhub/bitxid"
```

## Example

BitXID is already used by several great projects, among which [BitXHub](https://github.com/meshplus/bitxhub) is one of them. BitXHub DID has already registered on [DIF Universal Resolver](https://github.com/decentralized-identity/universal-resolver). You can find BitXHub DID Implementation [here](https://github.com/bitxhub/did-method-registry) and the design of BitXHub DID in the latest [BitXHub WhitePaper](https://upload.hyperchain.cn/BitXHub白皮书.pdf).

## Usage

Usage guide can be found in [here](./docs/usage-chn.md).

## Contact

Email: bitxhub@hyperchain.cn

Wechat: If you‘re interested in BitXID or BitXHub, please add the assistant to join our community group.

<img src="https://raw.githubusercontent.com/meshplus/bitxhub/master/docs/wechat.png" width="200" /><img src="https://raw.githubusercontent.com/meshplus/bitxhub/master/docs/official.png" width="206" />

## License

The BitXID library (i.e. all code outside of the cmd and internal directory) is licensed under the GNU Lesser General Public License v3.0, also included in our repository in the COPYING.LESSER file.

The BitXID binaries (i.e. all code inside of the cmd and internal directory) is licensed under the GNU General Public License v3.0, also included in our repository in the COPYING file.