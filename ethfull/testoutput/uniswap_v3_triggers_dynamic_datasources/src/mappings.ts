import { log } from "@graphprotocol/graph-ts";
import { Protobuf } from "as-proto/assembly";
import { JSON } from "assemblyscript-json";
import { Call, Event } from "../generated/schema";
import { Events } from "./pb/contract/v1/Events";
import { Calls } from "./pb/contract/v1/Calls";
import { EventsCalls } from "./pb/contract/v1/EventsCalls";


export function handleEventsAndCalls(bytes: Uint8Array): void {
  const eventsCalls: EventsCalls = Protobuf.decode<EventsCalls>(
    bytes,
    EventsCalls.decode
  );
  const events: Events | null = eventsCalls.events;
  if (events === null) {
    return;
  }
  const calls: Calls | null = eventsCalls.calls;
  if (calls === null) {
    return;
  }

  // Below you will find examples of how to save the decoded events.
  // These are only examples, you can modify them to suit your needs.
  for (let i = 0; i < events.factoryFeeAmountEnableds.length; i++) {
    const e = events.factoryFeeAmountEnableds[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("fee", e.fee);
    obj.set("tickSpacing", e.tickSpacing);
    evt.jsonValue = obj.toString();
    evt.type = "feeAmountEnabled";
    evt.save();
  }
  
  for (let i = 0; i < events.factoryOwnerChangeds.length; i++) {
    const e = events.factoryOwnerChangeds[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("newOwner", e.newOwner);
    obj.set("oldOwner", e.oldOwner);
    evt.jsonValue = obj.toString();
    evt.type = "ownerChanged";
    evt.save();
  }
  
  for (let i = 0; i < events.factoryPoolCreateds.length; i++) {
    const e = events.factoryPoolCreateds[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("fee", e.fee);
    obj.set("pool", e.pool);
    obj.set("tickSpacing", e.tickSpacing);
    obj.set("token0", e.token0);
    obj.set("token1", e.token1);
    evt.jsonValue = obj.toString();
    evt.type = "poolCreated";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsBurns.length; i++) {
    const e = events.poolsBurns[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("amount", e.amount);
    obj.set("amount0", e.amount0);
    obj.set("amount1", e.amount1);
    obj.set("owner", e.owner);
    obj.set("tickLower", e.tickLower);
    obj.set("tickUpper", e.tickUpper);
    evt.jsonValue = obj.toString();
    evt.type = "burn";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsCollects.length; i++) {
    const e = events.poolsCollects[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("amount0", e.amount0);
    obj.set("amount1", e.amount1);
    obj.set("owner", e.owner);
    obj.set("recipient", e.recipient);
    obj.set("tickLower", e.tickLower);
    obj.set("tickUpper", e.tickUpper);
    evt.jsonValue = obj.toString();
    evt.type = "collect";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsCollectProtocols.length; i++) {
    const e = events.poolsCollectProtocols[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("amount0", e.amount0);
    obj.set("amount1", e.amount1);
    obj.set("recipient", e.recipient);
    obj.set("sender", e.sender);
    evt.jsonValue = obj.toString();
    evt.type = "collectProtocol";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsFlashes.length; i++) {
    const e = events.poolsFlashes[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("amount0", e.amount0);
    obj.set("amount1", e.amount1);
    obj.set("paid0", e.paid0);
    obj.set("paid1", e.paid1);
    obj.set("recipient", e.recipient);
    obj.set("sender", e.sender);
    evt.jsonValue = obj.toString();
    evt.type = "flash";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsIncreaseObservationCardinalityNexts.length; i++) {
    const e = events.poolsIncreaseObservationCardinalityNexts[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("observationCardinalityNextNew", e.observationCardinalityNextNew);
    obj.set("observationCardinalityNextOld", e.observationCardinalityNextOld);
    evt.jsonValue = obj.toString();
    evt.type = "increaseObservationCardinalityNext";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsInitializes.length; i++) {
    const e = events.poolsInitializes[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("sqrtPriceX96", e.sqrtPriceX96);
    obj.set("tick", e.tick);
    evt.jsonValue = obj.toString();
    evt.type = "initialize";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsMints.length; i++) {
    const e = events.poolsMints[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("amount", e.amount);
    obj.set("amount0", e.amount0);
    obj.set("amount1", e.amount1);
    obj.set("owner", e.owner);
    obj.set("sender", e.sender);
    obj.set("tickLower", e.tickLower);
    obj.set("tickUpper", e.tickUpper);
    evt.jsonValue = obj.toString();
    evt.type = "mint";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsSetFeeProtocols.length; i++) {
    const e = events.poolsSetFeeProtocols[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("feeProtocol0New", e.feeProtocol0New);
    obj.set("feeProtocol0Old", e.feeProtocol0Old);
    obj.set("feeProtocol1New", e.feeProtocol1New);
    obj.set("feeProtocol1Old", e.feeProtocol1Old);
    evt.jsonValue = obj.toString();
    evt.type = "setFeeProtocol";
    evt.save();
  }
  
  for (let i = 0; i < events.poolsSwaps.length; i++) {
    const e = events.poolsSwaps[i];
    let evt = new Event(ID(e.evtTxHash, i));
    let obj = new JSON.Obj();
    obj.set("evtTxHash", e.evtTxHash);
    obj.set("evtIndex", e.evtIndex);
    obj.set("evtBlockTime", e.evtBlockTime);
    obj.set("evtBlockNumber", e.evtBlockNumber);
    obj.set("amount0", e.amount0);
    obj.set("amount1", e.amount1);
    obj.set("liquidity", e.liquidity);
    obj.set("recipient", e.recipient);
    obj.set("sender", e.sender);
    obj.set("sqrtPriceX96", e.sqrtPriceX96);
    obj.set("tick", e.tick);
    evt.jsonValue = obj.toString();
    evt.type = "swap";
    evt.save();
  }
  
  // Below you will find examples of how to save the decoded calls.
  // These are only examples, you can modify them to suit your needs.
  for (let i = 0; i < calls.factoryCallCreatePools.length; i++) {
    const c = calls.factoryCallCreatePools[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("fee", c.fee);
    obj.set("outputPool", c.outputPool);
    obj.set("tokenA", c.tokenA);
    obj.set("tokenB", c.tokenB);
    call.jsonValue = obj.toString();
    call.type = "createPool";
    call.save();
  }
  
  for (let i = 0; i < calls.factoryCallEnableFeeAmounts.length; i++) {
    const c = calls.factoryCallEnableFeeAmounts[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("fee", c.fee);
    obj.set("tickSpacing", c.tickSpacing);
    call.jsonValue = obj.toString();
    call.type = "enableFeeAmount";
    call.save();
  }
  
  for (let i = 0; i < calls.factoryCallSetOwners.length; i++) {
    const c = calls.factoryCallSetOwners[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("uOwner", c.uOwner);
    call.jsonValue = obj.toString();
    call.type = "setOwner";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallBurns.length; i++) {
    const c = calls.poolsCallBurns[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("amount", c.amount);
    obj.set("outputAmount0", c.outputAmount0);
    obj.set("outputAmount1", c.outputAmount1);
    obj.set("tickLower", c.tickLower);
    obj.set("tickUpper", c.tickUpper);
    call.jsonValue = obj.toString();
    call.type = "burn";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallCollects.length; i++) {
    const c = calls.poolsCallCollects[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("amount0Requested", c.amount0Requested);
    obj.set("amount1Requested", c.amount1Requested);
    obj.set("outputAmount0", c.outputAmount0);
    obj.set("outputAmount1", c.outputAmount1);
    obj.set("recipient", c.recipient);
    obj.set("tickLower", c.tickLower);
    obj.set("tickUpper", c.tickUpper);
    call.jsonValue = obj.toString();
    call.type = "collect";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallCollectProtocols.length; i++) {
    const c = calls.poolsCallCollectProtocols[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("amount0Requested", c.amount0Requested);
    obj.set("amount1Requested", c.amount1Requested);
    obj.set("outputAmount0", c.outputAmount0);
    obj.set("outputAmount1", c.outputAmount1);
    obj.set("recipient", c.recipient);
    call.jsonValue = obj.toString();
    call.type = "collectProtocol";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallFlashes.length; i++) {
    const c = calls.poolsCallFlashes[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("amount0", c.amount0);
    obj.set("amount1", c.amount1);
    obj.set("data", c.data);
    obj.set("recipient", c.recipient);
    call.jsonValue = obj.toString();
    call.type = "flash";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallIncreaseObservationCardinalityNexts.length; i++) {
    const c = calls.poolsCallIncreaseObservationCardinalityNexts[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("observationCardinalityNext", c.observationCardinalityNext);
    call.jsonValue = obj.toString();
    call.type = "increaseObservationCardinalityNext";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallInitializes.length; i++) {
    const c = calls.poolsCallInitializes[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("sqrtPriceX96", c.sqrtPriceX96);
    call.jsonValue = obj.toString();
    call.type = "initialize";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallMints.length; i++) {
    const c = calls.poolsCallMints[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("amount", c.amount);
    obj.set("data", c.data);
    obj.set("outputAmount0", c.outputAmount0);
    obj.set("outputAmount1", c.outputAmount1);
    obj.set("recipient", c.recipient);
    obj.set("tickLower", c.tickLower);
    obj.set("tickUpper", c.tickUpper);
    call.jsonValue = obj.toString();
    call.type = "mint";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallSetFeeProtocols.length; i++) {
    const c = calls.poolsCallSetFeeProtocols[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("feeProtocol0", c.feeProtocol0);
    obj.set("feeProtocol1", c.feeProtocol1);
    call.jsonValue = obj.toString();
    call.type = "setFeeProtocol";
    call.save();
  }
  
  for (let i = 0; i < calls.poolsCallSwaps.length; i++) {
    const c = calls.poolsCallSwaps[i];
    let call = new Call(ID(c.callTxHash, i));
    let obj = new JSON.Obj();
    obj.set("callTxHash", c.callTxHash);
    obj.set("callBlockTime", c.callBlockTime);
    obj.set("callBlockNumber", c.callBlockNumber);
    obj.set("callOrdinal", c.callOrdinal);
    obj.set("callSuccess", c.callSuccess);
    obj.set("amountSpecified", c.amountSpecified);
    obj.set("data", c.data);
    obj.set("outputAmount0", c.outputAmount0);
    obj.set("outputAmount1", c.outputAmount1);
    obj.set("recipient", c.recipient);
    obj.set("sqrtPriceLimitX96", c.sqrtPriceLimitX96);
    obj.set("zeroForOne", c.zeroForOne);
    call.jsonValue = obj.toString();
    call.type = "swap";
    call.save();
  }
  
}

function ID(trxHash: string, i: u32): string {
  return trxHash + "-" + i.toString();
}
