use super::config;
use super::models;
use super::storage;
use slog_scope::{error, info};
use tide;
use tide::prelude::*;
use tide::{Request, Response, StatusCode};

// API represents a REST API object
#[derive(Debug, Clone)]
pub struct API {
    db: storage::Storage,
    config: config::Config,
}

//TODO:(clintjedwards): Make it easier so that when you add a new field to the modal you don't
// have to update 'add' and 'update' functions. So that it leads to less bugs. We should be able
// to enumerate through and auto add them.

pub fn new() -> API {
    let config = config::get();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };

    let db = storage::new(&config.database_path);
    info!("initialized database"; "path" => &config.database_path);

    API { db, config }
}

impl API {
    pub async fn run_server(self, address: &str) -> Result<(), std::io::Error> {
        let mut webserver = tide::with_state(self);
        webserver.at("/").get(get_all_items_handler);
        webserver.at("/").post(add_item_handler);
        webserver.at("/:id").get(get_item_handler);
        webserver.at("/:id").put(update_item_handler);
        webserver.at("/:id").delete(delete_item_handler);

        info!("started http server"; "address" => address);

        webserver.listen(address).await?;
        Ok(())
    }
}

// get_all_items_handler retrieves an unpaginated list of all todo items.
async fn get_all_items_handler(req: Request<API>) -> tide::Result {
    let items = match req.state().db.get_all_items() {
        Ok(items) => items,
        Err(error) => {
            error!("could not get items"; "err" => format!("{}", error));
            let response = Response::builder(StatusCode::InternalServerError).build();
            return Ok(response);
        }
    };
    let response = Response::builder(StatusCode::Ok).body(json!(items)).build();
    Ok(response)
}

// add_item_handler stores a single todo item.
async fn add_item_handler(mut req: Request<API>) -> tide::Result {
    let add_item_request: models::AddItemRequest = req.body_json().await?;

    let mut new_item = models::new_item(req.state().config.id_length, &add_item_request.title);
    new_item.title = add_item_request.title;
    new_item.link = add_item_request.link;
    new_item.description = add_item_request.description;
    new_item.parent = add_item_request.parent;

    let committed_item = new_item.clone();
    match req.state().db.add_item(new_item) {
        Ok(_) => {}
        Err(e) => {
            let response = Response::builder(StatusCode::InternalServerError)
                .body(json!({ "failed to add item": format!("{}", e) }))
                .build();
            error!("could not add item"; "err" => format!("{}", e));
            return Ok(response);
        }
    }

    info!("added item"; "id" => &committed_item.id);
    let response = Response::builder(StatusCode::Created)
        .body(json!(committed_item))
        .build();
    Ok(response)
}

// update_item_handler determines which fields have been updated and updates the
// stored item.
async fn update_item_handler(mut req: Request<API>) -> tide::Result {
    let id: String = req.param("id")?;

    let update_item_request: models::UpdateItemRequest = req.body_json().await?;

    let updated_item = req.state().db.get_item(&id)?;
    let mut updated_item = match updated_item {
        Some(updated_item) => updated_item,
        None => {
            let response = Response::builder(StatusCode::NotFound).build();
            return Ok(response);
        }
    };

    //TODO(clintjedwards): find a solution for long if some let chains like this

    // Update only fields that have changed
    if let Some(title) = update_item_request.title {
        updated_item.title = title;
    }
    if let Some(completed) = update_item_request.completed {
        updated_item.completed = completed;
    }
    if let Some(link) = update_item_request.link {
        updated_item.link = Some(link);
    }
    if let Some(description) = update_item_request.description {
        updated_item.description = Some(description);
    }
    if let Some(parent) = update_item_request.parent {
        updated_item.parent = Some(parent);
    }
    if let Some(children) = update_item_request.children {
        updated_item.children = Some(children);
    }

    let committed_item = updated_item.clone();
    match req.state().db.update_item(updated_item) {
        Ok(_) => {}
        Err(e) => {
            let response = Response::builder(StatusCode::InternalServerError)
                .body(json!({ "err": format!("{}", e) }))
                .build();
            error!("could not update item"; "id" => &id, "error" => format!("{}", e));
            return Ok(response);
        }
    }

    info!("updated item"; "id" => &id);
    let response = Response::builder(StatusCode::Created)
        .body(json!(committed_item))
        .build();
    Ok(response)
}

// get_item_handler returns a single item by id.
async fn get_item_handler(req: Request<API>) -> tide::Result {
    let id: String = req.param("id")?;

    let item = match req.state().db.get_item(&id) {
        Ok(item) => item,
        Err(error) => {
            let response = Response::builder(StatusCode::InternalServerError)
                .body(json!({ "err": format!("{}", error) }))
                .build();
            error!("could not get item"; "id" => &id, "error" => format!("{}", error));
            return Ok(response);
        }
    };
    match item {
        Some(item) => {
            let response = Response::builder(StatusCode::Ok).body(json!(item)).build();
            return Ok(response);
        }
        None => {
            let response = Response::builder(StatusCode::NotFound).build();
            return Ok(response);
        }
    }
}

// delete_item_handler removes a single item by id.
// Deletion of a non-existant key will result in a successful call.
async fn delete_item_handler(req: Request<API>) -> tide::Result {
    let id: String = req.param("id")?;

    match req.state().db.delete_item(&id) {
        Ok(_) => {}
        Err(error) => {
            let response = Response::builder(StatusCode::InternalServerError).build();
            error!("could not remove item"; "id" => &id, "error" => format!("{}", error));
            return Ok(response);
        }
    };

    info!("deleted item"; "id" => &id);
    let response = Response::builder(StatusCode::Ok).build();
    Ok(response)
}
