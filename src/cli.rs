use super::config;
use super::models::{AddItemRequest, Item, Items, UpdateItemRequest};
use anyhow::{anyhow, Result};
use ptree;
use reqwest;
use std::collections::{HashMap, HashSet};
use std::env;
use std::fs;
use std::io::Write;
use std::process::Command;
use tempfile::Builder;
use which;

const VISUAL_VAR: &'static str = "VISUAL";
const EDITOR_VAR: &'static str = "EDITOR";
const DEFAULT_EDITOR: &'static str = "vi";

pub struct CLI {
    host: String,
}

//TODO(clintjedwards): Handle errors for all of this
//TODO(clintjedwards): prevent title from being an empty string

pub fn new() -> CLI {
    let config = config::get_cli();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };
    CLI { host: config.host }
}

impl<'a> CLI {
    pub fn list_todos(&self, include_completed: bool) -> Result<()> {
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

        builder.build(include_completed);

        Ok(())
    }

    pub fn add_todo(&self, item: AddItemRequest, interactive: bool) -> Result<()> {
        // We populate the item with default values so it looks good when presented in
        // TOML form
        let mut item = item.clone();
        item.description = Some(item.description.unwrap_or_default());
        item.parent = Some(item.parent.unwrap_or_default());
        item.link = Some(item.link.unwrap_or_default());

        if interactive {
            let mut file = Builder::new().suffix(".toml").rand_bytes(5).tempfile()?;

            let item_toml = toml::to_string_pretty(&item)?;
            write!(file, "{}", item_toml)?;

            open_editor(file.path().to_str().unwrap());

            let item_toml = fs::read_to_string(file.path())?;
            item = toml::from_str(&item_toml)?;
        }

        // We need to set parent back to none if it is empty
        if item.parent.is_some() && item.parent.clone().unwrap().is_empty() {
            item.parent = None;
        }

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
        println!("{}", item.short_format());

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

    // clean_up_todos removes all todo items which don't belong to an unresolved todo.
    pub fn clean_up_todos(&self) -> Result<()> {
        let client = reqwest::blocking::Client::new();

        let list_endpoint = self.host.clone();
        let response = reqwest::blocking::get(&list_endpoint)?
            .json::<Items>()
            .unwrap();

        for (id, item) in response.items.clone() {
            let remove_endpoint = format!("{}/{}", self.host.clone(), id);

            if !item.completed {
                continue;
            }

            let node = get_root_node(&item, &response.items);
            if node.completed {
                client.delete(&remove_endpoint).send()?;
            }

            continue;
        }

        Ok(())
    }

    pub fn update_todo(&self, id: &str, item: UpdateItemRequest, interactive: bool) -> Result<()> {
        // Clone the item so we can manipulate it
        let mut item = item.clone();

        if interactive {
            let get_endpoint = format!("{}/{}", self.host.clone(), id);
            let response = reqwest::blocking::get(&get_endpoint)?;
            if response.status().is_client_error() {
                return Err(anyhow!("could not find item {}", id));
            }

            let current_item = response.json::<UpdateItemRequest>().unwrap();

            item.description = Some(current_item.description.unwrap_or_default());
            item.children = Some(current_item.children.unwrap_or_default());
            item.completed = Some(current_item.completed.unwrap_or_default());
            item.link = Some(current_item.link.unwrap_or_default());
            item.parent = Some(current_item.parent.unwrap_or_default());
            item.title = Some(current_item.title.unwrap_or_default());

            let mut file = Builder::new().suffix(".toml").rand_bytes(5).tempfile()?;

            let item_toml = toml::to_string_pretty(&item)?;

            // The toml converter does not print fields which are None. Converting those fields
            // to some for the purposes of looking pretty was not trivial and probably requires
            // some conversion methods.
            writeln!(
                file,
                "{}",
                "# Fields: title, description, children, completed, parent, link\n"
            )?;
            write!(file, "{}", item_toml)?;

            open_editor(file.path().to_str().unwrap());

            let item_toml = fs::read_to_string(file.path())?;
            item = toml::from_str(&item_toml)?;
        }

        // We need to set parent back to none if it is empty
        if item.parent.is_some() && item.parent.clone().unwrap().is_empty() {
            item.parent = None;
        }

        let update_endpoint = format!("{}/{}", self.host.clone(), id);
        let client = reqwest::blocking::Client::new();
        client.put(&update_endpoint).json(&item).send()?;

        Ok(())
    }

    pub fn remove_todo(&self, id: &str) -> Result<()> {
        let remove_endpoint = format!("{}/{}", self.host.clone(), id);
        let client = reqwest::blocking::Client::new();
        client.delete(&remove_endpoint).send()?;

        Ok(())
    }
}

// Given any node in our tree, return the root node
fn get_root_node(item: &Item, item_map: &HashMap<String, Item>) -> Item {
    match get_parent_node(&item, &item_map) {
        Some(next_item) => get_root_node(&next_item, item_map),
        None => item.clone(),
    }
}

fn get_parent_node(item: &Item, item_map: &HashMap<String, Item>) -> Option<Item> {
    if item.parent.is_none() {
        return None;
    }

    match item_map.get(&item.parent.clone().unwrap()) {
        Some(parent) => Some(parent.clone()),
        None => None,
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
                None => self.add_to_tree(&item, include_completed),
                Some(_) => continue,
            }
        }

        ptree::print_tree(&self.tree.build()).unwrap();
    }

    // takes a given node and adds it to the given TreeBuilder object
    // it then looks at all child nodes and does the same
    // Once it is out of child nodes it returns
    fn add_to_tree(&mut self, item: &Item, include_completed: bool) {
        if self.visited.contains(&item.id.clone()) {
            return;
        }
        self.tree.begin_child(item.short_format());

        self.visited.insert(item.id.clone());

        if let Some(children) = &item.children {
            for child_id in children {
                if let Some(child) = &self.items_map.get(child_id) {
                    if !include_completed {
                        if child.completed {
                            continue;
                        }
                    }
                    self.add_to_tree(&child, include_completed)
                }
            }
        }
        self.tree.end_child();
    }
}

fn get_editor_path() -> String {
    if let Ok(path) = env::var(VISUAL_VAR) {
        return path;
    }

    if let Ok(path) = env::var(EDITOR_VAR) {
        return path;
    }

    let path =
        which::which(DEFAULT_EDITOR).expect("vi not installed; could not open a default editor");

    path.as_path().display().to_string()
}

fn open_editor(filename: &str) {
    let path = get_editor_path();

    // break the path apart so we can append the filename
    // appending the filename to most editors will open that file
    let mut path: Vec<&str> = path.split(char::is_whitespace).collect();
    path.push(filename);
    let path = path.as_slice();

    // reconstruct the path array into the command
    let mut command = Command::new(path[0]);
    command.args(&path[1..]);

    command.output().expect("could not open editor");
}
