use super::config;
use super::models::{Item, Items};
use anyhow::Result;
use ptree;
use reqwest;
use std::collections::{HashMap, HashSet};

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
    pub fn list_todos(&self) -> Result<()> {
        let list_endpoint = self.host.clone();
        let response = reqwest::blocking::get(&list_endpoint)?
            .json::<Items>()
            .unwrap();

        let todo_tree = ptree::TreeBuilder::new("Todo List".to_string());
        let mut builder = TreeBuilder {
            items_map: &response.items,
            visited: HashSet::new(),
            tree: todo_tree,
        };

        builder.build();

        Ok(())
    }

    pub fn add_todo(&self, item: Item) -> Result<()> {
        let add_endpoint = self.host.clone();
        let client = reqwest::blocking::Client::new();
        client.post(&add_endpoint).json(&item).send()?;

        Ok(())
    }
}

struct TreeBuilder<'a> {
    items_map: &'a HashMap<String, Item>,
    visited: HashSet<String>,
    tree: ptree::TreeBuilder,
}

// TreeBuilder allows us to list out all the todo items in a pretty tree format
// that is easily readable.
//
// Because of how the api to build the tree works we have to process all children
// for a parent before we close the parent and move on to the next one. This
// makes it so that we need to do a depth first search by each orphan node in order
// to get a proper tree.
impl<'a> TreeBuilder<'a> {
    fn build(&mut self) {
        for item in self.items_map.values() {
            match item.parent {
                None => self.add_to_tree(&item),
                Some(_) => continue,
            }
        }

        ptree::print_tree(&self.tree.build()).unwrap();
    }

    // takes a given node and adds it to the given TreeBuilder object
    // it then looks at all child nodes and does the same
    // Once it is out of child nodes it returns
    fn add_to_tree(&mut self, item: &Item) {
        if self.visited.contains(&item.id.clone()) {
            return;
        }
        self.tree.begin_child(item.pretty_print());

        self.visited.insert(item.id.clone());

        if let Some(children) = &item.children {
            for child_id in children {
                let child = &self.items_map[child_id];
                self.add_to_tree(&child)
            }
        }
        self.tree.end_child();
    }
}
