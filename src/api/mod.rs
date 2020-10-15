use super::config;
use super::storage;
use slog_scope::info;
use tide;

mod handlers;

//TODO(clintjedwards): can run server just be the constructor method for API?
pub async fn run_server(address: &str) -> Result<(), std::io::Error> {
    let config = config::get_config();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };

    let db = storage::Storage::new(&config.database_path);
    info!("initialized database"; "path" => &config.database_path);

    let test = handlers::API { db, config };

    let mut webserver = tide::with_state(test);
    webserver.at("/").get(handlers::get_all_items_handler);
    webserver.at("/").post(handlers::add_item_handler);
    webserver.at("/:id").get(handlers::get_item_handler);
    webserver.at("/:id").put(|_| async { Ok("Hello, world!") });
    webserver.at("/:id").delete(handlers::delete_item_handler);

    info!("started http server"; "address" => address);

    webserver.listen("127.0.0.1:8080").await?;
    Ok(())
}
