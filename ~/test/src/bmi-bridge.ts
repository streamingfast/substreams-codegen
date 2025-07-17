import {
  BMIWithdrawn as BMIWithdrawnEvent,
  OwnershipTransferred as OwnershipTransferredEvent,
  StakedBMI as StakedBMIEvent
} from "../generated/BMIBridge/BMIBridge"
import {
  BMIWithdrawn,
  OwnershipTransferred,
  StakedBMI
} from "../generated/schema"

export function handleBMIWithdrawn(event: BMIWithdrawnEvent): void {
  let entity = new BMIWithdrawn(
    event.transaction.hash.concatI32(event.logIndex.toI32())
  )
  entity.amountBMI = event.params.amountBMI
  entity.recipient = event.params.recipient

  entity.blockNumber = event.block.number
  entity.blockTimestamp = event.block.timestamp
  entity.transactionHash = event.transaction.hash

  entity.save()
}

export function handleOwnershipTransferred(
  event: OwnershipTransferredEvent
): void {
  let entity = new OwnershipTransferred(
    event.transaction.hash.concatI32(event.logIndex.toI32())
  )
  entity.previousOwner = event.params.previousOwner
  entity.newOwner = event.params.newOwner

  entity.blockNumber = event.block.number
  entity.blockTimestamp = event.block.timestamp
  entity.transactionHash = event.transaction.hash

  entity.save()
}

export function handleStakedBMI(event: StakedBMIEvent): void {
  let entity = new StakedBMI(
    event.transaction.hash.concatI32(event.logIndex.toI32())
  )
  entity.amountBMI = event.params.amountBMI
  entity.sender = event.params.sender

  entity.blockNumber = event.block.number
  entity.blockTimestamp = event.block.timestamp
  entity.transactionHash = event.transaction.hash

  entity.save()
}
