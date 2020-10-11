use super::models::{Item, Items};
use anyhow::{Context, Result};
use serial_test::serial;
use sled;

pub struct Storage {
    db: sled::Db,
}

impl Storage {
    pub fn new(path: &str) -> Storage {
        Storage {
            db: sled::open(path).expect("open"),
        }
    }

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
    pub fn add_item(&self, id: &str, item: Item) -> Result<()> {
        let raw_item: Vec<u8> = bincode::serialize(&item)
            .with_context(|| format!("Failed to encode item object {}", id))?;

        self.db
            .insert(id, raw_item)
            .with_context(|| format!("Failed to add item {}", id))?;

        Ok(())
    }

    // remove_item deletes a single item
    pub fn remove_item(&self, id: &str) -> Result<()> {
        self.db
            .remove(id)
            .with_context(|| format!("Failed to remove item {}", id))?;
        Ok(())
    }

    // update_item TODO(clintjedwards):
    pub fn update_item(&self, id: &str) -> Result<()> {
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    #[serial]
    fn test_get_all_items() {
        let db = Storage::new("/tmp/test.db");
        let mut test_item: Item = Default::default();
        let id = String::from("1");
        test_item.id = id.clone();
        test_item.title = String::from("test title");
        test_item.description = Some(String::from("test description"));

        let expected_item = test_item.clone();

        db.add_item("1", test_item).unwrap();
        let items = db.get_all_items().unwrap();
        let mut expected_map = std::collections::HashMap::new();
        expected_map.insert(id, expected_item);
        let expected_items = Items {
            items: expected_map,
        };

        assert_eq!(items, expected_items)
    }

    #[test]
    #[serial]
    fn test_get_item() {
        let db = Storage::new("/tmp/test.db");

        let mut expected_item: Item = Default::default();
        expected_item.id = String::from("1");
        expected_item.title = String::from("test title");
        expected_item.description = Some(String::from("test description"));

        let item = db.get_item("1").unwrap().unwrap();
        assert_eq!(item, expected_item)
    }

    //     #[test]
    //     fn test_add_item() {
    //         let db = Storage::new("/tmp/test.db");
    //         db.add_item("1", "some title").unwrap();
    //         let item = db.get_item("1").unwrap();
    //         dbg!(item.unwrap());
    //     }
}
