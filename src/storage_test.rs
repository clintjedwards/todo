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

    let (mut list_removals, mut list_additions) = find_list_updates(&current_list, &update_list);

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
    let mut expected_map = std::collections::BTreeMap::new();
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

    DB.update_item(updated_item).unwrap();

    let returned_item = DB.get_item(&"1".to_string()).unwrap().unwrap();
    assert_eq!(
        Some("test description 2".to_string()),
        returned_item.description
    );
    DB.clear();
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

    DB.update_item(updated_child_item).unwrap();

    let returned_parent = DB.get_item(&"parent".to_string()).unwrap().unwrap();
    let returned_parent_children = returned_parent.children.unwrap();

    assert!(returned_parent_children.contains(&"child".to_string()));
    DB.clear();
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

    DB.update_item(updated_child_item).unwrap();

    let returned_parent = DB.get_item(&"parent".to_string()).unwrap().unwrap();
    let returned_parent_children = returned_parent.children.unwrap();

    assert!(!returned_parent_children.contains(&"child".to_string()));
    DB.clear();
}

#[test]
#[serial]
// make sure that when you remove a child from an item that child's parent
// field is none.
fn update_item_with_removed_child() {
    let mut parent_item: Item = Default::default();
    parent_item.id = "parent".to_string();
    parent_item.title = "test title 1".to_string();
    parent_item.description = Some("test description 1".to_string());
    parent_item.children = Some(vec_of_strings!("child", "hello"));

    DB.add_item(parent_item).unwrap();

    let mut child_item: Item = Default::default();
    child_item.id = "child".to_string();
    child_item.title = "test title 1".to_string();
    child_item.description = Some("test description 1".to_string());

    DB.add_item(child_item).unwrap();

    let mut hello_item: Item = Default::default();
    hello_item.id = "hello".to_string();
    hello_item.title = "test title 1".to_string();
    hello_item.description = Some("test description 1".to_string());

    DB.add_item(hello_item).unwrap();

    let mut updated_parent_item = DB.get_item(&"parent".to_string()).unwrap().unwrap();
    updated_parent_item.children = Some(vec_of_strings!("hello"));

    DB.update_item(updated_parent_item).unwrap();

    let child = DB.get_item(&"child".to_string()).unwrap().unwrap();
    assert!(child.parent.is_none());

    DB.clear();
}

#[test]
#[serial]
// make sure that when you add a child that child get updated with its new parent
// any current parent should be updated so it doesn't own the child anymore.
fn update_item_with_added_child() {
    let mut parent_item: Item = Default::default();
    parent_item.id = "parent".to_string();
    parent_item.title = "test title 1".to_string();
    parent_item.description = Some("test description 1".to_string());
    parent_item.children = Some(vec_of_strings!("child"));

    DB.add_item(parent_item).unwrap();

    let mut child_item: Item = Default::default();
    child_item.id = "child".to_string();
    child_item.title = "test title 1".to_string();
    child_item.description = Some("test description 1".to_string());

    DB.add_item(child_item).unwrap();

    let mut hello_item: Item = Default::default();
    hello_item.id = "hello".to_string();
    hello_item.title = "test title 1".to_string();
    hello_item.description = Some("test description 1".to_string());

    DB.add_item(hello_item).unwrap();

    let mut updated_parent_item = DB.get_item(&"parent".to_string()).unwrap().unwrap();
    updated_parent_item.children = Some(vec_of_strings!("child", "hello"));

    DB.update_item(updated_parent_item).unwrap();

    let child = DB.get_item(&"hello".to_string()).unwrap().unwrap();
    assert!(child.parent.is_some());
    assert_eq!(&child.parent.unwrap(), &"parent".to_string());

    DB.clear();
}

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
