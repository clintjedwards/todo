use super::models::{Item, Items};
use anyhow::{Context, Result};
use sled;
use std::time::SystemTime;

#[derive(Debug, Clone)]
pub struct Storage {
    db: sled::Db,
}

pub fn new(path: &str) -> Storage {
    Storage {
        db: sled::open(path).expect("open"),
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
                Err(e) => return Err(anyhow::anyhow!(e)),
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
            //throw an error
        }

        // update parent if exists
        //TODO(clintjedwards): this needs to be in a transaction
        self.add_child_to_parent(item.parent.clone(), &item.id.clone())?;

        let id = item.id.clone();
        let raw_item: Vec<u8> = bincode::serialize(&item)
            .with_context(|| format!("Failed to encode item object {}", id))?;

        self.db
            .insert(item.id, raw_item)
            .with_context(|| format!("Failed to add item {}", id))?;

        Ok(())
    }

    pub fn update_item(&self, id: &str, mut item: Item) -> Result<()> {
        //TODO(clintjedwards): we need to handle the case where the user alters the children
        //TODO(clintjedwards): we need to handle the case where the user has changed the parent

        let old_item = self
            .get_item(id)
            .with_context(|| format!("Failed to get old item {}", id))?
            .unwrap();

        item.id = old_item.id.clone();
        item.added = old_item.added.clone();
        item.modified = SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)
            .unwrap()
            .as_secs();

        let raw_item: Vec<u8> =
            bincode::serialize(&item).with_context(|| format!("Failed to encode item {}", id))?;

        self.db
            .insert(id, raw_item)
            .with_context(|| format!("Failed to update item {}", id))?;

        Ok(())
    }

    // delete_item deletes a single item
    pub fn delete_item(&self, id: &str) -> Result<()> {
        self.db
            .remove(id)
            .with_context(|| format!("Failed to remove item {}", id))?;
        Ok(())
    }

    // Adds a parent to an already existing child
    fn add_parent_to_child(&self, child_id: &str, parent_id: &str) -> Result<()> {
        let mut child = self.get_item(child_id)?.unwrap();
        child.parent = Some(parent_id.clone().to_string());
        self.update_item(child_id, child)?;
        Ok(())
    }

    // Adds a new child to the parent item, if that child does not already exist.
    fn add_child_to_parent(&self, parent_id: Option<String>, child_id: &str) -> Result<()> {
        if parent_id.is_none() {
            return Ok(());
        }

        let parent_id = parent_id.unwrap();
        let mut parent = self.get_item(&parent_id)?.unwrap();
        match parent.children.take() {
            Some(mut children) => {
                children.push(child_id.clone().to_string());
            }
            None => {
                parent.children = Some(vec![child_id.clone().to_string()]);
            }
        }
        self.update_item(&parent_id, parent)?;
        Ok(())
    }

    // this function is only used in testing
    #[allow(dead_code)]
    fn clear(&self) {
        let _ = self.db.clear();
    }
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
        let db = sled::open("/tmp/test.db").expect("could not open db");
        db.clear().unwrap();

        Storage { db }
    }

    #[test]
    #[serial]
    fn test_add_single_item() {
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
    fn test_get_all_items() {
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
    fn test_add_child_item() {
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
}
