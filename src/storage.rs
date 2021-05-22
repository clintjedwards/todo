use super::models::{Item, Items};
use anyhow::{anyhow, Context, Result};
use std::collections::HashSet;
use std::time::SystemTime;

// TODO(clintjedwards): Transactions seem difficult to structure such that we can reuse
// functions here. This makes it so a lot of this code is duplicated.
//  Revisit this when sled is older and transactions are more figured out.

#[derive(Debug, Clone)]
pub struct Storage {
    db: sled::Db,
}

// new returns a new Storage struct with sled db backing or fails
pub fn new(path: &str) -> Storage {
    Storage {
        db: sled::open(path).expect("could not open sled database"),
    }
}

impl Storage {
    // get_all_items returns a unpaginated btreemap of all todo items
    pub fn get_all_items(&self) -> Result<Items> {
        let items_iter = self.db.iter();
        let mut items: Items = Default::default();

        for item in items_iter {
            match item {
                Ok(item_pair) => {
                    let (key, value) = item_pair;
                    let key = String::from_utf8(key.to_vec())?;
                    let value: Item = bincode::deserialize(&value.to_vec())?;
                    items.items.insert(key, value);
                }
                Err(e) => return Err(anyhow!(e)),
            }
        }

        Ok(items)
    }

    // get_item returns a single item by id
    pub fn get_item(&self, id: &str) -> Result<Option<Item>> {
        match self
            .db
            .get(id)
            .with_context(|| format!("Failed to get item {}", id))?
        {
            None => Ok(None),
            Some(value) => {
                let value: Item = bincode::deserialize(&value.to_vec())?;
                Ok(Some(value))
            }
        }
    }

    // add_item adds a single item
    pub fn add_item(&self, item: Item) -> Result<()> {
        if item.id.is_empty() {
            return Err(anyhow!("item must include an id"));
        }

        let transaction_result = self.db.transaction(|tx_db| {
            let result = link_child_to_parent(tx_db, &item.id.clone(), item.parent.clone());
            match result {
                Ok(_) => {}
                Err(e) => {
                    return sled::transaction::abort(e);
                }
            };

            let raw_item: Vec<u8> = match bincode::serialize(&item) {
                Ok(raw_item) => raw_item,
                Err(e) => return sled::transaction::abort(anyhow!(e)),
            };

            let item_id: &str = &item.id.clone();
            tx_db.insert(item_id, raw_item)?;

            Ok(())
        });

        match transaction_result {
            Ok(()) => Ok(()),
            Err(e) => Err(anyhow!(
                "could not complete add item {}; failed transaction: {}",
                item.id,
                e.to_string()
            )),
        }
    }

    // update_item takes the newer version of an item and replaces the old version with it
    // update_item also does intelligent things like detect changes in parent or children
    // properties and updates the correct dependencies to reflect the correct state.
    //
    // For example: updating an item with a new set of children will diff the previous version
    // and visit all child node and update their parent to the correct thing.
    pub fn update_item(&self, item: Item) -> Result<()> {
        let transaction_result = self.db.transaction(|tx_db| {
            let raw_old_item = tx_db.get(&item.id.clone())?.unwrap();
            let old_item: Item = match bincode::deserialize(&raw_old_item.to_vec()) {
                Ok(old_item) => old_item,
                Err(e) => return sled::transaction::abort(anyhow!(e)),
            };

            // If we have removed a parent unlink the child from the parent
            if item.parent.is_none() && old_item.parent.is_some() {
                let result = unlink_child_from_parent(
                    tx_db,
                    &item.id.clone(),
                    &old_item.parent.clone().unwrap(),
                );
                match result {
                    Ok(_) => {}
                    Err(e) => {
                        return sled::transaction::abort(e);
                    }
                };
            }

            // If we add or change the parent we should update the parent to contain the child id
            if item.parent.is_some() && (item.parent != old_item.parent) {
                let result = link_child_to_parent(tx_db, &item.id.clone(), item.parent.clone());
                match result {
                    Ok(_) => {}
                    Err(e) => {
                        return sled::transaction::abort(e);
                    }
                };
            }

            // If the user updates the children in any way we need to figure out what changed
            // so we can update the appropriate children that they are orphaned or adopted.
            let new_list_children = match item.children.clone() {
                Some(children) => children,
                None => vec![],
            };
            let old_list_children = match old_item.children.clone() {
                Some(children) => children,
                None => vec![],
            };

            let (removed_children, added_children) =
                find_list_updates(old_list_children, new_list_children);

            for child_id in added_children {
                let result = link_parent_to_child(tx_db, &child_id, &item.id.clone());
                match result {
                    Ok(_) => {}
                    Err(e) => {
                        return sled::transaction::abort(e);
                    }
                };
            }

            for child_id in removed_children {
                let result = unlink_parent_from_child(tx_db, &child_id);
                match result {
                    Ok(_) => {}
                    Err(e) => {
                        return sled::transaction::abort(e);
                    }
                };
            }

            let mut item = item.clone();
            item.id = old_item.id.clone();
            item.added = old_item.added;
            item.modified = get_current_epoch_time();

            let raw_item: Vec<u8> = match bincode::serialize(&item) {
                Ok(raw_item) => raw_item,
                Err(e) => return sled::transaction::abort(anyhow!(e)),
            };

            let item_id: &str = &item.id;
            tx_db.insert(item_id, raw_item)?;

            Ok(())
        });

        match transaction_result {
            Ok(()) => Ok(()),
            Err(e) => Err(anyhow!(
                "could not complete add item {}; failed transaction: {}",
                item.id,
                e.to_string()
            )),
        }
    }

