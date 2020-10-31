use colored::*;
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
    pub link: Option<String>,
    pub completed: bool,
    pub added: u64,    // Epoch date when item was created.
    pub modified: u64, // Epoch date when item was last edited.
}

pub fn new_item(id_length: usize, title: &str) -> Item {
    let mut new_item: Item = Default::default();
    new_item.id = generate_id(id_length);
    new_item.title = String::from(title);
    new_item.completed = false;
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
    // prints an item in the following format:
    // [someid] My precious title here :: some extended definition here
    // TODO(clintjedwards): this should probably be renamed to "format"
    pub fn format_colorized(&self) -> String {
        let mut string_builder = vec![];

        string_builder.push(format!("[{}]", self.id.blue()));
        string_builder.push(format!(" {}", self.title));

        match &self.link {
            Some(link) => string_builder.push(format!(" {} {}", "::".green(), link.yellow())),
            None => {}
        }

        match &self.description {
            Some(desc) => string_builder.push(format!(" \n\t      {}", desc)),
            None => {}
        }

        if self.completed {
            return format!("{}", string_builder.concat().dimmed());
        }

        string_builder.concat()
    }
}

#[derive(Serialize, Deserialize, Clone, Default)]
pub struct UpdateItemRequest {
    pub id: String,
    pub parent: Option<String>,
    pub children: Option<Vec<String>>,
    pub title: Option<String>,
    pub description: Option<String>,
    pub link: Option<String>,
    pub completed: bool,
}

#[derive(Debug, Serialize, Deserialize, Default, Eq, PartialEq)]
pub struct Items {
    pub items: HashMap<String, Item>,
}

fn generate_id(length: usize) -> String {
    let id: String = thread_rng()
        .sample_iter(&Alphanumeric)
        .take(length)
        .collect();

    id
}
