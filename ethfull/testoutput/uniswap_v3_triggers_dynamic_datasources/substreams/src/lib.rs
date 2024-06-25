mod abi;
mod pb;
use hex_literal::hex;
use pb::contract::v1 as contract;
use substreams::prelude::*;
use substreams::store;
use substreams::Hex;
use substreams_ethereum::pb::eth::v2 as eth;
use substreams_ethereum::Event;

#[allow(unused_imports)]
use num_traits::cast::ToPrimitive;
use std::str::FromStr;
use substreams::scalar::BigDecimal;

substreams_ethereum::init!();

const FACTORY_TRACKED_CONTRACT: [u8; 20] = hex!("1f98431c8ad98523631ae4a59f267346ea31f984");

fn map_factory_events(blk: &eth::Block, events: &mut contract::Events) {
    events.factory_fee_amount_enableds.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| log.address == FACTORY_TRACKED_CONTRACT)
                .filter_map(|log| {
                    if let Some(event) = abi::factory_contract::events::FeeAmountEnabled::match_and_decode(log) {
                        return Some(contract::FactoryFeeAmountEnabled {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            fee: event.fee.to_u64(),
                            tick_spacing: Into::<num_bigint::BigInt>::into(event.tick_spacing).to_i64().unwrap(),
                        });
                    }

                    None
                })
        })
        .collect());
    events.factory_owner_changeds.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| log.address == FACTORY_TRACKED_CONTRACT)
                .filter_map(|log| {
                    if let Some(event) = abi::factory_contract::events::OwnerChanged::match_and_decode(log) {
                        return Some(contract::FactoryOwnerChanged {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            new_owner: event.new_owner,
                            old_owner: event.old_owner,
                        });
                    }

                    None
                })
        })
        .collect());
    events.factory_pool_createds.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| log.address == FACTORY_TRACKED_CONTRACT)
                .filter_map(|log| {
                    if let Some(event) = abi::factory_contract::events::PoolCreated::match_and_decode(log) {
                        return Some(contract::FactoryPoolCreated {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            fee: event.fee.to_u64(),
                            pool: event.pool,
                            tick_spacing: Into::<num_bigint::BigInt>::into(event.tick_spacing).to_i64().unwrap(),
                            token0: event.token0,
                            token1: event.token1,
                        });
                    }

                    None
                })
        })
        .collect());
}
fn map_factory_calls(blk: &eth::Block, calls: &mut contract::Calls) {
    calls.factory_call_create_pools.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == FACTORY_TRACKED_CONTRACT && abi::factory_contract::functions::CreatePool::match_call(call))
                .filter_map(|call| {
                    match abi::factory_contract::functions::CreatePool::decode(call) {
                        Ok(decoded_call) => {
                            let output_pool = match abi::factory_contract::functions::CreatePool::output(&call.return_data) {
                                Ok(output_pool) => {output_pool}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::FactoryCreatePoolCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                fee: decoded_call.fee.to_u64(),
                                output_pool: output_pool,
                                token_a: decoded_call.token_a,
                                token_b: decoded_call.token_b,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.factory_call_enable_fee_amounts.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == FACTORY_TRACKED_CONTRACT && abi::factory_contract::functions::EnableFeeAmount::match_call(call))
                .filter_map(|call| {
                    match abi::factory_contract::functions::EnableFeeAmount::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::FactoryEnableFeeAmountCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                fee: decoded_call.fee.to_u64(),
                                tick_spacing: Into::<num_bigint::BigInt>::into(decoded_call.tick_spacing).to_i64().unwrap(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.factory_call_set_owners.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == FACTORY_TRACKED_CONTRACT && abi::factory_contract::functions::SetOwner::match_call(call))
                .filter_map(|call| {
                    match abi::factory_contract::functions::SetOwner::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::FactorySetOwnerCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                u_owner: decoded_call.u_owner,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
}

#[substreams::handlers::map]
fn zipped_events_calls(
    events: contract::Events,
    calls: contract::Calls,
) -> Result<contract::EventsCalls, substreams::errors::Error> {
    Ok(contract::EventsCalls {
        events: Some(events),
        calls: Some(calls),
    })
}
fn is_declared_dds_address(addr: &Vec<u8>, ordinal: u64, dds_store: &store::StoreGetInt64) -> bool {
    //    substreams::log::info!("Checking if address {} is declared dds address", Hex(addr).to_string());
    if dds_store.get_at(ordinal, Hex(addr).to_string()).is_some() {
        return true;
    }
    return false;
}
fn map_pools_events(
    blk: &eth::Block,
    dds_store: &store::StoreGetInt64,
    events: &mut contract::Events,
) {

    events.pools_burns.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::Burn::match_and_decode(log) {
                        return Some(contract::PoolsBurn {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            amount: event.amount.to_string(),
                            amount0: event.amount0.to_string(),
                            amount1: event.amount1.to_string(),
                            owner: event.owner,
                            tick_lower: Into::<num_bigint::BigInt>::into(event.tick_lower).to_i64().unwrap(),
                            tick_upper: Into::<num_bigint::BigInt>::into(event.tick_upper).to_i64().unwrap(),
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_collects.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::Collect::match_and_decode(log) {
                        return Some(contract::PoolsCollect {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            amount0: event.amount0.to_string(),
                            amount1: event.amount1.to_string(),
                            owner: event.owner,
                            recipient: event.recipient,
                            tick_lower: Into::<num_bigint::BigInt>::into(event.tick_lower).to_i64().unwrap(),
                            tick_upper: Into::<num_bigint::BigInt>::into(event.tick_upper).to_i64().unwrap(),
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_collect_protocols.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::CollectProtocol::match_and_decode(log) {
                        return Some(contract::PoolsCollectProtocol {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            amount0: event.amount0.to_string(),
                            amount1: event.amount1.to_string(),
                            recipient: event.recipient,
                            sender: event.sender,
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_flashes.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::Flash::match_and_decode(log) {
                        return Some(contract::PoolsFlash {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            amount0: event.amount0.to_string(),
                            amount1: event.amount1.to_string(),
                            paid0: event.paid0.to_string(),
                            paid1: event.paid1.to_string(),
                            recipient: event.recipient,
                            sender: event.sender,
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_increase_observation_cardinality_nexts.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::IncreaseObservationCardinalityNext::match_and_decode(log) {
                        return Some(contract::PoolsIncreaseObservationCardinalityNext {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            observation_cardinality_next_new: event.observation_cardinality_next_new.to_u64(),
                            observation_cardinality_next_old: event.observation_cardinality_next_old.to_u64(),
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_initializes.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::Initialize::match_and_decode(log) {
                        return Some(contract::PoolsInitialize {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            sqrt_price_x96: event.sqrt_price_x96.to_string(),
                            tick: Into::<num_bigint::BigInt>::into(event.tick).to_i64().unwrap(),
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_mints.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::Mint::match_and_decode(log) {
                        return Some(contract::PoolsMint {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            amount: event.amount.to_string(),
                            amount0: event.amount0.to_string(),
                            amount1: event.amount1.to_string(),
                            owner: event.owner,
                            sender: event.sender,
                            tick_lower: Into::<num_bigint::BigInt>::into(event.tick_lower).to_i64().unwrap(),
                            tick_upper: Into::<num_bigint::BigInt>::into(event.tick_upper).to_i64().unwrap(),
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_set_fee_protocols.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::SetFeeProtocol::match_and_decode(log) {
                        return Some(contract::PoolsSetFeeProtocol {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            fee_protocol0_new: event.fee_protocol0_new.to_u64(),
                            fee_protocol0_old: event.fee_protocol0_old.to_u64(),
                            fee_protocol1_new: event.fee_protocol1_new.to_u64(),
                            fee_protocol1_old: event.fee_protocol1_old.to_u64(),
                        });
                    }

                    None
                })
        })
        .collect());

    events.pools_swaps.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| is_declared_dds_address(&log.address, log.ordinal, dds_store))
                .filter_map(|log| {
                    if let Some(event) = abi::pools_contract::events::Swap::match_and_decode(log) {
                        return Some(contract::PoolsSwap {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            evt_address: Hex(&log.address).to_string(),
                            amount0: event.amount0.to_string(),
                            amount1: event.amount1.to_string(),
                            liquidity: event.liquidity.to_string(),
                            recipient: event.recipient,
                            sender: event.sender,
                            sqrt_price_x96: event.sqrt_price_x96.to_string(),
                            tick: Into::<num_bigint::BigInt>::into(event.tick).to_i64().unwrap(),
                        });
                    }

                    None
                })
        })
        .collect());
}
fn map_pools_calls(
    blk: &eth::Block,
    dds_store: &store::StoreGetInt64,
    calls: &mut contract::Calls,
) {
    calls.pools_call_burns.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::Burn::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::Burn::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pools_contract::functions::Burn::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolsBurnCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                amount: decoded_call.amount.to_string(),
                                output_amount0: output_amount0.to_string(),
                                output_amount1: output_amount1.to_string(),
                                tick_lower: Into::<num_bigint::BigInt>::into(decoded_call.tick_lower).to_i64().unwrap(),
                                tick_upper: Into::<num_bigint::BigInt>::into(decoded_call.tick_upper).to_i64().unwrap(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_collects.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::Collect::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::Collect::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pools_contract::functions::Collect::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolsCollectCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                amount0_requested: decoded_call.amount0_requested.to_string(),
                                amount1_requested: decoded_call.amount1_requested.to_string(),
                                output_amount0: output_amount0.to_string(),
                                output_amount1: output_amount1.to_string(),
                                recipient: decoded_call.recipient,
                                tick_lower: Into::<num_bigint::BigInt>::into(decoded_call.tick_lower).to_i64().unwrap(),
                                tick_upper: Into::<num_bigint::BigInt>::into(decoded_call.tick_upper).to_i64().unwrap(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_collect_protocols.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::CollectProtocol::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::CollectProtocol::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pools_contract::functions::CollectProtocol::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolsCollectProtocolCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                amount0_requested: decoded_call.amount0_requested.to_string(),
                                amount1_requested: decoded_call.amount1_requested.to_string(),
                                output_amount0: output_amount0.to_string(),
                                output_amount1: output_amount1.to_string(),
                                recipient: decoded_call.recipient,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_flashes.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::Flash::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::Flash::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolsFlashCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                amount0: decoded_call.amount0.to_string(),
                                amount1: decoded_call.amount1.to_string(),
                                data: decoded_call.data,
                                recipient: decoded_call.recipient,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_increase_observation_cardinality_nexts.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::IncreaseObservationCardinalityNext::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::IncreaseObservationCardinalityNext::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolsIncreaseObservationCardinalityNextCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                observation_cardinality_next: decoded_call.observation_cardinality_next.to_u64(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_initializes.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::Initialize::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::Initialize::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolsInitializeCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                sqrt_price_x96: decoded_call.sqrt_price_x96.to_string(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_mints.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::Mint::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::Mint::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pools_contract::functions::Mint::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolsMintCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                amount: decoded_call.amount.to_string(),
                                data: decoded_call.data,
                                output_amount0: output_amount0.to_string(),
                                output_amount1: output_amount1.to_string(),
                                recipient: decoded_call.recipient,
                                tick_lower: Into::<num_bigint::BigInt>::into(decoded_call.tick_lower).to_i64().unwrap(),
                                tick_upper: Into::<num_bigint::BigInt>::into(decoded_call.tick_upper).to_i64().unwrap(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_set_fee_protocols.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::SetFeeProtocol::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::SetFeeProtocol::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolsSetFeeProtocolCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                fee_protocol0: decoded_call.fee_protocol0.to_u64(),
                                fee_protocol1: decoded_call.fee_protocol1.to_u64(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.pools_call_swaps.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pools_contract::functions::Swap::match_call(call))
                .filter_map(|call| {
                    match abi::pools_contract::functions::Swap::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pools_contract::functions::Swap::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolsSwapCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                call_address: Hex(&call.address).to_string(),
                                amount_specified: decoded_call.amount_specified.to_string(),
                                data: decoded_call.data,
                                output_amount0: output_amount0.to_string(),
                                output_amount1: output_amount1.to_string(),
                                recipient: decoded_call.recipient,
                                sqrt_price_limit_x96: decoded_call.sqrt_price_limit_x96.to_string(),
                                zero_for_one: decoded_call.zero_for_one,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
}


#[substreams::handlers::store]
fn store_pools_created(blk: eth::Block, store: StoreSetInt64) {
    for rcpt in blk.receipts() {
        for log in rcpt
            .receipt
            .logs
            .iter()
            .filter(|log| log.address == FACTORY_TRACKED_CONTRACT)
        {
            if let Some(event) = abi::factory_contract::events::PoolCreated::match_and_decode(log) {
                store.set(log.ordinal, Hex(event.pool).to_string(), &1);
            }
        }
    }
}
#[substreams::handlers::map]
fn map_events(
    blk: eth::Block,
    store_pools: StoreGetInt64,
) -> Result<contract::Events, substreams::errors::Error> {
    let mut events = contract::Events::default();
    map_factory_events(&blk, &mut events);
    map_pools_events(&blk, &store_pools, &mut events);
    Ok(events)
}
#[substreams::handlers::map]
fn map_calls(
    blk: eth::Block,
    store_pools: StoreGetInt64,
    
) -> Result<contract::Calls, substreams::errors::Error> {
let mut calls = contract::Calls::default();
    map_factory_calls(&blk, &mut calls);
    map_pools_calls(&blk, &store_pools, &mut calls);
    Ok(calls)
}

