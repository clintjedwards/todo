use serde::Deserialize;

#[derive(Deserialize, Debug, Clone)]
pub struct Config {
    pub log_level: Option<String>,
    #[serde(default = "default_database_path")]
    pub database_path: String,
    #[serde(default = "default_id_length")]
    pub id_length: usize, // The length of autogenerated ids.
}

fn default_id_length() -> usize {
    5
}

fn default_database_path() -> String {
    String::from("/tmp/test.db")
}

pub fn get_config() -> Result<Config, envy::Error> {
    envy::from_env::<Config>()
}
