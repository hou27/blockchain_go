# Scenario

| NODE_ID |            3000(Full Node)            |                               3001                                |           3002(Mine Node)            |
| :-----: | :-----------------------------------: | :---------------------------------------------------------------: | :----------------------------------: |
|    1    |           createblockchain            |                         createblockchain                          |           createblockchain           |
|    2    |             send 3 coins              |                                 -                                 |                  -                   |
|    3    |               starnode                |                              db copy                              |               db copy                |
|    4    |                   -                   |                                 -                                 |               starnode               |
|    5    |                   -                   |                           send 3 coins                            |             sendVersion              |
|    6    |                   -                   |                          sendTx to 3000                           |                  -                   |
|    7    |      sendInv(tx, tx.ID) to 3002       |                                 -                                 |                  -                   |
|    8    |                   -                   |                                 -                                 |    sendGetData(tx, txID) to 3000     |
|    9    |            sendTx to 3002             |                                 -                                 |                  -                   |
|   10    |                   -                   |                                 -                                 | sendInv(block, newblockHash) to 3000 |
|   11    | sendGetData(block, blockHash) to 3002 |                                 -                                 |                  -                   |
|   12    |                   -                   |                                 -                                 |      sendBlock(&block) to 3000       |
|   13    |                   -                   |                             startnode                             |                  -                   |
|   14    |                   -                   |                            sendVersion                            |                  -                   |
|   15    |          sendVersion to 3001          |                                 -                                 |                  -                   |
|   16    |                   -                   |                       sendGetBlocks to 3000                       |                  -                   |
|   17    | sendInv(blocks, block hashes) to 3001 |                                 -                                 |                  -                   |
|   18    |                   -                   | sendGetData(blocks, blockHash) to 3000<br>( × len(payload.Items)) |                  -                   |
|   19    |       sendBlock(&block) to 3001       |                                 -                                 |                  -                   |
