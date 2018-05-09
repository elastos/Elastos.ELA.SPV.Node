# SPV node
This project is to implement an ELA node like program base on ELA.SPV SDK,
it provides the same RPC interfaces as the ELA full node program like `getblock` `gettransaction` etc,
and several extra interfaces `registeraddresses` `registeraddress`. With a SPV node, you can do almost
the same thing as an ELA full node through JSON-RPC interaction, with reduced data size and less computing resource.

There are several differences between an SPV node and an ELA full node.

- Reduced data size. SPV node only store transactions according to the `addresses` you registered to the SPV node.
- Less computing resource needed. SPV node only verify proof of work of blocks and do not verify transactions. However, an
ELA full node will do those verifications.
- No transaction pool and mining. SPV node do not receive unpacked transactions, and do not do block mining.
- Addresses registration are needed. SPV node use a transaction filter(in our implementation a bloom filter)
to filter corresponding transactions with the registered addresses.

> Address means the ELA account address in string format like `ENTogr92671PKrMmtWo3RLiYXfBTXUe13Z`, this address is converted from
an Uint168 which is used in the transaction output. So the transaction filter can filter corresponding transactions by go through it's outputs.

> Outpoint is a data structure include a transaction ID and output index. It indicates the reference of an transaction output.
If an address ever received an transaction output, there will be the outpoint reference to it. Any time you want to spend the
balance of an address, you must provide the reference of the balance which is an outpoint in the transaction input.
That means the transaction filter will find a transaction is corresponding with an address by go through it's inputs.

## JSON-RPC interfaces
SPV node following the RPC protocol standard.

### Request
A request including `id`, `jsonrpc`, `method` and `params` 4 parameters. `id` and `jsonrpc` is optional, interfaces can work without them.
```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getblock",
    "params":["5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf",2]
}
```
or
```json
{
    "method":"getblock",
    "params":["5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf",2]
}
```

### Response
A response including `id`, `jsonrpc`, `result` and `error` 4 parameters. Whether `id` or `jsonrpc` will returned is according to the request.
```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "result":"919369c9cc8ae901c8b4441b97852a9e9ff5f26570691f2f122c885e5b9ab886"
}
```
or
```json
{
    "result": "919369c9cc8ae901c8b4441b97852a9e9ff5f26570691f2f122c885e5b9ab886"
}
```
error
```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "error": {
        "code": -32603,
        "message": "internal error: header hash not exist on height 10000"
    }
}
```

### RegisterAddresses
On SPV node start, it will not start synchronize process until a `registeraddresses` message received.
This is an initialization message to tall the SPV node all the addresses you are interested, and then SPV node can get all
transactions corresponding with those addresses. This method can only call once on SPV node startup, when a SPV node was started,
use `registeraddress` instead to register new addresses.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"registeraddresses",
    "params":[["Ef2bDPwcUKguteJutJQCmjX2wgHVfkJ2Wq","ENTogr92671PKrMmtWo3RLiYXfBTXUe13Z","ETBBrgotZy3993o9bH75KxjLDgQxBCib6u","EUyNwnAh5SzzTtAPV1HkXzjUEbw2YqKsUM","EYUsEASwbPq9NcSswa8TsP7eVRwiiGwmdq"]]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "error": {
        "code": -32603,
        "message": "internal error: [RegisterAddresses] register addresses failed RegisterAddresses can only call once on SPV node start, please use RegisterAddress to register new addresses"
    }
}
```

### RegisterAddress
When SPV node is running, and you want to register a new address that you are interested, you will use this `registeraddresss` method.
NOTICE: if an address was created before and have historical transactions, it must be registered in the `registeraddresses` method on SPV node first startup, or the transactions corresponding
to this address may not be synchronized.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"registeraddress",
    "params":["Ef2bDPwcUKguteJutJQCmjX2wgHVfkJ2Wq"]
}
```

> Response

```json
{
    "id": 88888,
    "jsonrpc": "2.0",
    "error": {
        "code": -32603,
        "message": "internal error: [RegisterAddress] register address Ef2bDPwcUKguteJutJQCmjX2wgHVfkJ2Wq error address has already registered"
    }
}
```

### GetBlockCount
This `getblockcount` method is the same as it in the BTC RPC interfaces.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getblockcount"
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": 1203
}
```

### GetBlockHash
This `getblockhash` method is the same as it in the BTC RPC interfaces.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getblockhash",
    "params":[100]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": "ea68d4ca267b34ea57db9dadd894896a38107f91e58286269f65a7e6844bd6b7"
}
```

