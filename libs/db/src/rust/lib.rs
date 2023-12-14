use sea_orm::{DatabaseConnection,  Database, ConnectOptions};

pub async fn new() -> Result<DatabaseConnection, ()> {
    let db_url = makoto_config::db::Database::new().database_url.expect("cannot get database_url from env!");

    let mut connection_options = ConnectOptions::new(db_url);
    connection_options.max_connections(100).sqlx_logging(false);

    let db = Database::connect(connection_options).await.map_err(|err| {
        eprintln!("Failed to connect to database: {err}");
    })?;

    Ok(db)
}
