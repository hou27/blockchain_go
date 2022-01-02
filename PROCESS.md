# Scenario

| NOED_ID |            3000(Full Node)            |       3001       |           3002(Mine Node)            |
| :-----: | :-----------------------------------: | :--------------: | :----------------------------------: |
|    1    |           createblockchain            | createblockchain |           createblockchain           |
|    2    |             send 3 coins              |        -         |                  -                   |
|    3    |               starnode                |     db copy      |               db copy                |
|    4    |                   -                   |        -         |               starnode               |
|    5    |                   -                   |   send 3 coins   |             sendVersion              |
|    6    |                   -                   |  sendTx to 3000  |                  -                   |
|    7    |      sendInv(tx, tx.ID) to 3002       |        -         |                  -                   |
|    8    |                   -                   |        -         |    sendGetData(tx, txID) to 3000     |
|    9    |            sendTx to 3002             |        -         |                  -                   |
|   10    |                   -                   |        -         | sendInv(block, newblockHash) to 3000 |
|   11    | sendGetData(block, blockHash) to 3002 |        -         |                  -                   |
