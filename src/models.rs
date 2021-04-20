use colored::*;
use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::SystemTime;

#[derive(Serialize, Deserialize, Debug, Default, Eq, PartialEq, Clone)]
pub struct Item {
    pub added: u64,                    // Epoch date when item was created.
    pub children: Option<Vec<String>>, // List of children by ids.
    pub completed: bool,
    pub description: Option<String>,
    pub id: String,
    pub link: Option<String>,
    pub modified: u64, // Epoch date when item was last edited.
    pub parent: Option<String>,
    pub title: String,
}

pub fn new_item(id_length: usize, title: &str) -> Item {
    let mut new_item: Item = Default::default();
    new_item.added = get_current_epoch_time();
    new_item.completed = false;
    new_item.id = generate_id(id_length);
    new_item.modified = get_current_epoch_time();
    new_item.title = String::from(title);

    new_item
}

impl Item {
    // format prints an item in the following format:
    // [id] <title> :: <link>
    //          <description>
    pub fn short_format(&self) -> String {
        let mut string_builder = vec![];

        string_builder.push(format!("[{}]", self.id.blue()));
        string_builder.push(format!(" {}", self.title));

        match &self.link {
            Some(link) => {
                if !link.is_empty() {
                    string_builder.push(format!(" {} {}", "::".green(), link.yellow()))
                }
            }
            None => {}
        }

        match &self.description {
            Some(desc) => {
                if !desc.is_empty() {
                    string_builder.push(format!(" \n\t      {}", desc))
                }
            }
            None => {}
        }

        if self.completed {
            return format!("{}", string_builder.concat().dimmed());
        }

        string_builder.concat()
    }
}

#[derive(Debug, Serialize, Deserialize, Default, Eq, PartialEq)]
pub struct Items {
    pub items: HashMap<String, Item>,
}

#[derive(Serialize, Deserialize, Clone, Default)]
pub struct AddItemRequest {
    pub description: Option<String>,
    pub link: Option<String>,
    pub parent: Option<String>,
    pub title: String,
}

#[derive(Serialize, Deserialize, Clone, Default)]
pub struct UpdateItemRequest {
    pub description: Option<String>,
    pub children: Option<Vec<String>>,
    pub completed: Option<bool>,
    pub link: Option<String>,
    pub parent: Option<String>,
    pub title: Option<String>,
}

fn generate_id(length: usize) -> String {
    let id: String = thread_rng()
        .sample_iter(&Alphanumeric)
        .take(length)
        .collect();

    id
}

fn get_current_epoch_time() -> u64 {
    SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs()
}