### GetBestBlockHash
This `getbestblockhash` method is the same as it in the BTC RPC interfaces.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getbestblockhash"
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": "5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf"
}
```

### GetBlock
Mostly this method is the same as BTC PRC protocol, query a block with it's hash, but BTC support 3 formats, `0` is for serialized block, `1` is trimmed block info in json decode format,
 `2` is verbose block info in json decode format. By default, `getblock` method return the trimmed block info in json decode format.
SPV node do not support format `0` because the ELA block header is different from a BTC block header, that makes the block hash will be different
event the values in both block header are the same.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getblock",
    "params":["5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf"]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": {
        "hash": "5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf",
        "confirmations": 1,
        "strippedsize": 0,
        "size": 0,
        "weight": 0,
        "height": 1203,
        "version": 0,
        "versionhex": "00000000",
        "merkleroot": "f2d21d7ea4e4146d91495b2cc02a091af42f3dad1d57345e872230dfa5350d78",
        "tx": [
            "f2d21d7ea4e4146d91495b2cc02a091af42f3dad1d57345e872230dfa5350d78"
        ],
        "time": 1525855206,
        "mediantime": 1525855206,
        "nonce": 0,
        "bits": 545259519,
        "difficulty": "",
        "chainwork": "00000000",
        "previousblockhash": "e25e9074cc9c942382065c93a6b3eebad75de413ef623a942b37d9180ea9471a",
        "nextblockhash": "0000000000000000000000000000000000000000000000000000000000000000",
        "auxpow": "01000000010000000000000000000000000000000000000000000000000000000000000000000000002cfabe6d6d5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf0100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffff7f0000000000000000000000000000000000000000000000000000000000000000142b325b5d8b308d099472d244285f718fda3a4b23e927b76bd73a8d11ff498c42b3f25a0000000000000000"
    }
}
```

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getblock",
    "params":["5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf",2]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": {
        "hash": "5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf",
        "confirmations": 1,
        "strippedsize": 0,
        "size": 0,
        "weight": 0,
        "height": 1203,
        "version": 0,
        "versionhex": "00000000",
        "merkleroot": "f2d21d7ea4e4146d91495b2cc02a091af42f3dad1d57345e872230dfa5350d78",
        "tx": [
            {
                "txid": "f2d21d7ea4e4146d91495b2cc02a091af42f3dad1d57345e872230dfa5350d78",
                "hash": "f2d21d7ea4e4146d91495b2cc02a091af42f3dad1d57345e872230dfa5350d78",
                "size": 192,
                "vsize": 192,
                "version": 0,
                "locktime": 1203,
                "vin": [
                    {
                        "txid": "0000000000000000000000000000000000000000000000000000000000000000",
                        "vout": 65535,
                        "sequence": 4294967295
                    }
                ],
                "vout": [
                    {
                        "value": "0.01255707",
                        "n": 0,
                        "address": "8VYXVxKKSAxkmRrfmGpQR2Kc66XhG6m3ta",
                        "assetid": "b037db964a231458d2d6ffd5ea18944c4f90e63d547c5d3b9874df66a4ead0a3",
                        "outputlock": 0
                    },
                    {
                        "value": "0.02929985",
                        "n": 1,
                        "address": "ENTogr92671PKrMmtWo3RLiYXfBTXUe13Z",
                        "assetid": "b037db964a231458d2d6ffd5ea18944c4f90e63d547c5d3b9874df66a4ead0a3",
                        "outputlock": 0
                    }
                ],
                "blockhash": "5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf",
                "confirmations": 1,
                "time": 1525855206,
                "blocktime": 1525855206,
                "type": 0,
                "payloadversion": 4,
                "attributes": [
                    {
                        "usage": 0,
                        "data": "5dcc32c5d6959a73"
                    }
                ]
            }
        ],
        "time": 1525855206,
        "mediantime": 1525855206,
        "nonce": 0,
        "bits": 545259519,
        "difficulty": "",
        "chainwork": "00000000",
        "previousblockhash": "e25e9074cc9c942382065c93a6b3eebad75de413ef623a942b37d9180ea9471a",
        "nextblockhash": "0000000000000000000000000000000000000000000000000000000000000000",
        "auxpow": "01000000010000000000000000000000000000000000000000000000000000000000000000000000002cfabe6d6d5f4d138e9318d1e25600d5141628bf288acacd53e78e0ac9976938bfc69088cf0100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffff7f0000000000000000000000000000000000000000000000000000000000000000142b325b5d8b308d099472d244285f718fda3a4b23e927b76bd73a8d11ff498c42b3f25a0000000000000000"
    }
}
```

### GetBlockByHeight
This is an extend method which not provided in BTC RPC interfaces, this method returns the same value as `getblock` method.
The difference between `getblockbyheight` and `getblock` is `getblockbyheight` use height of a block to query the block info and `getblock` use block hash.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getblockbyheight",
    "params":[100,1]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": {
        "hash": "ea68d4ca267b34ea57db9dadd894896a38107f91e58286269f65a7e6844bd6b7",
        "confirmations": 1104,
        "strippedsize": 0,
        "size": 0,
        "weight": 0,
        "height": 100,
        "version": 0,
        "versionhex": "00000000",
        "merkleroot": "4cbfe9a000475cedd71c79b94c881bd77198a0ffd5b0c2262922b2cf1a41bb55",
        "tx": [
            "4cbfe9a000475cedd71c79b94c881bd77198a0ffd5b0c2262922b2cf1a41bb55"
        ],
        "time": 1525834514,
        "mediantime": 1525834514,
        "nonce": 0,
        "bits": 545259519,
        "difficulty": "",
        "chainwork": "0000044f",
        "previousblockhash": "fc85b26fd91ad907d9ed449ebf412fd4528e20505c1a16c1d1a1b2983de6a8ba",
        "nextblockhash": "e7791a5886e5ce8478c01d7cfb2a0232cc29a328d6f4a7c550ad24594bc5f0a0",
        "auxpow": "01000000010000000000000000000000000000000000000000000000000000000000000000000000002cfabe6d6dea68d4ca267b34ea57db9dadd894896a38107f91e58286269f65a7e6844bd6b70100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffff7f000000000000000000000000000000000000000000000000000000000000000023d668d9390863fb85616b815090636a2652b98455d6acf74cf30fa8a91c758f0263f25a0000000000000000"
    }
}
```

