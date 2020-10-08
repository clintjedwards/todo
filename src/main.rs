use clap::crate_version;
use clap::{App, AppSettings, Arg, SubCommand};
use slog::o;
use slog::Drain;
use slog_scope::info;
use std::error::Error;
use std::net::SocketAddr;
use warp::Filter;

mod config;
mod storage;

fn main() -> Result<(), Box<(dyn Error)>> {
    let _guard = init_logging();

    let subcommand_add = SubCommand::with_name("add")
        .about("Add an item to the todo list.")
        .arg(Arg::with_name("description").required(true).index(1));

    let subcommand_update = SubCommand::with_name("update")
        .about("Alter an already existing todo item.")
        .arg(Arg::with_name("id").required(true).index(1));

    let subcommand_remove = SubCommand::with_name("remove")
        .about("Remove an item from the todo list.")
        .arg(Arg::with_name("id").required(true).index(1));

    let subcommand_server = SubCommand::with_name("server")
        .about("Start Todo web service.")
        .arg(Arg::with_name("address").required(true).index(1));

    let app = App::new("Todo")
        .about("A simple todo list application")
        .version(crate_version!())
        .setting(AppSettings::SubcommandRequired)
        .subcommand(subcommand_add)
        .subcommand(subcommand_update)
        .subcommand(subcommand_remove)
        .subcommand(subcommand_server);

    let matches = app.get_matches();

    if let Some(sub_matcher) = matches.subcommand_matches("add") {
        let description = sub_matcher.value_of("description").unwrap();
    }

    if let Some(sub_matcher) = matches.subcommand_matches("updarte") {
        let id = sub_matcher.value_of("id").unwrap();
    }

    if let Some(sub_matcher) = matches.subcommand_matches("remove") {
        let id = sub_matcher.value_of("id").unwrap();
    }

    if let Some(sub_matcher) = matches.subcommand_matches("server") {
        let address = sub_matcher.value_of("address").unwrap();
        run_server(address)
    }

    return Ok(());
}

fn init_logging() -> slog_scope::GlobalLoggerGuard {
    let decorator = slog_term::PlainDecorator::new(std::io::stdout());
    let root_logger = slog_term::CompactFormat::new(decorator).build().fuse();
    let root_logger = slog_async::Async::new(root_logger).build().fuse();
    let log = slog::Logger::root(root_logger, o!());

    let guard = slog_scope::set_global_logger(log);

    guard
}

#[tokio::main]
async fn run_server(address: &str) {
    let config = config::get_config();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };

    let _ = storage::init_database(&config.database_path);
    info!("initialized database"; "path" => &config.database_path);

    // Match any request and return hello world!
    let routes = warp::any().map(|| "Hello, World!");
    let socket_address: SocketAddr = address.parse().expect("Unable to parse socket address");

    info!("started http server"; "address" => address);
    warp::serve(routes).run(socket_address).await;
}
