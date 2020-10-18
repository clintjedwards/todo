use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::SystemTime;

#[derive(Serialize, Deserialize, Debug, Default, Eq, PartialEq, Clone)]
pub struct Item {
    pub id: String,
    pub parent: Option<String>,
    pub children: Option<Vec<String>>,
    pub title: String,
    pub description: Option<String>,
    pub added: u64,    // Epoch date when item was created.
    pub modified: u64, // Epoch date when item was last edited.
}

pub fn new_item(id_length: usize, title: &str) -> Item {
    let mut new_item: Item = Default::default();
    new_item.id = generate_id(id_length);
    new_item.title = String::from(title);
    new_item.added = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs();
    new_item.modified = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs();

    new_item
}

impl Item {
    // [abc123] My precious title here :: some extended definition here
    pub fn pretty_print(&self) -> String {
        match &self.description {
            None => format!("[{}] {}", self.id, self.title),
            Some(description) => format!("[{}] {} :: {}", self.id, self.title, description),
        }
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
