use anyhow::{Context, Result};
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

    //TODO(clintjedwards):
    pub fn get_all_items(&self, id: &str) -> Result<Option<String>> {
        match self
            .db
            .get(id)
            .with_context(|| format!("Failed to get item {}", id))?
        {
            None => Ok(None),
            Some(buf) => Ok(Some(String::from_utf8(buf.to_vec())?)),
        }
    }

    //TODO(clintjedwards): remove the option from this
    pub fn get_item(&self, id: &str) -> Result<Option<String>> {
        match self
            .db
            .get(id)
            .with_context(|| format!("Failed to get item {}", id))?
        {
            None => Ok(None),
            Some(buf) => Ok(Some(String::from_utf8(buf.to_vec())?)),
        }
    }

    pub fn add_item(&self, id: &str, description: &str) -> Result<()> {
        self.db
            .insert(id, description)
            .with_context(|| format!("Failed to add item {}", id))?;

        Ok(())
    }

    pub fn remove_item(&self, id: &str) -> Result<()> {
        self.db
            .remove(id)
            .with_context(|| format!("Failed to remove item {}", id))?;
        Ok(())
    }

    pub fn update_item(&self, id: &str) -> Result<()> {
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_add_item() {
        let db = Storage::new("/tmp/test.db");
        db.add_item("1", "some title").unwrap();
        let item = db.get_item("1").unwrap();
        dbg!(item.unwrap());
    }
}