    // delete_item deletes a single item
    pub fn delete_item(&self, id: &str) -> Result<()> {
        self.db
            .remove(id)
            .with_context(|| format!("Failed to remove item {}", id))?;
        Ok(())
    }

    // this function is only used in testing
    #[allow(dead_code)]
    fn clear(&self) {
        let _ = self.db.clear();
    }
}

// Adds a parent to an already existing child
fn link_parent_to_child(
    db: &sled::transaction::TransactionalTree,
    child_id: &str,
    parent_id: &str,
) -> Result<()> {
    let raw_child = db.get(child_id)?.unwrap();
    let mut child: Item = bincode::deserialize(&raw_child.to_vec())?;

    child.parent = Some(parent_id.to_string());
    child.modified = get_current_epoch_time();

    let raw_child: Vec<u8> = bincode::serialize(&child)
        .with_context(|| format!("Failed to encode item {}", child_id))?;

    db.insert(child_id, raw_child)
        .with_context(|| format!("Failed to update item {}", child_id))?;

    Ok(())
}

// Adds a new child to the parent item, if that child does not already exist.
fn link_child_to_parent(
    db: &sled::transaction::TransactionalTree,
    child_id: &str,
    parent_id: Option<String>,
) -> Result<()> {
    if parent_id.is_none() {
        return Ok(());
    }

    let parent_id = parent_id.unwrap();
    let raw_parent_option = db.get(&parent_id)?;
    let raw_parent = match raw_parent_option {
        Some(parent) => parent,
        None => {
            return Err(anyhow!("could not find parent id {}", &parent_id));
        }
    };
    let mut parent: Item = bincode::deserialize(&raw_parent.to_vec())?;

    match parent.children.clone() {
        Some(mut children) => {
            if children.contains(&child_id.to_string()) {
                return Ok(());
            }
            children.push(child_id.to_string());
            parent.children = Some(children);
        }
        None => {
            parent.children = Some(vec![child_id.to_string()]);
        }
    }

    parent.modified = get_current_epoch_time();

    let raw_parent: Vec<u8> = bincode::serialize(&parent)
        .with_context(|| format!("Failed to encode item {}", parent_id))?;

    let parent_id: &str = &parent_id;
    db.insert(parent_id, raw_parent)
        .with_context(|| format!("Failed to update item {}", parent_id))?;

    Ok(())
}

// Removes child ID from parent
fn unlink_child_from_parent(
    db: &sled::transaction::TransactionalTree,
    child_id: &str,
    parent_id: &str,
) -> Result<()> {
    let raw_parent = db.get(&parent_id)?.unwrap();
    let mut parent: Item = bincode::deserialize(&raw_parent.to_vec())?;

    match parent.children.clone() {
        Some(mut children) => {
            children.retain(|id| id != child_id);
            parent.children = Some(children);
        }
        None => return Ok(()),
    }

    parent.modified = get_current_epoch_time();

    let raw_parent: Vec<u8> = bincode::serialize(&parent)
        .with_context(|| format!("Failed to encode item {}", parent_id))?;

    let parent_id: &str = &parent_id;
    db.insert(parent_id, raw_parent)
        .with_context(|| format!("Failed to update item {}", parent_id))?;

    Ok(())
}

// remove a parent id from a child
fn unlink_parent_from_child(
    db: &sled::transaction::TransactionalTree,
    child_id: &str,
) -> Result<()> {
    let raw_child = db.get(&child_id)?.unwrap();
    let mut child: Item = bincode::deserialize(&raw_child.to_vec())?;

    if child.parent.is_some() {
        unlink_child_from_parent(db, &child.id, &child.parent.unwrap())?;
    }
    child.parent = None;
    child.modified = get_current_epoch_time();

    let raw_child: Vec<u8> = bincode::serialize(&child)
        .with_context(|| format!("Failed to encode item {}", child_id))?;

    let child_id: &str = &child_id;
    db.insert(child_id, raw_child)
        .with_context(|| format!("Failed to update item {}", child_id))?;

    Ok(())
}

fn find_list_difference(list1: Vec<String>, list2: Vec<String>) -> Vec<String> {
    let list1: HashSet<_> = list1.iter().cloned().collect();
    let list2: HashSet<_> = list2.iter().cloned().collect();

    let diff = list1.difference(&list2).cloned().collect();
    diff
}

fn find_list_updates(old_list: Vec<String>, new_list: Vec<String>) -> (Vec<String>, Vec<String>) {
    let removals = find_list_difference(old_list.clone(), new_list.clone());
    let additions = find_list_difference(new_list, old_list);
    (removals, additions)
}

fn get_current_epoch_time() -> u64 {
    SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs()
}

#[cfg(test)]
#[path = "./storage_test.rs"]
mod storage_test;
