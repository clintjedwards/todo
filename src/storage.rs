use super::models::{Item, Items};
use anyhow::{anyhow, Context, Result};
use sled;
use std::collections::HashSet;
use std::time::SystemTime;

// TODO(clintjedwards): Transactions seem difficult to structure such that we can reuse
// functions here. This makes it so a lot of this code is duplicated.
//  Revisit this when sled is older and transactions are more figured out.

#[derive(Debug, Clone)]
pub struct Storage {
    db: sled::Db,
}

pub fn new(path: &str) -> Storage {
    Storage {
        db: sled::open(path).expect("could not open sled database"),
    }
}

impl Storage {
    // get_all_items returns a unpaginated hashmap of all todo items
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
            Ok(result) => Ok(result),
            Err(e) => Err(anyhow!(
                "could not complete add item {}; failed transaction: {}",
                item.id.clone(),
                e.to_string()
            )),
        }
    }

    pub fn update_item(&self, id: &str, item: Item) -> Result<()> {
        let transaction_result = self.db.transaction(|tx_db| {
            let raw_old_item = tx_db.get(id)?.unwrap();
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

            let (added_children, removed_children) =
                find_list_updates(&old_list_children, &new_list_children);

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
            item.added = old_item.added.clone();
            item.modified = SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs();

            let raw_item: Vec<u8> = match bincode::serialize(&item) {
                Ok(raw_item) => raw_item,
                Err(e) => return sled::transaction::abort(anyhow!(e)),
            };

            tx_db.insert(id, raw_item)?;

            Ok(())
        });

        match transaction_result {
            Ok(result) => Ok(result),
            Err(e) => Err(anyhow!(
                "could not complete add item {}; failed transaction: {}",
                item.id.clone(),
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
    parent_id: &str,
    child_id: &str,
) -> Result<()> {
    let raw_child = db.get(child_id)?.unwrap();
    let mut child: Item = bincode::deserialize(&raw_child.to_vec())?;

    child.parent = Some(parent_id.clone().to_string());
    child.modified = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs();

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
    let raw_parent = db.get(&parent_id)?.unwrap();
    let mut parent: Item = bincode::deserialize(&raw_parent.to_vec())?;

    match parent.children.clone() {
        Some(mut children) => {
            children.push(child_id.clone().to_string());
            parent.children = Some(children);
        }
        None => {
            parent.children = Some(vec![child_id.clone().to_string()]);
        }
    }

    parent.modified = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs();

    let raw_parent: Vec<u8> = bincode::serialize(&parent)
        .with_context(|| format!("Failed to encode item {}", parent_id))?;

    let parent_id: &str = &parent_id.clone();
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

    parent.modified = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs();

    let raw_parent: Vec<u8> = bincode::serialize(&parent)
        .with_context(|| format!("Failed to encode item {}", parent_id))?;

    let parent_id: &str = &parent_id.clone();
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

    child.parent = None;
    child.modified = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs();

    let raw_child: Vec<u8> = bincode::serialize(&child)
        .with_context(|| format!("Failed to encode item {}", child_id))?;

    let child_id: &str = &child_id.clone();
    db.insert(child_id, raw_child)
        .with_context(|| format!("Failed to update item {}", child_id))?;

    Ok(())
}

fn find_list_difference(list1: &Vec<String>, list2: &Vec<String>) -> Vec<String> {
    let list1: HashSet<_> = list1.iter().cloned().collect();
    let list2: HashSet<_> = list2.iter().cloned().collect();

    let diff = list1.difference(&list2).cloned().collect();
    diff
}

fn find_list_updates(old_list: &Vec<String>, new_list: &Vec<String>) -> (Vec<String>, Vec<String>) {
    let removals = find_list_difference(old_list, new_list);
    let additions = find_list_difference(new_list, old_list);
    return (removals, additions);
}

#[cfg(test)]
mod tests {
    use super::*;
    use lazy_static::lazy_static;
    use serial_test::serial;

    lazy_static! {
        static ref DB: Storage = setup_db();
    }

    fn setup_db() -> Storage {
        let db = sled::Config::new().temporary(true).open().unwrap();
        Storage { db }
    }

    macro_rules! vec_of_strings {
        ($($x:expr),*) => (vec![$($x.to_string()),*]);
    }

    #[test]
    fn test_find_list_difference() {
        let test_list1 = vec_of_strings!("1", "2", "3", "4");
        let test_list2 = vec_of_strings!("1", "3");

        let mut returned_list = find_list_difference(&test_list1, &test_list2);
        let mut expected_list = vec_of_strings!("2", "4");

        assert_eq!(returned_list.sort(), expected_list.sort())
    }

    #[test]
    fn test_find_list_updates() {
        let current_list = vec_of_strings!("1", "2", "3", "4");
        let update_list = vec_of_strings!("1", "3", "5", "6");

        let (mut list_removals, mut list_additions) =
            find_list_updates(&current_list, &update_list);

        let mut expected_additions = vec_of_strings!("5", "6");
        let mut expected_removals = vec_of_strings!("2", "4");

        assert_eq!(expected_additions.sort(), list_additions.sort());
        assert_eq!(expected_removals.sort(), list_removals.sort());
    }

    #[test]
    #[serial]
    fn add_single_item() {
        let mut test_item: Item = Default::default();
        test_item.id = "1".to_string();
        test_item.title = "test title 1".to_string();
        test_item.description = Some("test description 1".to_string());

        let expected_item = test_item.clone();

        DB.add_item(test_item).unwrap();
        let returned_item = DB.get_item("1").unwrap().unwrap();

        assert_eq!(expected_item, returned_item);
        DB.clear();
    }

    #[test]
    #[serial]
    fn get_all_items() {
        let mut test_item: Item = Default::default();
        test_item.id = "1".to_string();
        test_item.title = "test title 1".to_string();
        test_item.description = Some("test description 1".to_string());

        DB.add_item(test_item).unwrap();

        let expected_item = DB.get_item("1").unwrap().unwrap();

        let items = DB.get_all_items().unwrap();
        let mut expected_map = std::collections::HashMap::new();
        expected_map.insert(expected_item.id.clone(), expected_item);
        let expected_items = Items {
            items: expected_map,
        };

        assert_eq!(items, expected_items);
        DB.clear();
    }

    #[test]
    #[serial]
    // check that we can add a child item with a parent string and logic to update the parent works
    fn add_child_item_with_parent() {
        let mut test_item_1: Item = Default::default();
        test_item_1.id = "1".to_string();
        test_item_1.title = "test title 1".to_string();
        test_item_1.description = Some("test description 1".to_string());

        DB.add_item(test_item_1).unwrap();

        let mut test_item_2: Item = Default::default();
        test_item_2.id = "2".to_string();
        test_item_2.title = "test title 2".to_string();
        test_item_2.description = Some("test description 2".to_string());
        test_item_2.parent = Some("1".to_string());

        let expected_item = test_item_2.clone();

        DB.add_item(test_item_2).unwrap();
        let returned_item = DB.get_item("2").unwrap().unwrap();

        assert_eq!(expected_item, returned_item);

        let parent_item = DB.get_item("1").unwrap().unwrap();
        let expected_children = vec!["2"];

        assert_eq!(parent_item.children.unwrap(), expected_children);
        DB.clear();
    }

    #[test]
    #[serial]
    fn update_simple_item() {
        let mut test_item_1: Item = Default::default();
        test_item_1.id = "1".to_string();
        test_item_1.title = "test title 1".to_string();
        test_item_1.description = Some("test description 1".to_string());

        let mut updated_item = test_item_1.clone();

        DB.add_item(test_item_1).unwrap();

        updated_item.description = Some("test description 2".to_string());

        DB.update_item(&updated_item.id.clone(), updated_item)
            .unwrap();

        let returned_item = DB.get_item(&"1".to_string()).unwrap().unwrap();
        assert_eq!(
            Some("test description 2".to_string()),
            returned_item.description
        );
    }

    #[test]
    #[serial]
    // test that we correctly add the child_id to the parent when we update a child with a new parent id
    fn update_item_with_new_parent() {
        let mut parent_item: Item = Default::default();
        parent_item.id = "parent".to_string();
        parent_item.title = "test title 1".to_string();
        parent_item.description = Some("test description 1".to_string());

        DB.add_item(parent_item).unwrap();

        let mut child_item: Item = Default::default();
        child_item.id = "child".to_string();
        child_item.title = "test title 2".to_string();
        child_item.description = Some("test description 2".to_string());

        DB.add_item(child_item).unwrap();

        let mut updated_child_item = DB.get_item(&"child".to_string()).unwrap().unwrap();
        updated_child_item.parent = Some("parent".to_string());

        DB.update_item(&updated_child_item.id.clone(), updated_child_item)
            .unwrap();

        let returned_parent = DB.get_item(&"parent".to_string()).unwrap().unwrap();
        let returned_parent_children = returned_parent.children.unwrap();

        assert!(returned_parent_children.contains(&"child".to_string()));
    }

    #[test]
    #[serial]
    fn update_item_with_removed_parent() {
        let mut parent_item: Item = Default::default();
        parent_item.id = "parent".to_string();
        parent_item.title = "test title 1".to_string();
        parent_item.description = Some("test description 1".to_string());
        parent_item.children = Some(vec_of_strings!("child", "hello"));

        DB.add_item(parent_item).unwrap();

        let mut child_item: Item = Default::default();
        child_item.id = "child".to_string();
        child_item.title = "test title 2".to_string();
        child_item.description = Some("test description 2".to_string());
        child_item.parent = Some("parent".to_string());

        DB.add_item(child_item).unwrap();

        let mut updated_child_item = DB.get_item(&"child".to_string()).unwrap().unwrap();
        updated_child_item.parent = None;

        DB.update_item(&updated_child_item.id.clone(), updated_child_item)
            .unwrap();

        let returned_parent = DB.get_item(&"parent".to_string()).unwrap().unwrap();
        let returned_parent_children = returned_parent.children.unwrap();

        assert!(!returned_parent_children.contains(&"child".to_string()));
    }

    fn update_item_with_removed_child() {}
    fn update_item_with_added_child() {}

    #[test]
    #[serial]
    fn delete_single_item() {
        let mut test_item: Item = Default::default();
        test_item.id = "1".to_string();
        test_item.title = "test title 1".to_string();
        test_item.description = Some("test description 1".to_string());

        DB.add_item(test_item).unwrap();
        let returned_item = DB.get_item("1").unwrap();
        assert!(returned_item.is_some());

        DB.delete_item(&"1".to_string()).unwrap();
        let returned_item = DB.get_item("1").unwrap();
        assert!(returned_item.is_none());

        DB.clear();
    }
}
