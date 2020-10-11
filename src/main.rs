use clap::crate_version;
use clap::{App, AppSettings, Arg, SubCommand};
use slog::o;
use slog::Drain;
use std::error::Error;

mod api;
mod config;
mod models;
mod storage;

#[async_std::main]
async fn main() -> Result<(), Box<(dyn Error)>> {
    let _guard = init_logging();

    let subcommand_add = SubCommand::with_name("add")
        .about("Add an item to the todo list.")
        .arg(Arg::with_name("title").required(true).index(1));

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

    // It's okay to unwrap value_of calls since they are required and should not cause panics
    if let Some(sub_matcher) = matches.subcommand_matches("add") {
        let title = sub_matcher.value_of("title").unwrap();
    }

    if let Some(sub_matcher) = matches.subcommand_matches("update") {
        let id = sub_matcher.value_of("id").unwrap();
    }

    if let Some(sub_matcher) = matches.subcommand_matches("remove") {
        let id = sub_matcher.value_of("id").unwrap();
    }

    if let Some(sub_matcher) = matches.subcommand_matches("server") {
        let address = sub_matcher.value_of("address").unwrap();
        api::run_server(address).await?;
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
