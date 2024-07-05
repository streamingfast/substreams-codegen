import { BigInt, log } from "@graphprotocol/graph-ts";
import { Protobuf } from "as-proto/assembly";
import { Transaction } from "../generated/schema";
import { TransactionList } from "./pb/sf/substreams/cosmos/v1/TransactionList";

export function handleTransactions(bytes: Uint8Array): void {
  const transactionList: TransactionList = Protobuf.decode<TransactionList>(
    bytes,
    TransactionList.decode
  );
  const transactions = transactionList.transactions;

  log.info("Protobuf decoded, length: {}", [transactions.length.toString()]);

  for (let i = 0; i < transactions.length; i++) {
    const transaction = transactions[i];
    if (transaction == null) {
      continue;
    }

    const trxID = transaction.hash;

    let entity = new Transaction(trxID); // need to set an id
    entity.id = trxID;
    entity.resultCode = transaction.resultCode;
    entity.resultData = toHex(transaction.resultData);
    entity.resultLog = transaction.resultLog;
    entity.resultInfo = transaction.resultInfo;
    entity.resultGasWanted = BigInt.fromI64(transaction.resultGasWanted);
    entity.resultGasUsed = BigInt.fromI64(transaction.resultGasUsed);

    let signatures: string[] = [];
    for (let j = 0; j < transaction.signatures.length; j++) {
      signatures.push(toHex(transaction.signatures[j]));
    }

    entity.signatures = signatures;
    entity.save();
    log.debug("Entity saved: {}", [entity.id]);
  }
}

function toHex(bytes: Uint8Array): string {
  let hexString = "";
  for (let i = 0; i < bytes.length; i++) {
    hexString += ("00" + bytes[i].toString(16)).slice(-2);
  }
  return hexString;
}
