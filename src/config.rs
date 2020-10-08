use serde::Deserialize;

#[derive(Deserialize, Debug)]
pub struct Config {
    pub log_level: Option<String>,
    pub database_path: String,
}

pub fn get_config() -> Result<Config, envy::Error> {
    envy::from_env::<Config>()
}
