mod pb;
use pb::mydata::v1 as mydata;

use substreams_solana::pb::sf::solana::r#type::v1::Block;

#[substreams::handlers::map]
fn map_my_data(blk: Block) -> mydata::MyData {
    let mut my_data = mydata::MyData::default();
    my_data.block_hash = blk.blockhash.to_string();
    my_data.block_slot = blk.slot;
    my_data.block_timestamp = blk.block_time.clone().unwrap_or_default().timestamp as u64;
    my_data.transactions_len = blk.transactions.len() as u64;
    my_data.instructions_len = blk.walk_instructions().count() as u64;
    my_data
}
