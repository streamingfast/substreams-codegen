import {
  assert,
  describe,
  test,
  clearStore,
  beforeAll,
  afterAll
} from "matchstick-as/assembly/index"
import { BigInt, Address } from "@graphprotocol/graph-ts"
import { BMIWithdrawn } from "../generated/schema"
import { BMIWithdrawn as BMIWithdrawnEvent } from "../generated/BMIBridge/BMIBridge"
import { handleBMIWithdrawn } from "../src/bmi-bridge"
import { createBMIWithdrawnEvent } from "./bmi-bridge-utils"

// Tests structure (matchstick-as >=0.5.0)
// https://thegraph.com/docs/en/subgraphs/developing/creating/unit-testing-framework/#tests-structure

describe("Describe entity assertions", () => {
  beforeAll(() => {
    let amountBMI = BigInt.fromI32(234)
    let recipient = Address.fromString(
      "0x0000000000000000000000000000000000000001"
    )
    let newBMIWithdrawnEvent = createBMIWithdrawnEvent(amountBMI, recipient)
    handleBMIWithdrawn(newBMIWithdrawnEvent)
  })

  afterAll(() => {
    clearStore()
  })

  // For more test scenarios, see:
  // https://thegraph.com/docs/en/subgraphs/developing/creating/unit-testing-framework/#write-a-unit-test

  test("BMIWithdrawn created and stored", () => {
    assert.entityCount("BMIWithdrawn", 1)

    // 0xa16081f360e3847006db660bae1c6d1b2e17ec2a is the default address used in newMockEvent() function
    assert.fieldEquals(
      "BMIWithdrawn",
      "0xa16081f360e3847006db660bae1c6d1b2e17ec2a-1",
      "amountBMI",
      "234"
    )
    assert.fieldEquals(
      "BMIWithdrawn",
      "0xa16081f360e3847006db660bae1c6d1b2e17ec2a-1",
      "recipient",
      "0x0000000000000000000000000000000000000001"
    )

    // More assert options:
    // https://thegraph.com/docs/en/subgraphs/developing/creating/unit-testing-framework/#asserts
  })
})
