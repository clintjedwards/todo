use sled;

pub fn init_database(path: &str) -> sled::Db {
    sled::open(path).expect("open")
}
