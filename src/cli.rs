use super::config;
use super::models::Items;
use anyhow::Result;
use ptree;
use reqwest;

pub struct CLI {
    host: String,
}

pub fn new() -> CLI {
    let config = config::get_cli_config();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };
    CLI { host: config.host }
}

impl CLI {
    pub fn list_todos(self) -> Result<()> {
        // [abc123] My precious title here :: some extended definition here

        let list_endpoint = self.host;
        let req = reqwest::blocking::get(&list_endpoint)?
            .json::<Items>()
            .unwrap();
        let single_item = &req.items["TdshZ"];

        let tree = ptree::TreeBuilder::new(single_item.pretty_print())
            .begin_child("l_".to_string())
            .build();

        ptree::print_tree(&tree)?;
        Ok(())
        // use an http client to query for json
        // turn json into items hashmap
        // for each item we check if it has any children and then
        // attach those children to it recusively
        // we keep a list of todos which have been processed so we don't
        // reprocess them
    }
}
