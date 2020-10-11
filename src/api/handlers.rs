use super::super::models::Item;
use super::config;
use super::storage;
use tide::prelude::*;
use tide::{Request, Result};

struct API {
    db: storage::Storage,
    config: config::Config,
}

impl API {
    // pub async fn get_all_items_handler(&self, req: Request<()>) -> Result<()> {
    //     let items = self.db.get_all_items()?;
    // }
    // pub async fn add_item_handler(req: Request<()>) -> Result<()> {
    //     Item::new(id_length, title)
    //     Ok(())
    // }

    // pub async fn greet(req: Request<()>) -> Result<String> {
    //     let name = req.param("id").unwrap_or("world".to_owned());
    //     Ok(format!("Hello, {}!", name))
    // }
}
