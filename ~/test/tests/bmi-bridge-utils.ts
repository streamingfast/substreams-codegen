import { newMockEvent } from "matchstick-as"
import { ethereum, BigInt, Address } from "@graphprotocol/graph-ts"
import {
  BMIWithdrawn,
  OwnershipTransferred,
  StakedBMI
} from "../generated/BMIBridge/BMIBridge"

export function createBMIWithdrawnEvent(
  amountBMI: BigInt,
  recipient: Address
): BMIWithdrawn {
  let bmiWithdrawnEvent = changetype<BMIWithdrawn>(newMockEvent())

  bmiWithdrawnEvent.parameters = new Array()

  bmiWithdrawnEvent.parameters.push(
    new ethereum.EventParam(
      "amountBMI",
      ethereum.Value.fromUnsignedBigInt(amountBMI)
    )
  )
  bmiWithdrawnEvent.parameters.push(
    new ethereum.EventParam("recipient", ethereum.Value.fromAddress(recipient))
  )

  return bmiWithdrawnEvent
}

export function createOwnershipTransferredEvent(
  previousOwner: Address,
  newOwner: Address
): OwnershipTransferred {
  let ownershipTransferredEvent =
    changetype<OwnershipTransferred>(newMockEvent())

  ownershipTransferredEvent.parameters = new Array()

  ownershipTransferredEvent.parameters.push(
    new ethereum.EventParam(
      "previousOwner",
      ethereum.Value.fromAddress(previousOwner)
    )
  )
  ownershipTransferredEvent.parameters.push(
    new ethereum.EventParam("newOwner", ethereum.Value.fromAddress(newOwner))
  )

  return ownershipTransferredEvent
}

export function createStakedBMIEvent(
  amountBMI: BigInt,
  sender: Address
): StakedBMI {
  let stakedBmiEvent = changetype<StakedBMI>(newMockEvent())

  stakedBmiEvent.parameters = new Array()

  stakedBmiEvent.parameters.push(
    new ethereum.EventParam(
      "amountBMI",
      ethereum.Value.fromUnsignedBigInt(amountBMI)
    )
  )
  stakedBmiEvent.parameters.push(
    new ethereum.EventParam("sender", ethereum.Value.fromAddress(sender))
  )

  return stakedBmiEvent
}
