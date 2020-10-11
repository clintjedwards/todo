use super::config;
use super::storage;
use slog_scope::info;
use tide;

mod handlers;

pub async fn run_server(address: &str) -> Result<(), std::io::Error> {
    let config = config::get_config();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };

    let _ = storage::Storage::new(&config.database_path);
    info!("initialized database"; "path" => &config.database_path);

    let mut webserver = tide::new();
    webserver.at("/").get(|_| async { Ok("Hello, world!") });
    webserver.at("/:id").get(|_| async { Ok("Hello, world!") });
    webserver.at("/:id").post(|_| async { Ok("Hello, world!") });
    webserver.at("/:id").put(|_| async { Ok("Hello, world!") });
    webserver
        .at("/:id")
        .delete(|_| async { Ok("Hello, world!") });

    info!("started http server"; "address" => address);

    webserver.listen("127.0.0.1:8080").await?;
    Ok(())
}
