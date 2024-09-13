mod abi;
mod pb;
use hex_literal::hex;
use pb::contract::v1 as contract;
use substreams::prelude::*;
use substreams::store;
use substreams::Hex;
use substreams_database_change::pb::database::DatabaseChanges;
use substreams_database_change::tables::Tables as DatabaseChangeTables;
use substreams_entity_change::pb::entity::EntityChanges;
use substreams_entity_change::tables::Tables as EntityChangesTables;
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

fn map_pool_calls(
    blk: &eth::Block,
    dds_store: &store::StoreGetInt64,
    calls: &mut contract::Calls,
) {
    calls.pool_call_burns.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::Burn::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::Burn::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pool_contract::functions::Burn::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolBurnCall {
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
    calls.pool_call_collects.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::Collect::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::Collect::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pool_contract::functions::Collect::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolCollectCall {
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
    calls.pool_call_collect_protocols.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::CollectProtocol::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::CollectProtocol::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pool_contract::functions::CollectProtocol::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolCollectProtocolCall {
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
    calls.pool_call_flashes.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::Flash::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::Flash::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolFlashCall {
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
    calls.pool_call_increase_observation_cardinality_nexts.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::IncreaseObservationCardinalityNext::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::IncreaseObservationCardinalityNext::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolIncreaseObservationCardinalityNextCall {
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
    calls.pool_call_initializes.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::Initialize::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::Initialize::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolInitializeCall {
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
    calls.pool_call_mints.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::Mint::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::Mint::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pool_contract::functions::Mint::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolMintCall {
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
    calls.pool_call_set_fee_protocols.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::SetFeeProtocol::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::SetFeeProtocol::decode(call) {
                            Ok(decoded_call) => {
                            Some(contract::PoolSetFeeProtocolCall {
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
    calls.pool_call_swaps.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| is_declared_dds_address(&call.address, call.begin_ordinal, dds_store) && abi::pool_contract::functions::Swap::match_call(call))
                .filter_map(|call| {
                    match abi::pool_contract::functions::Swap::decode(call) {
                            Ok(decoded_call) => {
                            let (output_amount0, output_amount1) = match abi::pool_contract::functions::Swap::output(&call.return_data) {
                                Ok((output_amount0, output_amount1)) => {(output_amount0, output_amount1)}
                                Err(_) => Default::default(),
                            };
                            
                            Some(contract::PoolSwapCall {
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


fn db_factory_out(events: &contract::Events, tables: &mut DatabaseChangeTables) {
    // Loop over all the abis events to create table changes
    events.factory_fee_amount_enableds.iter().for_each(|evt| {
        tables
            .create_row("factory_fee_amount_enabled", [("evt_tx_hash", evt.evt_tx_hash.to_string()),("evt_index", evt.evt_index.to_string())])
            .set("evt_block_time", evt.evt_block_time.as_ref().unwrap())
            .set("evt_block_number", evt.evt_block_number)
            .set("fee", evt.fee)
            .set("tick_spacing", evt.tick_spacing);
    });
    events.factory_owner_changeds.iter().for_each(|evt| {
        tables
            .create_row("factory_owner_changed", [("evt_tx_hash", evt.evt_tx_hash.to_string()),("evt_index", evt.evt_index.to_string())])
            .set("evt_block_time", evt.evt_block_time.as_ref().unwrap())
            .set("evt_block_number", evt.evt_block_number)
            .set("new_owner", Hex(&evt.new_owner).to_string())
            .set("old_owner", Hex(&evt.old_owner).to_string());
    });
    events.factory_pool_createds.iter().for_each(|evt| {
        tables
            .create_row("factory_pool_created", [("evt_tx_hash", evt.evt_tx_hash.to_string()),("evt_index", evt.evt_index.to_string())])
            .set("evt_block_time", evt.evt_block_time.as_ref().unwrap())
            .set("evt_block_number", evt.evt_block_number)
            .set("fee", evt.fee)
            .set("pool", Hex(&evt.pool).to_string())
            .set("tick_spacing", evt.tick_spacing)
            .set("token0", Hex(&evt.token0).to_string())
            .set("token1", Hex(&evt.token1).to_string());
    });
}
fn db_pool_calls_out(calls: &contract::Calls, tables: &mut DatabaseChangeTables) {
    // Loop over all the abis calls to create table changes
    calls.pool_call_burns.iter().for_each(|call| {
        tables
            .create_row("pool_call_burn", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount", BigDecimal::from_str(&call.amount).unwrap())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("tick_lower", call.tick_lower)
            .set("tick_upper", call.tick_upper);
    });
    calls.pool_call_collects.iter().for_each(|call| {
        tables
            .create_row("pool_call_collect", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount0_requested", BigDecimal::from_str(&call.amount0_requested).unwrap())
            .set("amount1_requested", BigDecimal::from_str(&call.amount1_requested).unwrap())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string())
            .set("tick_lower", call.tick_lower)
            .set("tick_upper", call.tick_upper);
    });
    calls.pool_call_collect_protocols.iter().for_each(|call| {
        tables
            .create_row("pool_call_collect_protocol", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount0_requested", BigDecimal::from_str(&call.amount0_requested).unwrap())
            .set("amount1_requested", BigDecimal::from_str(&call.amount1_requested).unwrap())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string());
    });
    calls.pool_call_flashes.iter().for_each(|call| {
        tables
            .create_row("pool_call_flash", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount0", BigDecimal::from_str(&call.amount0).unwrap())
            .set("amount1", BigDecimal::from_str(&call.amount1).unwrap())
            .set("data", Hex(&call.data).to_string())
            .set("recipient", Hex(&call.recipient).to_string());
    });
    calls.pool_call_increase_observation_cardinality_nexts.iter().for_each(|call| {
        tables
            .create_row("pool_call_increase_observation_cardinality_next", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("observation_cardinality_next", call.observation_cardinality_next);
    });
    calls.pool_call_initializes.iter().for_each(|call| {
        tables
            .create_row("pool_call_initialize", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("sqrt_price_x96", BigDecimal::from_str(&call.sqrt_price_x96).unwrap());
    });
    calls.pool_call_mints.iter().for_each(|call| {
        tables
            .create_row("pool_call_mint", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount", BigDecimal::from_str(&call.amount).unwrap())
            .set("data", Hex(&call.data).to_string())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string())
            .set("tick_lower", call.tick_lower)
            .set("tick_upper", call.tick_upper);
    });
    calls.pool_call_set_fee_protocols.iter().for_each(|call| {
        tables
            .create_row("pool_call_set_fee_protocol", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("fee_protocol0", call.fee_protocol0)
            .set("fee_protocol1", call.fee_protocol1);
    });
    calls.pool_call_swaps.iter().for_each(|call| {
        tables
            .create_row("pool_call_swap", [("call_tx_hash", call.call_tx_hash.to_string()),("call_ordinal", call.call_ordinal.to_string())])
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount_specified", BigDecimal::from_str(&call.amount_specified).unwrap())
            .set("data", Hex(&call.data).to_string())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string())
            .set("sqrt_price_limit_x96", BigDecimal::from_str(&call.sqrt_price_limit_x96).unwrap())
            .set("zero_for_one", call.zero_for_one);
    });
}


fn graph_factory_out(events: &contract::Events, tables: &mut EntityChangesTables) {
    // Loop over all the abis events to create table changes
    events.factory_fee_amount_enableds.iter().for_each(|evt| {
        tables
            .create_row("factory_fee_amount_enabled", format!("{}-{}", evt.evt_tx_hash, evt.evt_index))
            .set("evt_tx_hash", &evt.evt_tx_hash)
            .set("evt_index", evt.evt_index)
            .set("evt_block_time", evt.evt_block_time.as_ref().unwrap())
            .set("evt_block_number", evt.evt_block_number)
            .set("fee", evt.fee)
            .set("tick_spacing", evt.tick_spacing);
    });
    events.factory_owner_changeds.iter().for_each(|evt| {
        tables
            .create_row("factory_owner_changed", format!("{}-{}", evt.evt_tx_hash, evt.evt_index))
            .set("evt_tx_hash", &evt.evt_tx_hash)
            .set("evt_index", evt.evt_index)
            .set("evt_block_time", evt.evt_block_time.as_ref().unwrap())
            .set("evt_block_number", evt.evt_block_number)
            .set("new_owner", Hex(&evt.new_owner).to_string())
            .set("old_owner", Hex(&evt.old_owner).to_string());
    });
    events.factory_pool_createds.iter().for_each(|evt| {
        tables
            .create_row("factory_pool_created", format!("{}-{}", evt.evt_tx_hash, evt.evt_index))
            .set("evt_tx_hash", &evt.evt_tx_hash)
            .set("evt_index", evt.evt_index)
            .set("evt_block_time", evt.evt_block_time.as_ref().unwrap())
            .set("evt_block_number", evt.evt_block_number)
            .set("fee", evt.fee)
            .set("pool", Hex(&evt.pool).to_string())
            .set("tick_spacing", evt.tick_spacing)
            .set("token0", Hex(&evt.token0).to_string())
            .set("token1", Hex(&evt.token1).to_string());
    });
}
fn graph_pool_calls_out(calls: &contract::Calls, tables: &mut EntityChangesTables) {
    // Loop over all the abis calls to create table changes
    calls.pool_call_burns.iter().for_each(|call| {
        tables
            .create_row("pool_call_burn", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount", BigDecimal::from_str(&call.amount).unwrap())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("tick_lower", call.tick_lower)
            .set("tick_upper", call.tick_upper);
    });
    calls.pool_call_collects.iter().for_each(|call| {
        tables
            .create_row("pool_call_collect", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount0_requested", BigDecimal::from_str(&call.amount0_requested).unwrap())
            .set("amount1_requested", BigDecimal::from_str(&call.amount1_requested).unwrap())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string())
            .set("tick_lower", call.tick_lower)
            .set("tick_upper", call.tick_upper);
    });
    calls.pool_call_collect_protocols.iter().for_each(|call| {
        tables
            .create_row("pool_call_collect_protocol", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount0_requested", BigDecimal::from_str(&call.amount0_requested).unwrap())
            .set("amount1_requested", BigDecimal::from_str(&call.amount1_requested).unwrap())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string());
    });
    calls.pool_call_flashes.iter().for_each(|call| {
        tables
            .create_row("pool_call_flash", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount0", BigDecimal::from_str(&call.amount0).unwrap())
            .set("amount1", BigDecimal::from_str(&call.amount1).unwrap())
            .set("data", Hex(&call.data).to_string())
            .set("recipient", Hex(&call.recipient).to_string());
    });
    calls.pool_call_increase_observation_cardinality_nexts.iter().for_each(|call| {
        tables
            .create_row("pool_call_increase_observation_cardinality_next", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("observation_cardinality_next", call.observation_cardinality_next);
    });
    calls.pool_call_initializes.iter().for_each(|call| {
        tables
            .create_row("pool_call_initialize", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("sqrt_price_x96", BigDecimal::from_str(&call.sqrt_price_x96).unwrap());
    });
    calls.pool_call_mints.iter().for_each(|call| {
        tables
            .create_row("pool_call_mint", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount", BigDecimal::from_str(&call.amount).unwrap())
            .set("data", Hex(&call.data).to_string())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string())
            .set("tick_lower", call.tick_lower)
            .set("tick_upper", call.tick_upper);
    });
    calls.pool_call_set_fee_protocols.iter().for_each(|call| {
        tables
            .create_row("pool_call_set_fee_protocol", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("fee_protocol0", call.fee_protocol0)
            .set("fee_protocol1", call.fee_protocol1);
    });
    calls.pool_call_swaps.iter().for_each(|call| {
        tables
            .create_row("pool_call_swap", format!("{}-{}", call.call_tx_hash, call.call_ordinal))
            .set("call_tx_hash", &call.call_tx_hash)
            .set("call_ordinal", call.call_ordinal)
            .set("call_block_time", call.call_block_time.as_ref().unwrap())
            .set("call_block_number", call.call_block_number)
            .set("call_success", call.call_success)
            .set("call_address", &call.call_address)
            .set("amount_specified", BigDecimal::from_str(&call.amount_specified).unwrap())
            .set("data", Hex(&call.data).to_string())
            .set("output_amount0", BigDecimal::from_str(&call.output_amount0).unwrap())
            .set("output_amount1", BigDecimal::from_str(&call.output_amount1).unwrap())
            .set("recipient", Hex(&call.recipient).to_string())
            .set("sqrt_price_limit_x96", BigDecimal::from_str(&call.sqrt_price_limit_x96).unwrap())
            .set("zero_for_one", call.zero_for_one);
    });
  }
#[substreams::handlers::store]
fn store_pool_created(blk: eth::Block, store: StoreSetInt64) {
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
) -> Result<contract::Events, substreams::errors::Error> {
    let mut events = contract::Events::default();
    map_factory_events(&blk, &mut events);
    Ok(events)
}
#[substreams::handlers::map]
fn map_calls(
    blk: eth::Block,
    store_pool: StoreGetInt64,
    
) -> Result<contract::Calls, substreams::errors::Error> {
let mut calls = contract::Calls::default();
    map_pool_calls(&blk, &store_pool, &mut calls);
    Ok(calls)
}

#[substreams::handlers::map]
fn db_out(events: contract::Events, calls: contract::Calls) -> Result<DatabaseChanges, substreams::errors::Error> {
    // Initialize Database Changes container
    let mut tables = DatabaseChangeTables::new();
    db_factory_out(&events, &mut tables);
    db_pool_calls_out(&calls, &mut tables);
    Ok(tables.to_database_changes())
}

#[substreams::handlers::map]
fn graph_out(events: contract::Events, calls: contract::Calls) -> Result<EntityChanges, substreams::errors::Error> {
    // Initialize Database Changes container
    let mut tables = EntityChangesTables::new();
    graph_factory_out(&events, &mut tables);
    graph_pool_calls_out(&calls, &mut tables);
    Ok(tables.to_entity_changes())
}
