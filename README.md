# SPV node
This project is to implement an ELA node like program base on ELA.SPV SDK,
it provides the same RPC interfaces as the ELA full node program like `getblock` `gettransaction` etc,
and several extra interfaces `registerdata` `registernewaddress`. With a SPV node, you can do almost
the same thing as an ELA full node through JSON-RPC interaction, with reduced data size and less computing resource.

There are several differences between an SPV node and an ELA full node.

- Reduced data size. SPV node only store transactions according to the `addresses` and `outpoints` you registered to
the SPV node.
- Less computing resource needed. SPV node only verify proof of work of blocks and do not verify transactions. However, an
ELA full node will do those verifications.
- No transaction pool and mining. SPV node do not receive unpacked transactions, and do not do block mining.
- Addresses and Outpoints registration are needed. SPV node use a transaction filter(in our implementation a bloom filter)
to filter corresponding transactions.

> Address means the ELA account address in string format like `ENTogr92671PKrMmtWo3RLiYXfBTXUe13Z`, this address is converted from
an Uint168 which is used in the transaction output. So the transaction filter can filter corresponding transactions by go through it's outputs.

> Outpoint is a data structure include a transaction ID and output index. It indicates the reference of an transaction output.
If an address ever received an transaction output, there will be the outpoint reference to it. Any time you want to spend the
balance of an address, you must provide the reference of the balance which is an outpoint in the transaction input.
That means the transaction filter will find a transaction is corresponding with an address by go through it's inputs.

## License
MIT License

Copyright (c) 2018 Elastos Foundation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.