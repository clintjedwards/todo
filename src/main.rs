use anyhow::Result;
use models::{Item, UpdateItemRequest};
use slog::o;
use slog::Drain;
use structopt::StructOpt;

mod api;
mod cli;
mod config;
mod models;
mod storage;

#[derive(Debug, StructOpt)]
#[structopt(name = "todo")]
enum Opt {
    /// Add an item to the todo list.
    Add {
        title: String,

        /// Give further color about what this todo item might be about.
        #[structopt(short, long)]
        description: Option<String>,

        /// Which todo item (by id) should this be a child of.
        #[structopt(short, long, name = "ID")]
        parent: Option<String>,

        /// URL to 3rd party application where todo might be tracked.
        #[structopt(short, long)]
        link: Option<String>,

        // Use a text editor
        #[structopt(short, long)]
        interactive: bool,
    },

    /// Alter an already existing todo item.
    Update {
        id: String,

        /// The title for the todo item.
        #[structopt(short, long)]
        title: Option<String>,

        /// Give further color about what this todo item might be about.
        #[structopt(short, long)]
        description: Option<String>,

        /// Which todo item (by id) should this be a child of.
        #[structopt(short, long, name = "ID")]
        parent: Option<String>,

        /// URL to 3rd party application where todo might be tracked.
        #[structopt(short, long)]
        link: Option<String>,

        /// Which todo items (by id) should this be a parent of; comma delimited.
        #[structopt(short, long)]
        children: Option<Vec<String>>,

        /// Mark a todo item as done.
        #[structopt(long = "complete")]
        completed: Option<bool>,

        // Use a text editor
        #[structopt(short, long)]
        interactive: bool,
    },

    /// Get an item by ID.
    Get { id: String },

    /// Remove an item from the todo list.
    Remove { id: String },

    /// Toggle the completeness of a task.
    Complete { id: String },

    /// List all outstanding todo items.
    List {
        /// Show items which have been completed already.
        #[structopt(short, long)]
        show_completed: bool,
    },

    /// Start the todo web service.
    Server {
        #[structopt(default_value = "localhost:8080")]
        address: String,
    },
}

#[async_std::main]
async fn main() -> Result<()> {
    let _guard = init_logging();
    let cli = cli::new();

    match Opt::from_args() {
        Opt::Add {
            title,
            description,
            parent,
            link,
            interactive,
        } => cli.add_todo(
            Item {
                title,
                description,
                parent,
                link,
                ..Default::default()
            },
            interactive,
        ),
        Opt::Complete { id } => cli.complete_todo(&id),
        Opt::Get { id } => cli.get_todo(&id),
        Opt::List { show_completed } => cli.list_todos(show_completed),
        Opt::Remove { id } => cli.remove_todo(&id),
        Opt::Server { address } => {
            let api = api::new();
            return Ok(api.run_server(&address).await?);
        }
        Opt::Update {
            id,
            title,
            description,
            parent,
            link,
            children,
            completed,
            interactive,
        } => cli.update_todo(
            &id,
            UpdateItemRequest {
                title,
                description,
                parent,
                link,
                children,
                completed,
            },
            interactive,
        ),
    }
}

fn init_logging() -> slog_scope::GlobalLoggerGuard {
    let decorator = slog_term::PlainDecorator::new(std::io::stdout());
    let root_logger = slog_term::CompactFormat::new(decorator).build().fuse();
    let root_logger = slog_async::Async::new(root_logger).build().fuse();
    let log = slog::Logger::root(root_logger, o!());

    let guard = slog_scope::set_global_logger(log);

    guard
}
