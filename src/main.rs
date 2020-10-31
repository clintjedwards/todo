use anyhow::{anyhow, Result};
use clap::crate_version;
use clap::{App, AppSettings, Arg, SubCommand};
use slog::o;
use slog::Drain;

mod api;
mod cli;
mod config;
mod models;
mod storage;

//TODO(clintjedwards): Do proper error checking everywhere

#[async_std::main]
async fn main() -> Result<()> {
    let _guard = init_logging();
    let cli = cli::new();

    let subcommand_add = SubCommand::with_name("add")
        .about("Add an item to the todo list.")
        .arg(Arg::with_name("title").required(true).index(1))
        .arg(
            Arg::with_name("description")
                .short("d")
                .long("description")
                .help("Give further color about what this todo item might be about")
                .takes_value(true),
        )
        .arg(
            Arg::with_name("parent")
                .short("p")
                .long("parent")
                .help("Which todo item (by id) should this be a child of")
                .takes_value(true)
                .value_name("id"),
        )
        .arg(
            Arg::with_name("link")
                .short("l")
                .long("link")
                .help("URL to 3rd party application where todo might be tracked")
                .takes_value(true),
        );

    let subcommand_update = SubCommand::with_name("update")
        .about("Alter an already existing todo item")
        .arg(Arg::with_name("id").required(true).index(1))
        .arg(
            Arg::with_name("title")
                .short("t")
                .long("title")
                .help("The title for the todo item")
                .takes_value(true),
        )
        .arg(
            Arg::with_name("description")
                .short("d")
                .long("description")
                .help("Give further color about what this todo item might be about")
                .takes_value(true),
        )
        .arg(
            Arg::with_name("parent")
                .short("p")
                .long("parent")
                .help("Which todo item (by id) should this be a child of")
                .takes_value(true)
                .value_name("id"),
        )
        .arg(
            Arg::with_name("children")
                .short("c")
                .long("children")
                .help("Which todo items (by id) should this be a parent of; Comma delimited")
                .takes_value(true)
                .value_name("comma delimited ids"),
        )
        .arg(
            Arg::with_name("link")
                .short("l")
                .long("link")
                .help("URL to 3rd party application link or similar")
                .takes_value(true),
        )
        .arg(
            Arg::with_name("completed")
                .long("completed")
                .help("Mark a todo item as done")
                .takes_value(true),
        );

    let subcommand_get = SubCommand::with_name("get")
        .about("Get an item by id")
        .arg(Arg::with_name("id").required(true).index(1));

    let subcommand_remove = SubCommand::with_name("remove")
        .about("Remove an item from the todo list")
        .arg(Arg::with_name("id").required(true).index(1));

    let subcommand_complete = SubCommand::with_name("complete")
        .about("Complete or uncomplete a task")
        .arg(Arg::with_name("id").required(true).index(1));

    let subcommand_list = SubCommand::with_name("list")
        .about("List all outstanding todo items")
        .arg(
            Arg::with_name("show_completed")
                .short("s")
                .long("show_completed")
                .help("Show items which have been completed already"),
        );

    let subcommand_server = SubCommand::with_name("server")
        .about("Start Todo web service")
        .arg(Arg::with_name("address").required(true).index(1));

    let app = App::new("Todo")
        .about(
            "A simple todo application\n\n
Use `NO_COLOR=true` to turn off terminal coloring.",
        )
        .version(crate_version!())
        .setting(AppSettings::SubcommandRequired)
        .subcommand(subcommand_add)
        .subcommand(subcommand_update)
        .subcommand(subcommand_remove)
        .subcommand(subcommand_list)
        .subcommand(subcommand_get)
        .subcommand(subcommand_complete)
        .subcommand(subcommand_server);

    let matches = app.get_matches();

    if let Some(sub_matcher) = matches.subcommand_matches("add") {
        let mut new_item: models::Item = Default::default();

        // It's okay to unwrap required value_of calls since they cannot be none
        new_item.title = sub_matcher.value_of("title").unwrap().to_string();
        if let Some(parent) = sub_matcher.value_of("parent") {
            new_item.parent = Some(parent.to_string());
        }
        if let Some(description) = sub_matcher.value_of("description") {
            new_item.description = Some(description.to_string());
        }
        if let Some(link) = sub_matcher.value_of("link") {
            new_item.link = Some(link.to_string());
        }

        cli.add_todo(new_item)?;
    }

    if let Some(sub_matcher) = matches.subcommand_matches("update") {
        let id = sub_matcher.value_of("id").unwrap();
        let title = sub_matcher.value_of("title");
        let description = sub_matcher.value_of("description");
        let parent = sub_matcher.value_of("parent");
        let link = sub_matcher.value_of("link");
        let completed_option = sub_matcher.value_of("completed");
        let completed = match completed_option {
            Some(c) => {
                let result = c.parse::<bool>();
                let result = match result {
                    Ok(boolean) => boolean,
                    Err(e) => {
                        eprintln!("expected boolean; found {}", &c);
                        return Err(anyhow!(e));
                    }
                };
                Some(result)
            }
            None => None,
        };

        let children_option = sub_matcher.values_of("children");
        let children = match children_option {
            Some(children_iter) => Some(children_iter.map(str::to_string).collect()),
            None => None,
        };

        let updated_item = models::UpdateItemRequest {
            title: title.map(str::to_string),
            description: description.map(str::to_string),
            parent: parent.map(str::to_string),
            link: link.map(str::to_string),
            completed,
            children,
        };

        cli.update_todo(id, updated_item)?;
    }

    if let Some(sub_matcher) = matches.subcommand_matches("remove") {
        let id = sub_matcher.value_of("id").unwrap();
        cli.remove_todo(id)?;
    }

    if let Some(sub_matcher) = matches.subcommand_matches("list") {
        let show_completed = sub_matcher.is_present("show_completed");
        cli.list_todos(show_completed)?;
    }

    if let Some(sub_matcher) = matches.subcommand_matches("get") {
        let id = sub_matcher.value_of("id").unwrap();
        cli.get_todo(id)?;
    }

    if let Some(sub_matcher) = matches.subcommand_matches("complete") {
        let id = sub_matcher.value_of("id").unwrap();
        cli.complete_todo(id)?;
    }

    if let Some(sub_matcher) = matches.subcommand_matches("server") {
        let address = sub_matcher.value_of("address").unwrap();
        let api = api::new();
        api.run_server(address).await?;
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
