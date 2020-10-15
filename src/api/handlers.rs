use super::super::models::Item;
use super::config;
use super::storage;
use tide::prelude::*;
use tide::{Request, Response, Result, StatusCode};

#[derive(Debug, Clone)]
pub struct API {
    pub db: storage::Storage,
    pub config: config::Config,
}

pub async fn get_all_items_handler(req: Request<API>) -> Result {
    let items = req.state().db.get_all_items()?;
    let response = Response::builder(StatusCode::Ok).body(json!(items)).build();
    Ok(response)
}

pub async fn add_item_handler(mut req: Request<API>) -> Result {
    #[derive(Deserialize, Clone)]
    struct AddItemRequest {
        parent: Option<String>,
        title: String,
        description: Option<String>,
    };
    let add_item_request: AddItemRequest = req.body_json().await?;

    let mut new_item = Item::new(req.state().config.id_length, &add_item_request.title);
    new_item.title = add_item_request.title;
    new_item.description = add_item_request.description;
    new_item.parent = add_item_request.parent;

    req.state().db.add_item(&new_item.id, &new_item)?;

    let response = Response::builder(StatusCode::Created)
        .body(json!(new_item))
        .build();
    Ok(response)
}

//TODO(clintjedwards):Error handle 404s and make sure erorrs are properly
//handled
pub async fn get_item_handler(req: Request<API>) -> Result {
    let id: String = req.param("id")?;

    let item = req.state().db.get_item(&id)?;

    let response = Response::builder(StatusCode::Ok).body(json!(item)).build();
    Ok(response)
}

pub async fn delete_item_handler(req: Request<API>) -> Result {
    let id: String = req.param("id")?;

    req.state().db.delete_item(&id)?;

    let response = Response::builder(StatusCode::Ok).build();
    Ok(response)
}
