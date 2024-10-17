mod abi;
mod pb;
use hex_literal::hex;
use pb::contract::v1 as contract;
use substreams::Hex;
use substreams_ethereum::pb::eth::v2 as eth;
use substreams_ethereum::Event;

#[allow(unused_imports)]
use num_traits::cast::ToPrimitive;
use std::str::FromStr;
use substreams::scalar::BigDecimal;

substreams_ethereum::init!();

const BAYC_TRACKED_CONTRACT: [u8; 20] = hex!("bc4ca0eda7647a8ab7c2061c2e118a18a936f13d");

fn map_bayc_events(blk: &eth::Block, events: &mut contract::Events) {
    events.bayc_approvals.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| log.address == BAYC_TRACKED_CONTRACT)
                .filter_map(|log| {
                    if let Some(event) = abi::bayc_contract::events::Approval::match_and_decode(log) {
                        return Some(contract::BaycApproval {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            approved: event.approved,
                            owner: event.owner,
                            token_id: event.token_id.to_string(),
                        });
                    }

                    None
                })
        })
        .collect());
    events.bayc_approval_for_alls.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| log.address == BAYC_TRACKED_CONTRACT)
                .filter_map(|log| {
                    if let Some(event) = abi::bayc_contract::events::ApprovalForAll::match_and_decode(log) {
                        return Some(contract::BaycApprovalForAll {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            approved: event.approved,
                            operator: event.operator,
                            owner: event.owner,
                        });
                    }

                    None
                })
        })
        .collect());
    events.bayc_ownership_transferreds.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| log.address == BAYC_TRACKED_CONTRACT)
                .filter_map(|log| {
                    if let Some(event) = abi::bayc_contract::events::OwnershipTransferred::match_and_decode(log) {
                        return Some(contract::BaycOwnershipTransferred {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            new_owner: event.new_owner,
                            previous_owner: event.previous_owner,
                        });
                    }

                    None
                })
        })
        .collect());
    events.bayc_transfers.append(&mut blk
        .receipts()
        .flat_map(|view| {
            view.receipt.logs.iter()
                .filter(|log| log.address == BAYC_TRACKED_CONTRACT)
                .filter_map(|log| {
                    if let Some(event) = abi::bayc_contract::events::Transfer::match_and_decode(log) {
                        return Some(contract::BaycTransfer {
                            evt_tx_hash: Hex(&view.transaction.hash).to_string(),
                            evt_index: log.block_index,
                            evt_block_time: Some(blk.timestamp().to_owned()),
                            evt_block_number: blk.number,
                            from: event.from,
                            to: event.to,
                            token_id: event.token_id.to_string(),
                        });
                    }

                    None
                })
        })
        .collect());
}
fn map_bayc_calls(blk: &eth::Block, calls: &mut contract::Calls) {
    calls.bayc_call_approves.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::Approve::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::Approve::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycApproveCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                to: decoded_call.to,
                                token_id: decoded_call.token_id.to_string(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_emergency_set_starting_index_blocks.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::EmergencySetStartingIndexBlock::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::EmergencySetStartingIndexBlock::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycEmergencySetStartingIndexBlockCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_flip_sale_states.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::FlipSaleState::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::FlipSaleState::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycFlipSaleStateCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_mint_apes.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::MintApe::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::MintApe::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycMintApeCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                number_of_tokens: decoded_call.number_of_tokens.to_string(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_renounce_ownerships.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::RenounceOwnership::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::RenounceOwnership::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycRenounceOwnershipCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_reserve_apes.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::ReserveApes::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::ReserveApes::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycReserveApesCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_safe_transfer_from_1s.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::SafeTransferFrom1::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::SafeTransferFrom1::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycSafeTransferFrom1call {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                from: decoded_call.from,
                                to: decoded_call.to,
                                token_id: decoded_call.token_id.to_string(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_safe_transfer_from_2s.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::SafeTransferFrom2::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::SafeTransferFrom2::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycSafeTransferFrom2call {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                from: decoded_call.from,
                                to: decoded_call.to,
                                token_id: decoded_call.token_id.to_string(),
                                u_data: decoded_call.u_data,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_set_approval_for_alls.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::SetApprovalForAll::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::SetApprovalForAll::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycSetApprovalForAllCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                approved: decoded_call.approved,
                                operator: decoded_call.operator,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_set_base_uris.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::SetBaseUri::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::SetBaseUri::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycSetBaseUriCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                base_uri: decoded_call.base_uri,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_set_provenance_hashes.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::SetProvenanceHash::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::SetProvenanceHash::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycSetProvenanceHashCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                provenance_hash: decoded_call.provenance_hash,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_set_reveal_timestamps.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::SetRevealTimestamp::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::SetRevealTimestamp::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycSetRevealTimestampCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                reveal_time_stamp: decoded_call.reveal_time_stamp.to_string(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_set_starting_indices.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::SetStartingIndex::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::SetStartingIndex::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycSetStartingIndexCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_transfer_froms.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::TransferFrom::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::TransferFrom::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycTransferFromCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                from: decoded_call.from,
                                to: decoded_call.to,
                                token_id: decoded_call.token_id.to_string(),
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_transfer_ownerships.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::TransferOwnership::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::TransferOwnership::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycTransferOwnershipCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
                                new_owner: decoded_call.new_owner,
                            })
                        },
                        Err(_) => None,
                    }
                })
        })
        .collect());
    calls.bayc_call_withdraws.append(&mut blk
        .transactions()
        .flat_map(|tx| {
            tx.calls.iter()
                .filter(|call| call.address == BAYC_TRACKED_CONTRACT && abi::bayc_contract::functions::Withdraw::match_call(call))
                .filter_map(|call| {
                    match abi::bayc_contract::functions::Withdraw::decode(call) {
                        Ok(decoded_call) => {
                            Some(contract::BaycWithdrawCall {
                                call_tx_hash: Hex(&tx.hash).to_string(),
                                call_block_time: Some(blk.timestamp().to_owned()),
                                call_block_number: blk.number,
                                call_ordinal: call.begin_ordinal,
                                call_success: !call.state_reverted,
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
#[substreams::handlers::map]
fn map_events(blk: eth::Block) -> Result<contract::Events, substreams::errors::Error> {
    let mut events = contract::Events::default();
    map_bayc_events(&blk, &mut events);
    Ok(events)
}
#[substreams::handlers::map]
fn map_calls(blk: eth::Block) -> Result<contract::Calls, substreams::errors::Error> {
let mut calls = contract::Calls::default();
    map_bayc_calls(&blk, &mut calls);
    Ok(calls)
}

