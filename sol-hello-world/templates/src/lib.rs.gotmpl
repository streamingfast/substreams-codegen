#[allow(unused)]
mod pb;
use pb::mydata::v1 as mydata;
use substreams_solana::b58;
use substreams_solana::pb::sf::solana::r#type::v1::Block;

// Get the program id for the pump fun program as byte array for fast comparison
// If you change this for another program, be sure to check substreams.yaml and:
// * If you are using block filters update the filter accordingly
// * Align the module's initialBlock with the creation of your program

const PUMP_FUN_PROGRAM_ID: [u8; 32] = b58!("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P");

#[substreams::handlers::map]
fn map_my_data(blk: Block) -> mydata::MyData {
    let mut account_vec: Vec<Vec<u8>> = Vec::new();

    blk.walk_instructions()
        .filter(|instruction| instruction.program_id() == PUMP_FUN_PROGRAM_ID)
        .for_each(|instruction| {
            instruction.accounts().into_iter().for_each(|account| {
                account_vec.push(account.0.clone());
            });
        });

    account_vec.sort();
    account_vec.dedup();

    let mut my_data = mydata::MyData::default();
    my_data.accounts = account_vec;
    my_data
}
