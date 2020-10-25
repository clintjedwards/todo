use super::config;
use super::models;
use super::storage;
use slog_scope::info;
use tide;
use tide::prelude::*;
use tide::{Request, Response, StatusCode};

#[derive(Debug, Clone)]
pub struct API {
    db: storage::Storage,
    config: config::Config,
}

pub fn new() -> API {
    let config = config::get_config();
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

        webserver.listen("127.0.0.1:8080").await?;
        Ok(())
    }
}

async fn get_all_items_handler(req: Request<API>) -> tide::Result {
    let items = req.state().db.get_all_items()?;
    let response = Response::builder(StatusCode::Ok).body(json!(items)).build();
    Ok(response)
}

async fn add_item_handler(mut req: Request<API>) -> tide::Result {
    #[derive(Deserialize, Clone)]
    struct AddItemRequest {
        parent: Option<String>,
        title: String,
        description: Option<String>,
    };
    let add_item_request: AddItemRequest = req.body_json().await?;

    let mut new_item = models::new_item(req.state().config.id_length, &add_item_request.title);
    new_item.title = add_item_request.title;
    new_item.description = add_item_request.description;
    new_item.parent = add_item_request.parent;

    let committed_item = new_item.clone();
    req.state().db.add_item(new_item)?;

    let response = Response::builder(StatusCode::Created)
        .body(json!(committed_item))
        .build();
    Ok(response)
}

async fn update_item_handler(mut req: Request<API>) -> tide::Result {
    let id: String = req.param("id")?;

    #[derive(Deserialize, Clone)]
    struct UpdateItemRequest {
        parent: Option<String>,
        children: Option<Vec<String>>,
        title: String,
        description: Option<String>,
    };
    let update_item_request: UpdateItemRequest = req.body_json().await?;

    let updated_item = req.state().db.get_item(&id)?;
    let mut updated_item = match updated_item {
        Some(updated_item) => updated_item,
        None => {
            let response = Response::builder(StatusCode::NotFound).build();
            return Ok(response);
        }
    };
    updated_item.title = update_item_request.title;
    updated_item.description = update_item_request.description;
    updated_item.parent = update_item_request.parent;
    updated_item.children = update_item_request.children;

    let committed_item = updated_item.clone();
    req.state().db.update_item(updated_item)?;

    let response = Response::builder(StatusCode::Created)
        .body(json!(committed_item))
        .build();
    Ok(response)
}

async fn get_item_handler(req: Request<API>) -> tide::Result {
    let id: String = req.param("id")?;

    let item = req.state().db.get_item(&id)?;
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

async fn delete_item_handler(req: Request<API>) -> tide::Result {
    let id: String = req.param("id")?;

    req.state().db.delete_item(&id)?;

    let response = Response::builder(StatusCode::Ok).build();
    Ok(response)
}
