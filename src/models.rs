use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Serialize, Deserialize, Debug, Default, Eq, PartialEq, Clone)]
pub struct Item {
    pub id: String,
    pub parent: Option<String>,
    pub children: Option<Vec<String>>,
    pub title: String,
    pub description: Option<String>,
}

impl Item {
    pub fn new(id_length: usize, title: &str) -> Item {
        let mut new_item: Item = Default::default();
        new_item.id = generate_id(id_length);
        new_item.title = String::from(title);

        new_item
    }
}

fn generate_id(length: usize) -> String {
    let id: String = thread_rng()
        .sample_iter(&Alphanumeric)
        .take(length)
        .collect();

    id
}

#[derive(Debug, Serialize, Deserialize, Default, Eq, PartialEq)]
pub struct Items {
    pub items: HashMap<String, Item>,
}
