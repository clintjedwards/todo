use super::config;
use super::models::{Item, Items, UpdateItemRequest};
use anyhow::{anyhow, Result};
use ptree;
use reqwest;
use std::collections::{HashMap, HashSet};

pub struct CLI {
    host: String,
}

//TODO(clintjedwards): Handle errors for all of this
//TODO(clintjedwards): prevent title from being an empty string

pub fn new() -> CLI {
    let config = config::get_cli_config();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };
    CLI { host: config.host }
}

impl CLI {
    pub fn list_todos(&self, include_completed: bool) -> Result<()> {
        let list_endpoint = self.host.clone();
        let response = reqwest::blocking::get(&list_endpoint)?
            .json::<Items>()
            .unwrap();

        //TODO(clintjedwards): update this to error handle

        let todo_tree = ptree::TreeBuilder::new("Todo List".to_string());
        let mut builder = TreeBuilder {
            items_map: &response.items,
            visited: HashSet::new(),
            tree: todo_tree,
        };

        builder.build(include_completed);

        Ok(())
    }

    pub fn add_todo(&self, item: Item) -> Result<()> {
        let add_endpoint = self.host.clone();
        let client = reqwest::blocking::Client::new();
        client.post(&add_endpoint).json(&item).send()?;

        Ok(())
    }

    pub fn get_todo(&self, id: &str) -> Result<()> {
        let get_endpoint = format!("{}/{}", self.host.clone(), id);
        let response = reqwest::blocking::get(&get_endpoint)?;
        if response.status().is_client_error() {
            return Err(anyhow!("could not find item {}", id));
        }

        let item = response.json::<Item>().unwrap();
        println!("{}", item.format_colorized());

        Ok(())
    }

    // complete_todo toggles the completion parameter
    pub fn complete_todo(&self, id: &str) -> Result<()> {
        let get_endpoint = format!("{}/{}", self.host.clone(), id);
        let response = reqwest::blocking::get(&get_endpoint)?;
        if response.status().is_client_error() {
            return Err(anyhow!("could not find item {}", id));
        }

        let old_item = response.json::<Item>().unwrap();

        let mut updated_item = old_item.clone();
        if updated_item.completed {
            updated_item.completed = false;
        } else {
            updated_item.completed = true;
        }

        let update_endpoint = format!("{}/{}", self.host.clone(), id);
        let client = reqwest::blocking::Client::new();
        client.put(&update_endpoint).json(&updated_item).send()?;

        Ok(())
    }

    pub fn update_todo(&self, item: UpdateItemRequest) -> Result<()> {
        let get_endpoint = format!("{}/{}", self.host.clone(), item.id.clone());
        let response = reqwest::blocking::get(&get_endpoint)?;
        if response.status().is_client_error() {
            return Err(anyhow!("could not find item {}", item.id.clone()));
        }

        let old_item = response.json::<Item>().unwrap();

        // Only replace the things that have changed
        let mut updated_item = old_item.clone();
        if let Some(title) = item.title {
            updated_item.title = title;
        }

        if let Some(parent) = item.parent {
            updated_item.parent = Some(parent);
        }

        if let Some(children) = item.children {
            updated_item.children = Some(children);
        }

        if let Some(description) = item.description {
            updated_item.description = Some(description);
        }

        if let Some(link) = item.link {
            updated_item.link = Some(link);
        }

        let update_endpoint = format!("{}/{}", self.host.clone(), item.id.clone());
        let client = reqwest::blocking::Client::new();
        client.put(&update_endpoint).json(&updated_item).send()?;

        Ok(())
    }

    pub fn remove_todo(&self, id: &str) -> Result<()> {
        let remove_endpoint = format!("{}/{}", self.host.clone(), id);
        let client = reqwest::blocking::Client::new();
        client.delete(&remove_endpoint).send()?;

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
    fn build(&mut self, include_completed: bool) {
        //TODO(clintjedwards): make this list sortable by different filters
        //TODO(clintjedwards): auto-sort completed at the end of the list
        for item in self.items_map.values() {
            if !include_completed {
                if item.completed {
                    continue;
                }
            }
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
        self.tree.begin_child(item.format_colorized());

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