### GetRawTransaction
This method is to query a transaction with it's hash, as the same in BTC protocol. As an extend, SPV node support 3 formats of return value `btc` `ela` and `json`,
you can specify the format in the request. And also this method support the same boolean parameter in BTC protocol to specify return value format.
By default, this method will return a transaction in BTC serialized format.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getrawtransaction",
    "params":["4cbfe9a000475cedd71c79b94c881bd77198a0ffd5b0c2262922b2cf1a41bb55",true]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": {
        "txid": "4cbfe9a000475cedd71c79b94c881bd77198a0ffd5b0c2262922b2cf1a41bb55",
        "hash": "4cbfe9a000475cedd71c79b94c881bd77198a0ffd5b0c2262922b2cf1a41bb55",
        "size": 192,
        "vsize": 192,
        "version": 0,
        "locktime": 100,
        "vin": [
            {
                "txid": "0000000000000000000000000000000000000000000000000000000000000000",
                "vout": 65535,
                "sequence": 4294967295
            }
        ],
        "vout": [
            {
                "value": "0.01255707",
                "n": 0,
                "address": "8VYXVxKKSAxkmRrfmGpQR2Kc66XhG6m3ta",
                "assetid": "b037db964a231458d2d6ffd5ea18944c4f90e63d547c5d3b9874df66a4ead0a3",
                "outputlock": 0
            },
            {
                "value": "0.02929985",
                "n": 1,
                "address": "ENTogr92671PKrMmtWo3RLiYXfBTXUe13Z",
                "assetid": "b037db964a231458d2d6ffd5ea18944c4f90e63d547c5d3b9874df66a4ead0a3",
                "outputlock": 0
            }
        ],
        "blockhash": "ea68d4ca267b34ea57db9dadd894896a38107f91e58286269f65a7e6844bd6b7",
        "confirmations": 1104,
        "time": 1525834514,
        "blocktime": 1525834514,
        "type": 0,
        "payloadversion": 4,
        "attributes": [
            {
                "usage": 0,
                "data": "b7ab3641fe3dea79"
            }
        ]
    }
}
```

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"getrawtransaction",
    "params":["4cbfe9a000475cedd71c79b94c881bd77198a0ffd5b0c2262922b2cf1a41bb55","btc"]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": "00000000010000000000000000000000000000000000000000000000000000000000000000ffff000000ffffffff021b2913000000000015129e9cf1c5f336fcf3a6c954444ed482c5d916e50641b52c000000000015213a3b4511636bf45a582a02b2ee0a0d3c9c52dfe164000000"
}
```

### SendRawTransaction
This method is to send a transaction into the P2P network, and support both `btc` and `ela` transaction.
The first parameter is the transaction in hex string format, the second parameter is the transaction format `btc` or `ela`.
By default, `sendrawtransaction` treat the received data as `btc` format.

> Request

```json
{
    "id":123456,
    "jsonrpc":"2.0",
    "method":"sendrawtransaction",
    "params":["02000100133535373730303637393139343737373934313001fbce23c9a879c865a8f9f6a5f4a8f21894c191197122b4966bda9abee681771200000000000001b037db964a231458d2d6ffd5ea18944c4f90e63d547c5d3b9874df66a4ead0a300e1f505000000000000000021a81fe609252821249a6d648500e818bb6b3bd7e0b3040000014140ae416b4a4b20e39aa5fdd6d43b9adbbf41f6b5f44c8d22c87835580115edce4c68f080a507098cc878ca2ce0ff52e314519c38318133d769fcffe7d263f18696232102fcc4423da8bb717419c0f193a22d0fb03a1773344f01a0bdd4cdf8dc2c18bf33ac","ela"]
}
```

> Response

```json
{
    "id": 123456,
    "jsonrpc": "2.0",
    "result": "132ec7f354bb539200d13c596741083effff79d10621dac33a7f071c88e478ca"
}
```
