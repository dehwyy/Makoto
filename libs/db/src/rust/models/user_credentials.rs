//! `SeaORM` Entity. Generated by sea-orm-codegen 0.12.6

use sea_orm::entity::prelude::*;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel, Eq)]
#[sea_orm(table_name = "user_credentials")]
pub struct Model {
    #[sea_orm(primary_key, auto_increment = false)]
    pub id: Uuid,
    #[sea_orm(column_type = "Text", unique)]
    pub email: String,
    #[sea_orm(column_type = "Text", nullable)]
    pub password: Option<String>,
    pub created_at: Option<DateTimeWithTimeZone>,
    #[sea_orm(column_type = "Text", unique)]
    pub username: String,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::user_oauths::Entity")]
    UserOauths,
    #[sea_orm(has_many = "super::user_tokens::Entity")]
    UserTokens,
}

impl Related<super::user_oauths::Entity> for Entity {
    fn to() -> RelationDef {
        Relation::UserOauths.def()
    }
}

impl Related<super::user_tokens::Entity> for Entity {
    fn to() -> RelationDef {
        Relation::UserTokens.def()
    }
}

impl ActiveModelBehavior for ActiveModel {}
